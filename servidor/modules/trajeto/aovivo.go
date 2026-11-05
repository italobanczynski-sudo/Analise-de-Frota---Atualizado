package trajeto

import (
	"context"
	"errors"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"analisefrota/adapters/gobraxapi"
	"analisefrota/common/utils"
	"analisefrota/modules/trajeto/trajeto_repo"
)

// Frota ao vivo: onde cada veículo está agora e se a ignição está ligada. A API só dá a última posição do intervalo,
// sem ignição nem velocidade, mas o rastreador tem dois ritmos (conferido em 05/11/2026 com veículos reais):
// com a ignição ligada manda uma posição por minuto; desligada, uma a cada 10 minutos. Então são duas chamadas por
// placa: a última posição das últimas 24 h (t1) e a última entre 5 min e 90 s antes dela (t0). Com t1 recente e t0
// a menos de 3 min de t1, a ignição está ligada, e a distância entre as duas dá a velocidade (≥ 5 km/h = em
// movimento; abaixo = motor ligado parado). Sem t0 nesse trecho = desligado (só o sinal de 10 em 10 min).
// Sem posição há mais de 25 min = sem sinal.
const (
	VivoPlacasMax   = 150
	vivoJanela      = 24 * time.Hour
	vivoAntIni      = 5 * time.Minute // a posição anterior é procurada entre 5 min e 90 s antes da atual
	vivoAntFim      = 90 * time.Second
	vivoLigadoGap   = 3 * time.Minute  // ignição ligada: posições a cada minuto
	vivoLigadoIdade = 4 * time.Minute  // e a última com no máximo 4 min
	vivoSemSinal    = 25 * time.Minute // desligado manda a cada 10 min: mais de 25 min sem nada = sem sinal
	vivoMovimento   = 5.0              // km/h: abaixo disso conta como parado (ruído do GPS)
	vivoCache       = 90 * time.Second // vários painéis abertos no mesmo cliente não repetem as chamadas
	vivoOrcamento   = 40 * time.Second
	StatusMovendo   = "movimento"
	StatusLigado    = "ligado" // motor ligado parado
	StatusDesligado = "desligado"
	StatusSemSinal  = "sem_sinal"
	StatusNaoAchou  = "nao_encontrado"
)

var ErrPlacas = errors.New("informe de 1 a 150 placas, separadas por vírgula")

// VeiculoVivo_t é a situação de um veículo agora
type VeiculoVivo_t struct {
	Placa      string   `json:"placa"`
	Status     string   `json:"status"`
	Data       string   `json:"data,omitempty"` // última posição em UTC, "AAAA-MM-DD HH:MM:SS"
	Lat        *float64 `json:"lat,omitempty"`
	Lon        *float64 `json:"lon,omitempty"`
	MinAtras   *int     `json:"min_atras,omitempty"`  // minutos desde a última posição
	Velocidade *float64 `json:"velocidade,omitempty"` // km/h média entre as duas últimas posições
}

// PontoRastro_t é um ponto do rastro (só com a ignição ligada)
type PontoRastro_t struct {
	Data   string  `json:"data"` // UTC, "AAAA-MM-DD HH:MM:SS"
	Lat    float64 `json:"lat"`
	Lon    float64 `json:"lon"`
	Status string  `json:"status"`
}

// AoVivo_t: Completo = false quando o tempo do pedido acabou antes de todas as placas (peça de novo: as já vistas
// vêm do cache). Rastro: os pontos de ignição ligada de cada placa desde a data pedida.
type AoVivo_t struct {
	Veiculos []VeiculoVivo_t            `json:"veiculos"`
	Rastro   map[string][]PontoRastro_t `json:"rastro"`
	Em       string                     `json:"em"` // hora do servidor (UTC, RFC 3339)
	Faltam   int                        `json:"faltam"`
	Completo bool                       `json:"completo"`
}

type vivoCacheItem_t struct {
	v  VeiculoVivo_t
	em time.Time
}

type vivo_t struct {
	api   gobraxapi.Posicoes_I
	repo  trajeto_repo.RastroRepo_I
	agora func() time.Time
	mu    sync.Mutex
	cache map[string]vivoCacheItem_t
}

type Vivo_I interface {
	// Situacao busca (ou tira do cache) a situação de cada placa e grava no rastro o que veio da API
	Situacao(ctx context.Context, clienteID int64, token string, placas []string) (AoVivo_t, error)
	// Rastro devolve os pontos de ignição ligada das placas desde a data
	Rastro(ctx context.Context, clienteID int64, placas []string, desde time.Time) (map[string][]PontoRastro_t, error)
	// LimparRastro apaga o rastro anterior à data
	LimparRastro(ctx context.Context, antes time.Time) (int64, error)
}

var (
	vivoUnico     *vivo_t
	vivoUnicoOnce sync.Once
)

// NewAoVivo devolve o serviço único (o cache é compartilhado entre os pedidos e a tarefa automática)
func NewAoVivo() Vivo_I {
	vivoUnicoOnce.Do(func() {
		vivoUnico = &vivo_t{api: gobraxapi.NewPosicoes(), repo: trajeto_repo.NewRastroRepo(), agora: time.Now, cache: map[string]vivoCacheItem_t{}}
	})
	return vivoUnico
}

// NewAoVivoCom recebe a API, o repositório e o relógio (usado nos testes)
func NewAoVivoCom(api gobraxapi.Posicoes_I, repo trajeto_repo.RastroRepo_I, agora func() time.Time) Vivo_I {
	return &vivo_t{api: api, repo: repo, agora: agora, cache: map[string]vivoCacheItem_t{}}
}

func (s *vivo_t) Situacao(ctx context.Context, clienteID int64, token string, placas []string) (AoVivo_t, error) {
	ps := LimparPlacas(placas)
	if len(ps) == 0 || len(ps) > VivoPlacasMax {
		return AoVivo_t{}, ErrPlacas
	}
	agora := s.agora()
	res := AoVivo_t{Veiculos: []VeiculoVivo_t{}, Rastro: map[string][]PontoRastro_t{}, Em: agora.UTC().Format(time.RFC3339), Completo: true}
	inicio := time.Now()
	for _, p := range ps {
		chave := strconv.FormatInt(clienteID, 10) + "|" + p
		s.mu.Lock()
		c, ok := s.cache[chave]
		s.mu.Unlock()
		if ok && agora.Sub(c.em) < vivoCache {
			res.Veiculos = append(res.Veiculos, c.v)
			continue
		}
		if time.Since(inicio) > vivoOrcamento {
			res.Faltam++
			res.Completo = false
			continue
		}
		v, err := s.buscar(ctx, token, p, agora)
		if err != nil {
			return AoVivo_t{}, err
		}
		s.mu.Lock()
		s.cache[chave] = vivoCacheItem_t{v: v, em: agora}
		s.mu.Unlock()
		s.gravar(ctx, clienteID, v)
		res.Veiculos = append(res.Veiculos, v)
	}
	return res, nil
}

// gravar põe a posição no rastro; falha no banco não derruba a situação ao vivo (só fica no log)
func (s *vivo_t) gravar(ctx context.Context, clienteID int64, v VeiculoVivo_t) {
	if v.Lat == nil || v.Lon == nil || v.Data == "" {
		return
	}
	t, err := time.ParseInLocation(utils.LayoutAPI, v.Data, time.UTC)
	if err != nil {
		return
	}
	row := trajeto_repo.RastroRow_t{ClienteID: clienteID, Placa: v.Placa, Data: t, Lat: *v.Lat, Lon: *v.Lon, Status: v.Status, Velocidade: v.Velocidade}
	if err := s.repo.SalvarRastro(ctx, row); err != nil {
		logrus.Warnf("ao vivo: gravar rastro de %s: %v", v.Placa, err)
	}
}

func (s *vivo_t) Rastro(ctx context.Context, clienteID int64, placas []string, desde time.Time) (map[string][]PontoRastro_t, error) {
	out := map[string][]PontoRastro_t{}
	ps := LimparPlacas(placas)
	if len(ps) == 0 {
		return out, nil
	}
	rows, err := s.repo.ListarRastro(ctx, clienteID, ps, desde)
	if err != nil {
		return nil, err
	}
	for _, r := range rows {
		out[r.Placa] = append(out[r.Placa], PontoRastro_t{Data: r.Data.UTC().Format(utils.LayoutAPI), Lat: r.Lat, Lon: r.Lon, Status: r.Status})
	}
	return out, nil
}

func (s *vivo_t) LimparRastro(ctx context.Context, antes time.Time) (int64, error) {
	return s.repo.LimparRastro(ctx, antes)
}

func (s *vivo_t) buscar(ctx context.Context, token, placa string, agora time.Time) (VeiculoVivo_t, error) {
	v := VeiculoVivo_t{Placa: placa, Status: StatusSemSinal}
	ult, err := s.api.UltimaPosicao(ctx, token, placa, agora.Add(-vivoJanela), agora)
	if errors.Is(err, gobraxapi.ErrVeiculoNaoEncontrado) {
		v.Status = StatusNaoAchou
		return v, nil
	}
	if err != nil || ult == nil {
		return v, err
	}
	t1, err := time.ParseInLocation(utils.LayoutAPI, ult.Data, time.UTC)
	if err != nil {
		return v, nil
	}
	lat, lon := ult.Lat, ult.Lon
	idade := agora.Sub(t1)
	min := int(math.Max(0, idade.Minutes()))
	v.Data, v.Lat, v.Lon, v.MinAtras = ult.Data, &lat, &lon, &min
	if idade > vivoSemSinal {
		return v, nil
	}
	v.Status = StatusDesligado
	if idade > vivoLigadoIdade {
		return v, nil // a última posição já tem uns minutos: é o sinal de 10 em 10 min
	}
	ant, err := s.api.UltimaPosicao(ctx, token, placa, t1.Add(-vivoAntIni), t1.Add(-vivoAntFim))
	if err != nil || ant == nil {
		return v, err
	}
	t0, err := time.ParseInLocation(utils.LayoutAPI, ant.Data, time.UTC)
	dt := t1.Sub(t0)
	if err != nil || dt <= 0 || dt > vivoLigadoGap {
		return v, nil
	}
	vel := math.Round(distKm(ant.Lat, ant.Lon, lat, lon)/dt.Hours()*10) / 10
	v.Velocidade = &vel
	v.Status = StatusLigado
	if vel >= vivoMovimento {
		v.Status = StatusMovendo
	}
	return v, nil
}

// LimparPlacas deixa só as placas válidas, sem hífen, maiúsculas e sem repetir
func LimparPlacas(placas []string) []string {
	vistas := map[string]bool{}
	out := []string{}
	for _, p := range placas {
		c := utils.ChavePlaca(p)
		if !placaValida.MatchString(c) || vistas[c] {
			continue
		}
		vistas[c] = true
		out = append(out, c)
	}
	return out
}

// distKm: distância em linha reta entre dois pontos (haversine)
func distKm(lat1, lon1, lat2, lon2 float64) float64 {
	const r = math.Pi / 180
	dLa, dLo := (lat2-lat1)*r, (lon2-lon1)*r
	h := math.Sin(dLa/2)*math.Sin(dLa/2) + math.Cos(lat1*r)*math.Cos(lat2*r)*math.Sin(dLo/2)*math.Sin(dLo/2)
	return 2 * 6371 * math.Asin(math.Sqrt(h))
}

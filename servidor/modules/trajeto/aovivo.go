package trajeto

import (
	"context"
	"errors"
	"math"
	"strconv"
	"sync"
	"time"

	"analisefrota/adapters/gobraxapi"
	"analisefrota/common/utils"
)

// Frota ao vivo: onde cada veículo está agora. A API só dá a última posição do intervalo, sem velocidade nem ignição,
// então são duas chamadas por placa: a última posição das últimas 24 h e a última até 4 min antes dela. A distância
// entre as duas dá a velocidade média nesse trecho (em movimento ou parado). Sem posição há mais de 30 min = sem sinal.
const (
	VivoPlacasMax  = 150
	vivoJanela     = 24 * time.Hour
	vivoRecuo      = 4 * time.Minute  // a posição anterior é a última até 4 min antes da atual
	vivoSemSinal   = 30 * time.Minute // sem posição nova há mais que isso = sem sinal
	vivoTrechoMax  = 30 * time.Minute // as duas posições mais distantes que isso não dão velocidade
	vivoMovimento  = 5.0              // km/h: abaixo disso conta como parado (ruído do GPS)
	vivoCache      = 90 * time.Second // vários painéis abertos no mesmo cliente não repetem as chamadas
	vivoOrcamento  = 40 * time.Second
	StatusMovendo  = "movimento"
	StatusParado   = "parado"
	StatusSemSinal = "sem_sinal"
	StatusNaoAchou = "nao_encontrado"
)

var ErrPlacas = errors.New("informe de 1 a 150 placas, separadas por vírgula")

// VeiculoVivo_t é a situação de um veículo agora
type VeiculoVivo_t struct {
	Placa      string   `json:"placa"`
	Status     string   `json:"status"`
	Data       string   `json:"data,omitempty"` // última posição, como a API mandou (UTC)
	Lat        *float64 `json:"lat,omitempty"`
	Lon        *float64 `json:"lon,omitempty"`
	MinAtras   *int     `json:"min_atras,omitempty"`  // minutos desde a última posição
	Velocidade *float64 `json:"velocidade,omitempty"` // km/h média entre as duas últimas posições
}

// AoVivo_t: Completo = false quando o tempo do pedido acabou antes de todas as placas (peça de novo: as já vistas
// vêm do cache)
type AoVivo_t struct {
	Veiculos []VeiculoVivo_t `json:"veiculos"`
	Em       string          `json:"em"` // hora do servidor (UTC, RFC 3339)
	Faltam   int             `json:"faltam"`
	Completo bool            `json:"completo"`
}

type vivoCacheItem_t struct {
	v  VeiculoVivo_t
	em time.Time
}

type vivo_t struct {
	api   gobraxapi.Posicoes_I
	agora func() time.Time
	mu    sync.Mutex
	cache map[string]vivoCacheItem_t
}

type Vivo_I interface {
	Situacao(ctx context.Context, clienteID int64, token string, placas []string) (AoVivo_t, error)
}

var (
	vivoUnico     *vivo_t
	vivoUnicoOnce sync.Once
)

// NewAoVivo devolve o serviço único (o cache é compartilhado entre os pedidos)
func NewAoVivo() Vivo_I {
	vivoUnicoOnce.Do(func() { vivoUnico = &vivo_t{api: gobraxapi.NewPosicoes(), agora: time.Now, cache: map[string]vivoCacheItem_t{}} })
	return vivoUnico
}

// NewAoVivoCom recebe a API e o relógio (usado nos testes)
func NewAoVivoCom(api gobraxapi.Posicoes_I, agora func() time.Time) Vivo_I {
	return &vivo_t{api: api, agora: agora, cache: map[string]vivoCacheItem_t{}}
}

func (s *vivo_t) Situacao(ctx context.Context, clienteID int64, token string, placas []string) (AoVivo_t, error) {
	ps := limparPlacas(placas)
	if len(ps) == 0 || len(ps) > VivoPlacasMax {
		return AoVivo_t{}, ErrPlacas
	}
	agora := s.agora()
	res := AoVivo_t{Veiculos: []VeiculoVivo_t{}, Em: agora.UTC().Format(time.RFC3339), Completo: true}
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
		res.Veiculos = append(res.Veiculos, v)
	}
	return res, nil
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
	min := int(math.Max(0, agora.Sub(t1).Minutes()))
	v.Data, v.Lat, v.Lon, v.MinAtras = ult.Data, &lat, &lon, &min
	if agora.Sub(t1) > vivoSemSinal {
		return v, nil
	}
	v.Status = StatusParado
	ant, err := s.api.UltimaPosicao(ctx, token, placa, agora.Add(-vivoJanela), t1.Add(-vivoRecuo))
	if err != nil || ant == nil {
		return v, err
	}
	t0, err := time.ParseInLocation(utils.LayoutAPI, ant.Data, time.UTC)
	dt := t1.Sub(t0)
	if err != nil || dt <= 0 || dt > vivoTrechoMax {
		return v, nil
	}
	vel := math.Round(distKm(ant.Lat, ant.Lon, lat, lon)/dt.Hours()*10) / 10
	v.Velocidade = &vel
	if vel >= vivoMovimento {
		v.Status = StatusMovendo
	}
	return v, nil
}

// limparPlacas deixa só as placas válidas, sem hífen, maiúsculas e sem repetir
func limparPlacas(placas []string) []string {
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

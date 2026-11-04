// Package trajeto monta o trajeto de um veículo num dia a partir da API de posições. A API só devolve a última
// posição de cada intervalo, então o dia é perguntado em janelas de 5 minutos; cada janela consultada fica
// guardada no banco e só é pedida de novo enquanto ainda podia mudar.
package trajeto

import (
	"context"
	"time"

	"analisefrota/adapters/gobraxapi"
	"analisefrota/modules/trajeto/trajeto_repo"
)

type Service_I interface {
	// Amostrar devolve os pontos do trajeto da placa entre as horas de e ate (0 a 24) do dia de Brasília
	Amostrar(ctx context.Context, clienteID int64, token, placa, dia string, de, ate int) (Trajeto_t, error)
}

type trajeto_t struct {
	repo  trajeto_repo.Repository_I
	api   gobraxapi.Posicoes_I
	agora func() time.Time
}

func NewTrajeto() Service_I {
	return &trajeto_t{repo: trajeto_repo.NewRepository(), api: gobraxapi.NewPosicoes(), agora: time.Now}
}

// NewTrajetoCom recebe o repositório, a API e o relógio (usado nos testes)
func NewTrajetoCom(repo trajeto_repo.Repository_I, api gobraxapi.Posicoes_I, agora func() time.Time) Service_I {
	return &trajeto_t{repo: repo, api: api, agora: agora}
}

func (t *trajeto_t) Amostrar(ctx context.Context, clienteID int64, token, placa, dia string, de, ate int) (Trajeto_t, error) {
	agora := t.agora()
	placa, err := validar(placa, dia, de, ate, agora)
	if err != nil {
		return Trajeto_t{}, err
	}
	js, err := janelas(dia, de, ate, agora)
	if err != nil {
		return Trajeto_t{}, err
	}
	res := Trajeto_t{Pontos: []Ponto_t{}, Janelas: len(js), Completo: true}
	if len(js) == 0 {
		return res, nil
	}
	salvas, err := t.repo.Listar(ctx, clienteID, placa, js[0], js[len(js)-1].Add(Passo))
	if err != nil {
		return Trajeto_t{}, err
	}
	porJanela := map[int64]trajeto_repo.PosicaoRow_t{}
	for _, r := range salvas {
		porJanela[r.JanelaIni.Unix()] = r
	}
	usadas := make([]trajeto_repo.PosicaoRow_t, 0, len(js))
	inicio := time.Now()
	for _, j := range js {
		if r, ok := porJanela[j.Unix()]; ok && definitiva(r) {
			usadas = append(usadas, r)
			continue
		}
		// o tempo do pedido acabou: devolve o que já tem; o próximo pedido continua de onde parou
		if time.Since(inicio) > orcamento {
			res.Faltam++
			res.Completo = false
			continue
		}
		pos, err := t.api.UltimaPosicao(ctx, token, placa, j, j.Add(Passo-time.Second))
		if err != nil {
			return Trajeto_t{}, err
		}
		row := trajeto_repo.PosicaoRow_t{ClienteID: clienteID, Placa: placa, JanelaIni: j, Vazio: pos == nil, BuscadoEm: t.agora()}
		if pos != nil {
			lat, lon := pos.Lat, pos.Lon
			row.Data, row.Lat, row.Lon = pos.Data, &lat, &lon
		}
		if err := t.repo.Salvar(ctx, row); err != nil {
			return Trajeto_t{}, err
		}
		usadas = append(usadas, row)
	}
	res.Pontos = pontosDe(usadas)
	return res, nil
}

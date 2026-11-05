package gobraxapi

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// ErrVeiculoNaoEncontrado: a API não conhece a placa (erro 400-1 "Vehicle not found")
var ErrVeiculoNaoEncontrado = errors.New("a API não encontrou o veículo (placa não cadastrada na Gobrax)")

// Posicoes_I fica fora do Client_I de propósito: a busca de hora em hora não usa posição, e os dublês
// de teste dela não precisam implementar este método.
type Posicoes_I interface {
	// UltimaPosicao devolve a última posição do veículo na janela (nil quando a API não tem posição nela).
	// A API só devolve a ÚLTIMA posição do intervalo, uma placa por chamada (limite de 10 por segundo).
	UltimaPosicao(ctx context.Context, token, placa string, ini, fim time.Time) (*Posicao_t, error)
}

// NewPosicoes devolve o mesmo cliente único (e a mesma fila de chamadas) do NewClient
func NewPosicoes() Posicoes_I {
	return NewClient().(*Client_t)
}

// NewPosicoesCom monta um cliente com base e HTTP próprios (usado nos testes)
func NewPosicoesCom(base string, hc *http.Client, intervalo time.Duration) Posicoes_I {
	return NewClientCom(base, hc, intervalo).(*Client_t)
}

func (c *Client_t) UltimaPosicao(ctx context.Context, token, placa string, ini, fim time.Time) (*Posicao_t, error) {
	var r registros_t[PosicaoRec_t]
	if err := c.get(ctx, "v1/vehicle-last-position", janela(ini, fim, placa), token, &r); err != nil {
		if errors.Is(err, ErrSemDados) {
			return nil, nil
		}
		return nil, err
	}
	for _, rec := range r.Lista {
		p := rec.LastPosition
		if p.Lat.Valor == nil || p.Lon.Valor == nil || p.Date == "" {
			continue
		}
		return &Posicao_t{Data: dataPosicaoUTC(p.Date), Lat: *p.Lat.Valor, Lon: *p.Lon.Valor}, nil
	}
	return nil, nil
}

// dataPosicaoUTC: a janela pedida (startDate/endDate) é lida em UTC, mas lastPosition.date volta no horário de Brasília
// (conferido em 05/11/2026: a última posição de um veículo rodando vem sempre 3 h antes do fim da janela pedida).
// Daqui para a frente tudo trata a data da posição como UTC, como as datas de entrada. Formato inesperado fica como veio.
func dataPosicaoUTC(data string) string {
	t, err := time.ParseInLocation("2006-01-02 15:04:05", data, time.FixedZone("BRT", -3*60*60))
	if err != nil {
		return data
	}
	return t.UTC().Format("2006-01-02 15:04:05")
}

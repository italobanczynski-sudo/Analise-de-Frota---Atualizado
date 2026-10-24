package gobraxapi

import (
	"context"
	"net/url"
)

type Premiacao_I interface {
	// Premiacao devolve nota, km e bonificação de cada motorista por mês (meses no formato MM-AAAA)
	Premiacao(ctx context.Context, token, mesIni, mesFim string) ([]PremiacaoRec_t, error)
}

func (c *Client_t) Premiacao(ctx context.Context, token, mesIni, mesFim string) ([]PremiacaoRec_t, error) {
	var r registros_t[PremiacaoRec_t]
	err := c.get(ctx, "v2/driversOverview", url.Values{"documentNumbers": {""}, "startDate": {mesIni}, "endDate": {mesFim}}, token, &r)
	return r.Lista, err
}

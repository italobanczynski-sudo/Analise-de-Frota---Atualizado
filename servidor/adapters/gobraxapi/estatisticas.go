package gobraxapi

import (
	"context"
	"time"
)

type Estatisticas_I interface {
	// Estatisticas devolve km, litros, frenagens, velocidade média e odômetro de cada veículo no período
	Estatisticas(ctx context.Context, token string, ini, fim time.Time) ([]EstatisticaRec_t, error)
}

func (c *Client_t) Estatisticas(ctx context.Context, token string, ini, fim time.Time) ([]EstatisticaRec_t, error) {
	var r registros_t[EstatisticaRec_t]
	err := c.get(ctx, "v1/vehicle-statistics", janela(ini, fim), token, &r)
	return r.Lista, err
}

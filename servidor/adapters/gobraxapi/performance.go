package gobraxapi

import (
	"context"
	"time"
)

type Performance_I interface {
	// Performance devolve os indicadores de condução e os motoristas vinculados de cada veículo no período
	Performance(ctx context.Context, token string, ini, fim time.Time) ([]PerformanceRec_t, error)
}

func (c *Client_t) Performance(ctx context.Context, token string, ini, fim time.Time) ([]PerformanceRec_t, error) {
	var r registros_t[PerformanceRec_t]
	err := c.get(ctx, "v1/vehicle-performance", janela(ini, fim), token, &r)
	return r.Lista, err
}

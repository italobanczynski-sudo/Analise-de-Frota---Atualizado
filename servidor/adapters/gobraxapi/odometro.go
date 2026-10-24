package gobraxapi

import (
	"context"
	"time"
)

type Odometro_I interface {
	// Odometro devolve o último odômetro de cada veículo no período
	Odometro(ctx context.Context, token string, ini, fim time.Time) ([]OdometroRec_t, error)
}

func (c *Client_t) Odometro(ctx context.Context, token string, ini, fim time.Time) ([]OdometroRec_t, error) {
	var r registros_t[OdometroRec_t]
	err := c.get(ctx, "v2/vehicle-odometer", janela(ini, fim), token, &r)
	return r.Lista, err
}

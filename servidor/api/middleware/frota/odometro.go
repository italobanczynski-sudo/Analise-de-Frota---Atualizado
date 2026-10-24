package frota

import (
	"context"

	"analisefrota/api/middleware"
	modcliente "analisefrota/modules/cliente"
	"analisefrota/modules/odometro"
)

// Odometros devolve o último odômetro de cada veículo do cliente
func (m Frota_t) Odometros(ctx context.Context) *middleware.Response_t {
	if _, err := modcliente.NewCliente().Buscar(ctx, m.ClienteID); err != nil {
		return middleware.RetErro(err)
	}
	ods, err := odometro.NewOdometro().Listar(ctx, m.ClienteID)
	if err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetOkData(ods)
}

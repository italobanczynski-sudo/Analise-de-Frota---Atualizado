package frota

import (
	"context"

	"analisefrota/api/middleware"
	modcliente "analisefrota/modules/cliente"
	"analisefrota/modules/diario"
)

// Registros devolve o dia a dia de cada veículo do cliente no período
func (m Frota_t) Registros(ctx context.Context) *middleware.Response_t {
	if _, err := modcliente.NewCliente().Buscar(ctx, m.ClienteID); err != nil {
		return middleware.RetErro(err)
	}
	regs, err := diario.NewDiario().Listar(ctx, m.ClienteID, m.Inicio, m.Fim)
	if err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetOkData(regs)
}

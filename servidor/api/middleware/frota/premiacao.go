package frota

import (
	"context"

	"analisefrota/api/middleware"
	modcliente "analisefrota/modules/cliente"
	"analisefrota/modules/premiacao"
)

// Premiacoes devolve nota, km e bonificação dos motoristas do cliente (um mês ou todos)
func (m Frota_t) Premiacoes(ctx context.Context) *middleware.Response_t {
	if _, err := modcliente.NewCliente().Buscar(ctx, m.ClienteID); err != nil {
		return middleware.RetErro(err)
	}
	ps, err := premiacao.NewPremiacao().Listar(ctx, m.ClienteID, m.Mes)
	if err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetOkData(ps)
}

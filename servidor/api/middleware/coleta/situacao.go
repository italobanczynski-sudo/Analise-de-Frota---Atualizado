package coleta

import (
	"context"

	"analisefrota/api/middleware"
	modcliente "analisefrota/modules/cliente"
	modcoleta "analisefrota/modules/coleta"
)

// Situacao devolve como está a busca do cliente
func (m Coleta_t) Situacao(ctx context.Context) *middleware.Response_t {
	if _, err := modcliente.NewCliente().Buscar(ctx, m.ClienteID); err != nil {
		return middleware.RetErro(err)
	}
	s, err := modcoleta.NewColeta().Situacao(ctx, m.ClienteID)
	if err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetOkData(s)
}

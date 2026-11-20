package cliente

import (
	"context"

	"analisefrota/api/middleware"
	modcliente "analisefrota/modules/cliente"
)

// Excluir apaga o cliente e todos os dados dele (só administrador; não tem volta)
func (m Cliente_t) Excluir(ctx context.Context) *middleware.Response_t {
	if err := modcliente.NewCliente().Excluir(ctx, m.ID); err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetOkData(map[string]bool{"excluido": true})
}

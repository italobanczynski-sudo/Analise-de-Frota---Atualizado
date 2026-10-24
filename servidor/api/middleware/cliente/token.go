package cliente

import (
	"context"

	"analisefrota/api/middleware"
	modcliente "analisefrota/modules/cliente"
)

// SalvarToken troca o token da API do cliente (cifrado no banco; nunca devolvido)
func (m Cliente_t) SalvarToken(ctx context.Context) *middleware.Response_t {
	if err := modcliente.NewCliente().SalvarToken(ctx, m.ID, m.Token); err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetOkData(map[string]bool{"salvo": true})
}

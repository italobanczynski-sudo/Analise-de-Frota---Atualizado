package cliente

import (
	"context"

	"analisefrota/api/middleware"
	modcliente "analisefrota/modules/cliente"
)

// Criar cadastra o cliente e, se veio token, já o salva cifrado
func (m Cliente_t) Criar(ctx context.Context) *middleware.Response_t {
	svc := modcliente.NewCliente()
	c, err := svc.Criar(ctx, m.Nome, m.HubspotID)
	if err != nil {
		return middleware.RetErro(err)
	}
	if m.Token != "" {
		if err := svc.SalvarToken(ctx, c.ID, m.Token); err != nil {
			return middleware.RetErro(err)
		}
		c.TemToken = true
	}
	return middleware.RetCriado(c)
}

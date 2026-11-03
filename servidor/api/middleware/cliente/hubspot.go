package cliente

import (
	"context"

	"analisefrota/api/middleware"
	modcliente "analisefrota/modules/cliente"
	"analisefrota/modules/hubspot"
)

// Hubspot devolve a leitura do cliente no HubSpot (guardada por alguns minutos; atualizar busca de novo)
func (m Cliente_t) Hubspot(ctx context.Context, atualizar bool) *middleware.Response_t {
	c, err := modcliente.NewCliente().Buscar(ctx, m.ID)
	if err != nil {
		return middleware.RetErro(err)
	}
	r, err := hubspot.NewHubspot().Resumo(ctx, c.Nome, c.HubspotID, atualizar)
	if err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetOkData(r)
}

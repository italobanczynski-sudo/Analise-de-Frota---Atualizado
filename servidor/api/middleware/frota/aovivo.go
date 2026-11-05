package frota

import (
	"context"

	"analisefrota/api/middleware"
	modcliente "analisefrota/modules/cliente"
	"analisefrota/modules/trajeto"
)

// AoVivo_t é o pedido da situação atual das placas de um cliente
type AoVivo_t struct {
	ClienteID int64
	Placas    []string
}

// Buscar devolve onde cada placa está agora (em movimento, parada ou sem sinal). O token é decifrado só para
// chamar a API e nunca sai daqui.
func (m AoVivo_t) Buscar(ctx context.Context) *middleware.Response_t {
	cli := modcliente.NewCliente()
	if _, err := cli.Buscar(ctx, m.ClienteID); err != nil {
		return middleware.RetErro(err)
	}
	token, err := cli.Token(ctx, m.ClienteID)
	if err != nil {
		return middleware.RetErro(err)
	}
	r, err := trajeto.NewAoVivo().Situacao(ctx, m.ClienteID, token, m.Placas)
	if err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetOkData(r)
}

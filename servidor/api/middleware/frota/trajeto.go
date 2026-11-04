package frota

import (
	"context"

	"analisefrota/api/middleware"
	modcliente "analisefrota/modules/cliente"
	"analisefrota/modules/trajeto"
)

// Trajeto_t é o pedido do trajeto de um veículo num dia (horas de Brasília, de 0 a 24)
type Trajeto_t struct {
	ClienteID int64
	Placa     string
	Dia       string // AAAA-MM-DD
	De, Ate   int
}

// Buscar devolve os pontos do trajeto, amostrados a cada 5 minutos. O token do cliente é decifrado só para
// chamar a API e nunca sai daqui.
func (m Trajeto_t) Buscar(ctx context.Context) *middleware.Response_t {
	cli := modcliente.NewCliente()
	if _, err := cli.Buscar(ctx, m.ClienteID); err != nil {
		return middleware.RetErro(err)
	}
	token, err := cli.Token(ctx, m.ClienteID)
	if err != nil {
		return middleware.RetErro(err)
	}
	t, err := trajeto.NewTrajeto().Amostrar(ctx, m.ClienteID, token, m.Placa, m.Dia, m.De, m.Ate)
	if err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetOkData(t)
}

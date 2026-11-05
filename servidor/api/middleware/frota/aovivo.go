package frota

import (
	"context"
	"time"

	"analisefrota/api/middleware"
	modcliente "analisefrota/modules/cliente"
	"analisefrota/modules/trajeto"
)

// AoVivo_t é o pedido da situação atual das placas de um cliente e do rastro delas desde Desde
type AoVivo_t struct {
	ClienteID int64
	Placas    []string
	Desde     time.Time
}

// Buscar devolve onde cada placa está agora (em movimento, motor ligado parado, desligada ou sem sinal) e o rastro
// de ignição ligada desde Desde. O token é decifrado só para chamar a API e nunca sai daqui.
func (m AoVivo_t) Buscar(ctx context.Context) *middleware.Response_t {
	cli := modcliente.NewCliente()
	if _, err := cli.Buscar(ctx, m.ClienteID); err != nil {
		return middleware.RetErro(err)
	}
	token, err := cli.Token(ctx, m.ClienteID)
	if err != nil {
		return middleware.RetErro(err)
	}
	svc := trajeto.NewAoVivo()
	r, err := svc.Situacao(ctx, m.ClienteID, token, m.Placas)
	if err != nil {
		return middleware.RetErro(err)
	}
	if r.Rastro, err = svc.Rastro(ctx, m.ClienteID, m.Placas, m.Desde); err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetOkData(r)
}

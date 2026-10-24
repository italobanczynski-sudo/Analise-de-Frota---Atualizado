package usuario

import (
	"context"

	"analisefrota/api/middleware"
	modusuario "analisefrota/modules/usuario"
)

// NovoConvite gera outro link para a pessoa (o anterior deixa de valer). Serve para quem perdeu o link
// ou esqueceu a senha: ao aceitar, a senha nova substitui a antiga.
func (m Usuario_t) NovoConvite(ctx context.Context) *middleware.Response_t {
	svc := modusuario.NewUsuario()
	conv, err := svc.NovoConvite(ctx, m.ID)
	if err != nil {
		return middleware.RetErro(err)
	}
	u, err := svc.Buscar(ctx, m.ID)
	if err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetOkData(ComConvite_t{Usuario: u, Link: linkConvite(conv.Token), ExpiraEm: conv.ExpiraEm})
}

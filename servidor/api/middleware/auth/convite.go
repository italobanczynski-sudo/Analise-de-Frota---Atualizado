package auth

import (
	"context"

	"analisefrota/api/middleware"
	"analisefrota/modules/usuario"
)

// AceitarConvite grava a senha escolhida pela pessoa e já a deixa logada
func (m Auth_t) AceitarConvite(ctx context.Context) (*middleware.Response_t, *Entrada_t) {
	u, err := usuario.NewUsuario().AceitarConvite(ctx, m.Token, m.Senha)
	if err != nil {
		return middleware.RetErro(err), nil
	}
	return abrirSessao(ctx, u)
}

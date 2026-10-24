package auth

import (
	"context"

	"analisefrota/api/middleware"
	"analisefrota/modules/sessao"
)

// Sair encerra a sessão do cookie
func (m Auth_t) Sair(ctx context.Context) *middleware.Response_t {
	if err := sessao.NewSessao().Encerrar(ctx, m.Token); err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetOkData(map[string]bool{"saiu": true})
}

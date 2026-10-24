package auth

import (
	"context"
	"errors"
	"time"

	"analisefrota/api/middleware"
	"analisefrota/modules/sessao"
	"analisefrota/modules/usuario"
)

// Entrar confere e-mail e senha e abre uma sessão
func (m Auth_t) Entrar(ctx context.Context) (*middleware.Response_t, *Entrada_t) {
	agora := time.Now()
	if bloqueado(m.Email, m.Origem, agora) {
		return middleware.RetErro(middleware.ErrMuitasTentativas), nil
	}
	u, err := usuario.NewUsuario().Autenticar(ctx, m.Email, m.Senha)
	if err != nil {
		if errors.Is(err, usuario.ErrLogin) {
			registrarErro(m.Email, m.Origem, agora)
		}
		return middleware.RetErro(err), nil
	}
	limparErros(m.Email, m.Origem)
	return abrirSessao(ctx, u)
}

func abrirSessao(ctx context.Context, u usuario.Usuario_t) (*middleware.Response_t, *Entrada_t) {
	ses, err := sessao.NewSessao().Criar(ctx, u.ID)
	if err != nil {
		return middleware.RetErro(err), nil
	}
	return middleware.RetOkData(u), &Entrada_t{Usuario: u, Token: ses.Token, ExpiraEm: ses.ExpiraEm}
}

package auth

import (
	"context"

	"analisefrota/modules/sessao"
	"analisefrota/modules/usuario"
)

// Identificar diz quem é o dono do cookie; usuário desativado depois do login perde o acesso na hora
func (m Auth_t) Identificar(ctx context.Context) (usuario.Usuario_t, error) {
	id, err := sessao.NewSessao().Validar(ctx, m.Token)
	if err != nil {
		return usuario.Usuario_t{}, err
	}
	u, err := usuario.NewUsuario().Buscar(ctx, id)
	if err != nil {
		return usuario.Usuario_t{}, sessao.ErrSessaoInvalida
	}
	if !u.Ativo {
		return usuario.Usuario_t{}, usuario.ErrInativo
	}
	return u, nil
}

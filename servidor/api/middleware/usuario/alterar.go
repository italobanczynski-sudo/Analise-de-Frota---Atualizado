package usuario

import (
	"context"

	"github.com/sirupsen/logrus"

	"analisefrota/api/middleware"
	"analisefrota/modules/sessao"
	modusuario "analisefrota/modules/usuario"
)

// Alterar troca o perfil e ativa/desativa a pessoa. Se algo mudou, os logins dela são encerrados
// (ela entra de novo já com o perfil novo; desativada, não entra mais).
func (m Usuario_t) Alterar(ctx context.Context) *middleware.Response_t {
	svc := modusuario.NewUsuario()
	antes, err := svc.Buscar(ctx, m.ID)
	if err != nil {
		return middleware.RetErro(err)
	}
	u, err := svc.Alterar(ctx, m.ID, m.Papel, m.Ativo)
	if err != nil {
		return middleware.RetErro(err)
	}
	if antes.Papel != u.Papel || antes.Ativo != u.Ativo {
		if err := sessao.NewSessao().EncerrarDoUsuario(ctx, u.ID); err != nil {
			return middleware.RetErro(err)
		}
		logrus.Infof("usuarios: %d alterou %d: %s/%v → %s/%v", m.QuemPede, u.ID, antes.Papel, antes.Ativo, u.Papel, u.Ativo)
	}
	return middleware.RetOkData(u)
}

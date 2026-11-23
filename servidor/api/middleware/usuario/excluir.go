package usuario

import (
	"context"

	"github.com/sirupsen/logrus"

	"analisefrota/api/middleware"
	modusuario "analisefrota/modules/usuario"
)

// Excluir apaga a pessoa de vez: os logins dela caem na hora (as sessões saem junto) e o que ela registrou fica sem autor
func (m Usuario_t) Excluir(ctx context.Context) *middleware.Response_t {
	u, err := modusuario.NewUsuario().Excluir(ctx, m.ID, m.QuemPede)
	if err != nil {
		return middleware.RetErro(err)
	}
	logrus.Infof("usuarios: %d excluiu %d (%s)", m.QuemPede, u.ID, u.Papel)
	return middleware.RetOkData(map[string]bool{"excluido": true})
}

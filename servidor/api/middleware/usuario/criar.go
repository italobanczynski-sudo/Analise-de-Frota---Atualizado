package usuario

import (
	"context"

	"github.com/sirupsen/logrus"

	"analisefrota/api/middleware"
	modusuario "analisefrota/modules/usuario"
)

// Criar cadastra a pessoa e devolve o link de convite para ela criar a senha
func (m Usuario_t) Criar(ctx context.Context) *middleware.Response_t {
	u, conv, err := modusuario.NewUsuario().Criar(ctx, m.Nome, m.Email, m.Papel)
	if err != nil {
		return middleware.RetErro(err)
	}
	logrus.Infof("usuarios: %d cadastrou %d (%s)", m.QuemPede, u.ID, u.Papel)
	return middleware.RetCriado(ComConvite_t{Usuario: u, Link: linkConvite(conv.Token), ExpiraEm: conv.ExpiraEm})
}

// CriarAdmin cria o primeiro administrador pela linha de comando e devolve o link de convite
func CriarAdmin(ctx context.Context, nome, email string) (string, error) {
	_, conv, err := modusuario.NewUsuario().Criar(ctx, nome, email, modusuario.PapelAdmin)
	if err != nil {
		return "", err
	}
	return linkConvite(conv.Token), nil
}

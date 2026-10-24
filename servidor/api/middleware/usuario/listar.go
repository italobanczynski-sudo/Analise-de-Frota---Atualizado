package usuario

import (
	"context"

	"analisefrota/api/middleware"
	modusuario "analisefrota/modules/usuario"
)

// Listar devolve todas as pessoas cadastradas
func (m Usuario_t) Listar(ctx context.Context) *middleware.Response_t {
	us, err := modusuario.NewUsuario().Listar(ctx)
	if err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetOkData(us)
}

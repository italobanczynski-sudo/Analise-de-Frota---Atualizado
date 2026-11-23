package usuario

import (
	"github.com/gin-gonic/gin"

	"analisefrota/api/controller"
	mwusuario "analisefrota/api/middleware/usuario"
)

// Excluir: DELETE /api/v1/usuarios/:id (só administrador; não tem volta)
func Excluir(c *gin.Context) {
	id, ok := controller.IDParam(c)
	if !ok {
		return
	}
	controller.Response(c, mwusuario.Usuario_t{ID: id, QuemPede: controller.UsuarioLogado(c).ID}.Excluir(c.Request.Context()))
}

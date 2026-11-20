package cliente

import (
	"github.com/gin-gonic/gin"

	"analisefrota/api/controller"
	mwcliente "analisefrota/api/middleware/cliente"
)

// Excluir: DELETE /api/v1/clientes/:id
func Excluir(c *gin.Context) {
	id, ok := controller.IDParam(c)
	if !ok {
		return
	}
	controller.Response(c, mwcliente.Cliente_t{ID: id}.Excluir(c.Request.Context()))
}

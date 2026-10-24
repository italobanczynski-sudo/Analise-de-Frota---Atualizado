package cliente

import (
	"github.com/gin-gonic/gin"

	"analisefrota/api/controller"
	mwcliente "analisefrota/api/middleware/cliente"
)

// Listar: GET /api/v1/clientes
func Listar(c *gin.Context) {
	controller.Response(c, mwcliente.Cliente_t{}.Listar(c.Request.Context()))
}

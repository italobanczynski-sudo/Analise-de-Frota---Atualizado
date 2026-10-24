package cliente

import (
	"github.com/gin-gonic/gin"

	"analisefrota/api/controller"
	mwcliente "analisefrota/api/middleware/cliente"
	"analisefrota/common/utils"
)

// SalvarToken: PUT /api/v1/clientes/:id/token {"token"}
func SalvarToken(c *gin.Context) {
	id, ok := controller.IDParam(c)
	if !ok {
		return
	}
	var req TokenRequest
	if err := c.ShouldBindJSON(&req); err != nil || utils.ValidateStruct(req) != nil {
		controller.RetFail(c, "envie o token da API Gobrax no campo token")
		return
	}
	controller.Response(c, mwcliente.Cliente_t{ID: id, Token: req.Token}.SalvarToken(c.Request.Context()))
}

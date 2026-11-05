package cliente

import (
	"github.com/gin-gonic/gin"

	"analisefrota/api/controller"
	mwcliente "analisefrota/api/middleware/cliente"
	"analisefrota/common/utils"
)

// Criar: POST /api/v1/clientes {"nome", "token"}
func Criar(c *gin.Context) {
	var req CriarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		controller.RetFail(c, "corpo da requisição inválido: envie JSON com nome e token")
		return
	}
	if err := utils.ValidateStruct(req); err != nil {
		controller.RetFail(c, "dados inválidos: o nome é obrigatório (até 120 letras)")
		return
	}
	controller.Response(c, mwcliente.Cliente_t{Nome: req.Nome, Token: req.Token}.Criar(c.Request.Context()))
}

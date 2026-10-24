package cliente

import (
	"github.com/gin-gonic/gin"

	"analisefrota/api/controller"
	mwcliente "analisefrota/api/middleware/cliente"
	"analisefrota/common/utils"
)

// Criar: POST /api/v1/clientes {"nome", "hubspot_id", "token"}
func Criar(c *gin.Context) {
	var req CriarRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		controller.RetFail(c, "corpo da requisição inválido: envie JSON com nome, hubspot_id e token")
		return
	}
	if err := utils.ValidateStruct(req); err != nil {
		controller.RetFail(c, "dados inválidos: o nome é obrigatório (até 120 letras) e o ID do HubSpot tem só números")
		return
	}
	controller.Response(c, mwcliente.Cliente_t{Nome: req.Nome, HubspotID: req.HubspotID, Token: req.Token}.Criar(c.Request.Context()))
}

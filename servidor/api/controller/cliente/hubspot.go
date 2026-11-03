package cliente

import (
	"github.com/gin-gonic/gin"

	"analisefrota/api/controller"
	mwcliente "analisefrota/api/middleware/cliente"
)

// Hubspot: GET /api/v1/clientes/:id/hubspot[?atualizar=1] — perfil e relacionamento do cliente no HubSpot
func Hubspot(c *gin.Context) {
	id, ok := controller.IDParam(c)
	if !ok {
		return
	}
	controller.Response(c, mwcliente.Cliente_t{ID: id}.Hubspot(c.Request.Context(), c.Query("atualizar") == "1"))
}

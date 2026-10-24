package coleta

import (
	"github.com/gin-gonic/gin"

	"analisefrota/api/controller"
	mwcoleta "analisefrota/api/middleware/coleta"
)

// Iniciar: POST /api/v1/clientes/:id/coleta (busca agora, em segundo plano)
func Iniciar(c *gin.Context) {
	id, ok := controller.IDParam(c)
	if !ok {
		return
	}
	controller.Response(c, mwcoleta.Coleta_t{ClienteID: id}.Iniciar(c.Request.Context()))
}

// Situacao: GET /api/v1/clientes/:id/coleta
func Situacao(c *gin.Context) {
	id, ok := controller.IDParam(c)
	if !ok {
		return
	}
	controller.Response(c, mwcoleta.Coleta_t{ClienteID: id}.Situacao(c.Request.Context()))
}

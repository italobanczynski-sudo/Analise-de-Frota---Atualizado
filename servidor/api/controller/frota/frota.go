package frota

import (
	"github.com/gin-gonic/gin"

	"analisefrota/api/controller"
	mwfrota "analisefrota/api/middleware/frota"
	"analisefrota/common/utils"
)

// Registros: GET /api/v1/clientes/:id/registros?inicio=AAAA-MM-DD&fim=AAAA-MM-DD
func Registros(c *gin.Context) {
	id, ok := controller.IDParam(c)
	if !ok {
		return
	}
	var req PeriodoRequest
	_ = c.ShouldBindQuery(&req)
	if err := utils.ValidateStruct(req); err != nil {
		controller.RetFail(c, "informe inicio e fim no formato AAAA-MM-DD")
		return
	}
	controller.Response(c, mwfrota.Frota_t{ClienteID: id, Inicio: req.Inicio, Fim: req.Fim}.Registros(c.Request.Context()))
}

// Odometros: GET /api/v1/clientes/:id/odometros
func Odometros(c *gin.Context) {
	id, ok := controller.IDParam(c)
	if !ok {
		return
	}
	controller.Response(c, mwfrota.Frota_t{ClienteID: id}.Odometros(c.Request.Context()))
}

// Premiacoes: GET /api/v1/clientes/:id/premiacoes?mes=AAAA-MM (sem mes: todos)
func Premiacoes(c *gin.Context) {
	id, ok := controller.IDParam(c)
	if !ok {
		return
	}
	var req MesRequest
	_ = c.ShouldBindQuery(&req)
	if err := utils.ValidateStruct(req); err != nil {
		controller.RetFail(c, "mes no formato AAAA-MM")
		return
	}
	controller.Response(c, mwfrota.Frota_t{ClienteID: id, Mes: req.Mes}.Premiacoes(c.Request.Context()))
}

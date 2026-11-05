package frota

import (
	"strings"

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

// Trajeto: GET /api/v1/clientes/:id/trajeto?placa=ABC1D23&dia=AAAA-MM-DD&de=0&ate=24
// Um ponto a cada 5 minutos (a API só dá a última posição de cada intervalo). Um pedido do dia inteiro pode
// voltar com completo=false: as janelas já buscadas ficam no banco e o próximo pedido continua de onde parou.
func Trajeto(c *gin.Context) {
	id, ok := controller.IDParam(c)
	if !ok {
		return
	}
	var req TrajetoRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		controller.RetFail(c, "informe placa, dia (AAAA-MM-DD) e, se quiser, de e ate em horas (0 a 24)")
		return
	}
	if err := utils.ValidateStruct(req); err != nil {
		controller.RetFail(c, "informe placa, dia (AAAA-MM-DD) e, se quiser, de e ate em horas (0 a 24)")
		return
	}
	controller.Response(c, mwfrota.Trajeto_t{ClienteID: id, Placa: req.Placa, Dia: req.Dia, De: req.De, Ate: req.Ate}.Buscar(c.Request.Context()))
}

// AoVivo: GET /api/v1/clientes/:id/ao-vivo?placas=ABC1D23,DEF4G56 (até 150)
// Onde cada placa está agora: em movimento, parada ou sem sinal, com a velocidade média entre as duas últimas
// posições. completo=false: o tempo do pedido acabou; peça de novo (as placas já vistas vêm do cache de 90 s).
func AoVivo(c *gin.Context) {
	id, ok := controller.IDParam(c)
	if !ok {
		return
	}
	var req AoVivoRequest
	if err := c.ShouldBindQuery(&req); err != nil || utils.ValidateStruct(req) != nil {
		controller.RetFail(c, "informe placas separadas por vírgula")
		return
	}
	controller.Response(c, mwfrota.AoVivo_t{ClienteID: id, Placas: strings.Split(req.Placas, ",")}.Buscar(c.Request.Context()))
}

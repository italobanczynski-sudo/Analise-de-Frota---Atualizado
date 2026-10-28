package mensagem

import (
	"github.com/gin-gonic/gin"

	"analisefrota/api/controller"
	mwmensagem "analisefrota/api/middleware/mensagem"
	"analisefrota/common/utils"
)

type DiaRequest struct {
	Dia string `form:"dia" validate:"omitempty,datetime=2006-01-02"`
}

type MarcaRequest struct {
	Dia       string `form:"dia" json:"dia" validate:"omitempty,datetime=2006-01-02"`
	Motorista string `form:"motorista" json:"motorista" validate:"required,max=120"`
}

type MotoristaRequest struct {
	Motorista string `form:"motorista" validate:"required,max=120"`
}

// Listar: GET /api/v1/clientes/:id/mensagens?dia=AAAA-MM-DD (sem dia: ontem)
func Listar(c *gin.Context) {
	id, ok := controller.IDParam(c)
	if !ok {
		return
	}
	var req DiaRequest
	_ = c.ShouldBindQuery(&req)
	if utils.ValidateStruct(req) != nil {
		controller.RetFail(c, "dia no formato AAAA-MM-DD")
		return
	}
	controller.Response(c, mwmensagem.Mensagem_t{ClienteID: id, Dia: req.Dia}.Listar(c.Request.Context()))
}

// Marcar: PUT /api/v1/clientes/:id/mensagens/enviada {"dia", "motorista"}
func Marcar(c *gin.Context) {
	id, ok := controller.IDParam(c)
	if !ok {
		return
	}
	var req MarcaRequest
	if err := c.ShouldBindJSON(&req); err != nil || utils.ValidateStruct(req) != nil {
		controller.RetFail(c, "envie o dia (AAAA-MM-DD) e o motorista")
		return
	}
	controller.Response(c, mwmensagem.Mensagem_t{ClienteID: id, UsuarioID: controller.UsuarioLogado(c).ID, Dia: req.Dia, Motorista: req.Motorista}.Marcar(c.Request.Context()))
}

// Desmarcar: DELETE /api/v1/clientes/:id/mensagens/enviada?dia=AAAA-MM-DD&motorista=Nome
func Desmarcar(c *gin.Context) {
	id, ok := controller.IDParam(c)
	if !ok {
		return
	}
	var req MarcaRequest
	_ = c.ShouldBindQuery(&req)
	if utils.ValidateStruct(req) != nil {
		controller.RetFail(c, "informe o dia (AAAA-MM-DD) e o motorista")
		return
	}
	controller.Response(c, mwmensagem.Mensagem_t{ClienteID: id, Dia: req.Dia, Motorista: req.Motorista}.Desmarcar(c.Request.Context()))
}

// Acompanhamento: GET /api/v1/clientes/:id/mensagens/acompanhamento?motorista=Nome
func Acompanhamento(c *gin.Context) {
	id, ok := controller.IDParam(c)
	if !ok {
		return
	}
	var req MotoristaRequest
	_ = c.ShouldBindQuery(&req)
	if utils.ValidateStruct(req) != nil {
		controller.RetFail(c, "informe o motorista")
		return
	}
	controller.Response(c, mwmensagem.Mensagem_t{ClienteID: id, Motorista: req.Motorista}.Acompanhamento(c.Request.Context()))
}

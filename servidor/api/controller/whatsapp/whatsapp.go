package whatsapp

import (
	"github.com/gin-gonic/gin"

	"analisefrota/api/controller"
	mwwhatsapp "analisefrota/api/middleware/whatsapp"
	"analisefrota/common/utils"
)

// Painel: GET /api/v1/clientes/:id/whatsapp
func Painel(c *gin.Context) {
	id, ok := controller.IDParam(c)
	if !ok {
		return
	}
	controller.Response(c, mwwhatsapp.WhatsApp_t{ClienteID: id}.Painel(c.Request.Context()))
}

// Previa: GET /api/v1/clientes/:id/whatsapp/previa?dia=AAAA-MM-DD (sem dia: ontem)
func Previa(c *gin.Context) {
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
	controller.Response(c, mwwhatsapp.WhatsApp_t{ClienteID: id, Dia: req.Dia}.Previa(c.Request.Context()))
}

// Enviar: POST /api/v1/clientes/:id/whatsapp/enviar {"dia"} (sem dia: ontem)
func Enviar(c *gin.Context) {
	id, ok := controller.IDParam(c)
	if !ok {
		return
	}
	var req DiaRequest
	if c.Request.ContentLength > 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			controller.RetFail(c, "corpo da requisição inválido: envie JSON com dia")
			return
		}
	}
	if utils.ValidateStruct(req) != nil {
		controller.RetFail(c, "dia no formato AAAA-MM-DD")
		return
	}
	controller.Response(c, mwwhatsapp.WhatsApp_t{ClienteID: id, Dia: req.Dia}.Enviar(c.Request.Context()))
}

// SalvarContato: PUT /api/v1/clientes/:id/whatsapp/contatos {"motorista", "telefone", "autorizado"}
func SalvarContato(c *gin.Context) {
	id, ok := controller.IDParam(c)
	if !ok {
		return
	}
	var req ContatoRequest
	if err := c.ShouldBindJSON(&req); err != nil || utils.ValidateStruct(req) != nil {
		controller.RetFail(c, "envie motorista, telefone e autorizado")
		return
	}
	controller.Response(c, mwwhatsapp.WhatsApp_t{ClienteID: id, UsuarioID: controller.UsuarioLogado(c).ID,
		Motorista: req.Motorista, Telefone: req.Telefone, Autorizado: req.Autorizado}.SalvarContato(c.Request.Context()))
}

// AtivarContato: PUT /api/v1/clientes/:id/whatsapp/contatos/ativo {"motorista", "ativo"}
func AtivarContato(c *gin.Context) {
	id, ok := controller.IDParam(c)
	if !ok {
		return
	}
	var req AtivoRequest
	if err := c.ShouldBindJSON(&req); err != nil || utils.ValidateStruct(req) != nil {
		controller.RetFail(c, "envie motorista e ativo")
		return
	}
	controller.Response(c, mwwhatsapp.WhatsApp_t{ClienteID: id, Motorista: req.Motorista, Ativo: req.Ativo}.AtivarContato(c.Request.Context()))
}

// RemoverContato: DELETE /api/v1/clientes/:id/whatsapp/contatos?motorista=Nome
func RemoverContato(c *gin.Context) {
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
	controller.Response(c, mwwhatsapp.WhatsApp_t{ClienteID: id, Motorista: req.Motorista}.RemoverContato(c.Request.Context()))
}

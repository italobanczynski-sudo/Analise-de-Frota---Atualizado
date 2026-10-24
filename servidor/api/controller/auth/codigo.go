package auth

import (
	"github.com/gin-gonic/gin"

	"analisefrota/api/controller"
	mwauth "analisefrota/api/middleware/auth"
	"analisefrota/common/utils"
)

// SolicitarCodigo: POST /api/v1/auth/codigo {"email"} (primeiro acesso ou senha esquecida)
func SolicitarCodigo(c *gin.Context) {
	var req CodigoRequest
	if err := c.ShouldBindJSON(&req); err != nil || utils.ValidateStruct(req) != nil {
		controller.RetFail(c, "informe um e-mail válido")
		return
	}
	controller.Response(c, mwauth.Auth_t{Email: req.Email, Origem: c.ClientIP()}.SolicitarCodigo(c.Request.Context()))
}

// ConfirmarCodigo: POST /api/v1/auth/codigo/confirmar {"email", "codigo", "senha"}
func ConfirmarCodigo(c *gin.Context) {
	var req ConfirmarCodigoRequest
	if err := c.ShouldBindJSON(&req); err != nil || utils.ValidateStruct(req) != nil {
		controller.RetFail(c, "informe o e-mail, o código de 6 dígitos e a senha nova")
		return
	}
	resp, ent := mwauth.Auth_t{Email: req.Email, Codigo: req.Codigo, Senha: req.Senha, Origem: c.ClientIP()}.ConfirmarCodigo(c.Request.Context())
	if ent != nil {
		controller.DefinirCookieSessao(c, ent.Token, ent.ExpiraEm)
	}
	controller.Response(c, resp)
}

package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"analisefrota/api/controller"
	mwauth "analisefrota/api/middleware/auth"
	"analisefrota/common/utils"
)

// Entrar: POST /api/v1/auth/entrar {"email", "senha"}
func Entrar(c *gin.Context) {
	var req EntrarRequest
	if err := c.ShouldBindJSON(&req); err != nil || utils.ValidateStruct(req) != nil {
		controller.RetFail(c, "informe e-mail e senha")
		return
	}
	resp, ent := mwauth.Auth_t{Email: req.Email, Senha: req.Senha, Origem: c.ClientIP()}.Entrar(c.Request.Context())
	if ent != nil {
		controller.DefinirCookieSessao(c, ent.Token, ent.ExpiraEm)
	}
	controller.Response(c, resp)
}

// AceitarConvite: POST /api/v1/auth/convite {"token", "senha"}
func AceitarConvite(c *gin.Context) {
	var req ConviteRequest
	if err := c.ShouldBindJSON(&req); err != nil || utils.ValidateStruct(req) != nil {
		controller.RetFail(c, "link de convite incompleto: abra o link inteiro que você recebeu")
		return
	}
	resp, ent := mwauth.Auth_t{Token: req.Token, Senha: req.Senha}.AceitarConvite(c.Request.Context())
	if ent != nil {
		controller.DefinirCookieSessao(c, ent.Token, ent.ExpiraEm)
	}
	controller.Response(c, resp)
}

// Sair: POST /api/v1/auth/sair
func Sair(c *gin.Context) {
	token, _ := c.Cookie(controller.CookieSessao)
	resp := mwauth.Auth_t{Token: token}.Sair(c.Request.Context())
	controller.LimparCookieSessao(c)
	controller.Response(c, resp)
}

// Eu: GET /api/v1/auth/eu (quem está logado)
func Eu(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"ok": true, "dados": controller.UsuarioLogado(c)})
}

package api

import (
	"errors"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"

	"analisefrota/api/controller"
	"analisefrota/api/middleware"
	mwauth "analisefrota/api/middleware/auth"
	"analisefrota/environment"
	"analisefrota/modules/usuario"
)

// autenticar exige um login válido (cookie de sessão) e guarda quem é no contexto
func autenticar() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, _ := c.Cookie(controller.CookieSessao)
		u, err := mwauth.Auth_t{Token: token}.Identificar(c.Request.Context())
		if err != nil {
			status := http.StatusUnauthorized
			if errors.Is(err, usuario.ErrInativo) {
				status = http.StatusForbidden
			}
			controller.LimparCookieSessao(c)
			c.AbortWithStatusJSON(status, gin.H{"ok": false, "erro": "entre na plataforma para continuar"})
			return
		}
		controller.GuardarUsuario(c, controller.UsuarioCtx_t{ID: u.ID, Nome: u.Nome, Email: u.Email, Papel: u.Papel})
		c.Next()
	}
}

// exigirPapel deixa passar só os perfis listados (vem depois de autenticar)
func exigirPapel(papeis ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !slices.Contains(papeis, controller.UsuarioLogado(c).Papel) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"ok": false, "erro": middleware.ErrSemPermissao.Error()})
			return
		}
		c.Next()
	}
}

// exigirCabecalhoApp: pedidos que mudam algo precisam do cabeçalho X-Requisicao, que só a própria plataforma manda.
// Junto com o cookie SameSite=Strict, impede que outro site faça pedidos em nome de quem está logado (CSRF).
func exigirCabecalhoApp() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead && c.GetHeader("X-Requisicao") != "plataforma" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"ok": false, "erro": "pedido recusado: use a própria plataforma"})
			return
		}
		c.Next()
	}
}

// cabecalhosSeguranca: a página só carrega o que vem do próprio servidor e não pode ser embutida em outro site
func cabecalhosSeguranca() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		// a única origem externa é a fonte Inter (Google Fonts), usada pelo CSS das telas
		h.Set("Content-Security-Policy", "default-src 'self'; img-src 'self' data:; style-src 'self' https://fonts.googleapis.com; font-src https://fonts.gstatic.com; script-src 'self'; connect-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("X-Frame-Options", "DENY")
		if environment.Cfg.Servidor.HTTPS() {
			h.Set("Strict-Transport-Security", "max-age=31536000")
		}
		c.Next()
	}
}

// cors libera o navegador só para as origens configuradas em ORIGENS_PERMITIDAS
func cors() gin.HandlerFunc {
	origens := environment.Cfg.Servidor.OrigensPermitidas
	return func(c *gin.Context) {
		if o := c.GetHeader("Origin"); o != "" && slices.Contains(origens, o) {
			c.Header("Access-Control-Allow-Origin", o)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Headers", "Content-Type, X-Requisicao")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

package api

import (
	"crypto/subtle"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"

	"analisefrota/environment"
)

// exigirChave protege a API até a fase de login: a requisição precisa do cabeçalho X-Chave-Admin
func exigirChave() gin.HandlerFunc {
	chave := []byte(environment.Cfg.Seguranca.ChaveAdmin)
	return func(c *gin.Context) {
		if subtle.ConstantTimeCompare([]byte(c.GetHeader("X-Chave-Admin")), chave) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"ok": false, "erro": "envie a chave de acesso no cabeçalho X-Chave-Admin"})
			return
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
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Headers", "Content-Type, X-Chave-Admin")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

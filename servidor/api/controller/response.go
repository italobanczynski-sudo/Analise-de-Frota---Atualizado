// Package controller tem os ajudantes de resposta HTTP comuns a todos os handlers.
package controller

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"analisefrota/api/middleware"
	"analisefrota/environment"
)

// RetFail responde 400 com a mensagem de validação da entrada
func RetFail(c *gin.Context, msg string) {
	c.JSON(http.StatusBadRequest, gin.H{"ok": false, "erro": msg})
}

// Response escreve a resposta do middleware: {"ok": true, "dados": ...} ou {"ok": false, "erro": ...}
func Response(c *gin.Context, r *middleware.Response_t) {
	if r.Erro != "" {
		c.JSON(r.Status, gin.H{"ok": false, "erro": r.Erro})
		return
	}
	c.JSON(r.Status, gin.H{"ok": true, "dados": r.Data})
}

// CookieSessao é o nome do cookie de login
const CookieSessao = "rg_sessao"

// chave do contexto Gin onde fica quem está logado (preenchida pela verificação de login em api/auth.go)
const chaveUsuario = "usuario_logado"

// UsuarioCtx_t é quem fez a requisição
type UsuarioCtx_t struct {
	ID    int64  `json:"id"`
	Nome  string `json:"nome"`
	Email string `json:"email"`
	Papel string `json:"papel"`
}

func GuardarUsuario(c *gin.Context, u UsuarioCtx_t) { c.Set(chaveUsuario, u) }

// UsuarioLogado devolve quem fez a requisição (as rotas protegidas sempre têm)
func UsuarioLogado(c *gin.Context) UsuarioCtx_t {
	u, _ := c.Get(chaveUsuario)
	v, _ := u.(UsuarioCtx_t)
	return v
}

// DefinirCookieSessao grava o login: HttpOnly (o JavaScript não lê), SameSite=Strict (outro site não usa), Secure em https
func DefinirCookieSessao(c *gin.Context, token string, expira time.Time) {
	http.SetCookie(c.Writer, &http.Cookie{Name: CookieSessao, Value: token, Path: "/", Expires: expira, MaxAge: int(time.Until(expira).Seconds()),
		HttpOnly: true, Secure: environment.Cfg.Servidor.HTTPS(), SameSite: http.SameSiteStrictMode})
}

func LimparCookieSessao(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{Name: CookieSessao, Value: "", Path: "/", MaxAge: -1,
		HttpOnly: true, Secure: environment.Cfg.Servidor.HTTPS(), SameSite: http.SameSiteStrictMode})
}

// IDParam lê o :id da rota; responde 400 e devolve false se não for um número positivo
func IDParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		RetFail(c, "id do cliente inválido")
		return 0, false
	}
	return id, true
}

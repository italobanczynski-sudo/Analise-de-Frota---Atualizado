// Package controller tem os ajudantes de resposta HTTP comuns a todos os handlers.
package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"analisefrota/api/middleware"
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

// IDParam lê o :id da rota; responde 400 e devolve false se não for um número positivo
func IDParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		RetFail(c, "id do cliente inválido")
		return 0, false
	}
	return id, true
}

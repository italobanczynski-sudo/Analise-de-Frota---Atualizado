// Package web entrega as telas da plataforma (HTML, CSS, JS e imagens embutidos no binário).
package web

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed paginas/*.html
var paginas embed.FS

//go:embed estatico
var estatico embed.FS

// Pagina devolve um handler que serve a tela pedida (sem cache: a versão nova aparece na hora)
func Pagina(nome string) gin.HandlerFunc {
	conteudo, err := paginas.ReadFile("paginas/" + nome)
	if err != nil {
		panic("web: página não encontrada: " + nome)
	}
	return func(c *gin.Context) {
		c.Header("Cache-Control", "no-store")
		c.Data(http.StatusOK, "text/html; charset=utf-8", conteudo)
	}
}

// Estatico serve CSS, JS e imagens
func Estatico() http.FileSystem {
	sub, _ := fs.Sub(estatico, "estatico")
	return http.FS(sub)
}

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

// Estatico serve CSS, JS e imagens (só arquivos: pasta dá 404, sem listar o conteúdo)
func Estatico() http.FileSystem {
	sub, _ := fs.Sub(estatico, "estatico")
	return semPastas{http.FS(sub)}
}

type semPastas struct{ http.FileSystem }

func (s semPastas) Open(nome string) (http.File, error) {
	f, err := s.FileSystem.Open(nome)
	if err != nil {
		return nil, err
	}
	if info, err := f.Stat(); err != nil || info.IsDir() {
		f.Close()
		return nil, fs.ErrNotExist
	}
	return f, nil
}

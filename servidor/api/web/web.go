// Package web entrega as telas da plataforma (HTML, CSS, JS e imagens embutidos no binário).
package web

import (
	"bytes"
	"embed"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed paginas/*.html
var paginas embed.FS

//go:embed estatico
var estatico embed.FS

// painel: cópia do index.html e do leitor de Excel da raiz do repositório, feita pelo Dockerfile
//
//go:embed painel
var painel embed.FS

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

// O painel tem scripts e estilos dentro do próprio HTML e busca o preço do diesel e a API Gobrax direto do navegador
const cspPainel = "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; " +
	"font-src 'self' data: https://fonts.gstatic.com; img-src 'self' data: blob: https:; " +
	"connect-src 'self' https://combustivelapi.com.br https://gateway-v3.gobrax.com.br:8889; frame-ancestors 'none'; base-uri 'none'; form-action 'self'"

// Painel serve o index.html (o mesmo do GitHub Pages) ligado à base do servidor por /estatico/painel-base.js
func Painel() gin.HandlerFunc {
	html, err := painel.ReadFile("painel/index.html")
	if err == nil {
		html = bytes.Replace(html, []byte(`src="vendor/xlsx.full.min.js"`), []byte(`src="/painel/xlsx.full.min.js"`), 1)
		// a ponte vem antes de qualquer script do painel: ele encontra a base em window.claude, como no Artifact
		html = bytes.Replace(html, []byte("<script"), []byte(`<script src="/estatico/painel-base.js"></script>`+"\n<script"), 1)
	}
	return func(c *gin.Context) {
		if len(html) == 0 {
			c.String(http.StatusServiceUnavailable, "Painel não incluído nesta versão do servidor (construa a imagem com o Dockerfile).")
			return
		}
		c.Header("Cache-Control", "no-store")
		c.Header("Content-Security-Policy", cspPainel)
		c.Data(http.StatusOK, "text/html; charset=utf-8", html)
	}
}

// PainelXLSX serve o leitor de Excel usado pelo painel para importar planilhas
func PainelXLSX() gin.HandlerFunc {
	js, _ := painel.ReadFile("painel/xlsx.full.min.js")
	return func(c *gin.Context) {
		if len(js) == 0 {
			c.Status(http.StatusNotFound)
			return
		}
		c.Header("Cache-Control", "public, max-age=86400")
		c.Data(http.StatusOK, "text/javascript; charset=utf-8", js)
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

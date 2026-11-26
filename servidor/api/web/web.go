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

// painel: cópia do painel-v2 montado (painel-v2/dist), feita pelo Dockerfile
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

// O painel tem scripts e estilos dentro do próprio HTML e busca o preço do diesel e a API Gobrax direto do navegador.
// O mapa do trajeto usa o Leaflet do cdnjs (script e estilo); os quadros do mapa (OpenStreetMap) já entram em img-src https:.
const cspPainel = "default-src 'self'; script-src 'self' 'unsafe-inline' https://cdnjs.cloudflare.com; " +
	"style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://cdnjs.cloudflare.com; " +
	"font-src 'self' data: https://fonts.gstatic.com; img-src 'self' data: blob: https:; " +
	"connect-src 'self' https://combustivelapi.com.br https://gateway-v3.gobrax.com.br:8889; frame-ancestors 'none'; base-uri 'none'; form-action 'self'"

// Painel serve o painel (painel-v2, identidade visual Gobrax). Ele mesmo lê os dados do servidor
// (/api/v1/painel/dados, ao vivo, mensagens do dia, enviadas e telefones).
func Painel() gin.HandlerFunc {
	html, _ := painel.ReadFile("painel/novo.html")
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

package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"

	"analisefrota/api/controller"
	"analisefrota/api/web"
	"analisefrota/environment"
	"analisefrota/modules/usuario"
)

// Testes dos filtros HTTP que não dependem do banco: cabeçalho da plataforma, perfil, CORS, segurança e proxy.
// O login completo (com banco) está em testes/e2e.sh.
type FiltrosSuite struct{ suite.Suite }

func (s *FiltrosSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	environment.Cfg = &environment.Config_t{Servidor: environment.ServidorConfig_t{
		URLPublica: "https://frota.exemplo.com", OrigensPermitidas: []string{"https://frota.exemplo.com"},
	}}
}

func (s *FiltrosSuite) pedir(r http.Handler, metodo, caminho string, cab map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(metodo, caminho, nil)
	for k, v := range cab {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func (s *FiltrosSuite) TestPedidoQueMudaAlgoPrecisaDoCabecalhoDaPlataforma() {
	r := gin.New()
	r.Use(exigirCabecalhoApp())
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.POST("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	s.Equal(http.StatusOK, s.pedir(r, http.MethodGet, "/x", nil).Code, "leitura não exige o cabeçalho")
	s.Equal(http.StatusForbidden, s.pedir(r, http.MethodPost, "/x", nil).Code, "outro site não manda o cabeçalho (CSRF)")
	s.Equal(http.StatusForbidden, s.pedir(r, http.MethodPost, "/x", map[string]string{"X-Requisicao": "outro"}).Code)
	s.Equal(http.StatusOK, s.pedir(r, http.MethodPost, "/x", map[string]string{"X-Requisicao": "plataforma"}).Code)
}

func (s *FiltrosSuite) TestPerfil() {
	casos := []struct {
		papel  string
		papeis []string
		status int
	}{
		{usuario.PapelAdmin, []string{usuario.PapelAdmin}, http.StatusOK},
		{usuario.PapelColaborador, []string{usuario.PapelAdmin}, http.StatusForbidden},
		{usuario.PapelLeitor, []string{usuario.PapelAdmin, usuario.PapelColaborador}, http.StatusForbidden},
		{usuario.PapelColaborador, []string{usuario.PapelAdmin, usuario.PapelColaborador}, http.StatusOK},
		{"", []string{usuario.PapelAdmin}, http.StatusForbidden}, // sem ninguém no contexto
	}
	for _, c := range casos {
		r := gin.New()
		papel := c.papel
		r.Use(func(ctx *gin.Context) {
			if papel != "" {
				controller.GuardarUsuario(ctx, controller.UsuarioCtx_t{ID: 1, Papel: papel})
			}
		}, exigirPapel(c.papeis...))
		r.GET("/x", func(ctx *gin.Context) { ctx.Status(http.StatusOK) })
		s.Equal(c.status, s.pedir(r, http.MethodGet, "/x", nil).Code, "%s em %v", c.papel, c.papeis)
	}
}

func (s *FiltrosSuite) TestCORSSoParaAOrigemConfigurada() {
	r := gin.New()
	r.Use(cors())
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := s.pedir(r, http.MethodGet, "/x", map[string]string{"Origin": "https://frota.exemplo.com"})
	s.Equal("https://frota.exemplo.com", w.Header().Get("Access-Control-Allow-Origin"))
	s.Equal("true", w.Header().Get("Access-Control-Allow-Credentials"))

	w = s.pedir(r, http.MethodGet, "/x", map[string]string{"Origin": "https://site-malicioso.exemplo"})
	s.Empty(w.Header().Get("Access-Control-Allow-Origin"))
}

func (s *FiltrosSuite) TestCabecalhosDeSeguranca() {
	r := gin.New()
	r.Use(cabecalhosSeguranca())
	r.GET("/x", func(c *gin.Context) { c.Status(http.StatusOK) })
	h := s.pedir(r, http.MethodGet, "/x", nil).Header()
	s.Contains(h.Get("Content-Security-Policy"), "frame-ancestors 'none'")
	s.Contains(h.Get("Content-Security-Policy"), "script-src 'self';")
	s.Equal("DENY", h.Get("X-Frame-Options"))
	s.Equal("nosniff", h.Get("X-Content-Type-Options"))
	s.NotEmpty(h.Get("Strict-Transport-Security"), "URL_PUBLICA em https liga o HSTS")
}

// Atrás do nginx, cada pessoa precisa ter o próprio IP; sem proxy configurado, o cabeçalho é ignorado.
func (s *FiltrosSuite) TestIPRealSoDeProxyConfiavel() {
	ipVisto := func(proxies []string) string {
		environment.Cfg.Servidor.ProxiesConfiaveis = proxies
		r := NewRouter()
		var ip string
		r.GET("/ip", func(c *gin.Context) { ip = c.ClientIP() })
		req := httptest.NewRequest(http.MethodGet, "/ip", nil)
		req.RemoteAddr = "172.18.0.5:41000" // o nginx na rede do Docker
		req.Header.Set("X-Forwarded-For", "203.0.113.7")
		r.ServeHTTP(httptest.NewRecorder(), req)
		return ip
	}
	s.Equal("203.0.113.7", ipVisto([]string{"172.16.0.0/12"}))
	s.Equal("172.18.0.5", ipVisto(nil), "sem proxy configurado, ninguém forja o IP pelo cabeçalho")
}

func (s *FiltrosSuite) TestEstaticoNaoListaPasta() {
	r := gin.New()
	r.StaticFS("/estatico", web.Estatico())
	s.Equal(http.StatusOK, s.pedir(r, http.MethodGet, "/estatico/comum.js", nil).Code)
	s.Equal(http.StatusNotFound, s.pedir(r, http.MethodGet, "/estatico/", nil).Code)
	s.Equal(http.StatusNotFound, s.pedir(r, http.MethodGet, "/estatico/nada.js", nil).Code)
}

func TestFiltrosSuite(t *testing.T) { suite.Run(t, new(FiltrosSuite)) }

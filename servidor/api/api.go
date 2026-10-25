// Package api sobe o servidor HTTP (Gin).
package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"analisefrota/environment"
)

type Service_t struct {
	porta string
}

func NewService() *Service_t {
	return &Service_t{porta: environment.Cfg.Servidor.Porta}
}

// NewRouter monta o Gin com os cabeçalhos de segurança, o CORS e todas as rotas (os testes de HTTP usam o mesmo)
func NewRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	// o IP só vem de X-Forwarded-For quando a conexão chega de um proxy configurado (ninguém forja o IP para fugir da trava)
	proxies := environment.Cfg.Servidor.ProxiesConfiaveis
	if len(proxies) == 0 {
		proxies = nil
	}
	if err := r.SetTrustedProxies(proxies); err != nil {
		logrus.Fatalf("PROXIES_CONFIAVEIS inválido: %v", err)
	}
	r.Use(gin.Recovery(), gin.LoggerWithConfig(gin.LoggerConfig{SkipPaths: []string{"/saude"}}), cabecalhosSeguranca(), cors())
	GetRoutes(r)
	return r
}

func (s *Service_t) Start() {
	if environment.Cfg.Servidor.HTTPS() {
		logrus.Infof("login só por HTTPS (URL_PUBLICA=%s): aberta por http://, a plataforma não consegue manter ninguém logado", environment.Cfg.Servidor.URLPublica)
	}
	r := NewRouter()
	srv := &http.Server{
		Addr:              ":" + s.porta,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
	}
	logrus.Infof("servidor ouvindo na porta %s", s.porta)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logrus.Fatalf("servidor parou: %v", err)
	}
}

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

func (s *Service_t) Start() {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery(), gin.LoggerWithConfig(gin.LoggerConfig{SkipPaths: []string{"/saude"}}), cors())
	GetRoutes(r)
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

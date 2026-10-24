package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	ctcliente "analisefrota/api/controller/cliente"
	ctcoleta "analisefrota/api/controller/coleta"
	ctfrota "analisefrota/api/controller/frota"
)

// GetRoutes registra todas as rotas do servidor
func GetRoutes(r *gin.Engine) {
	r.GET("/saude", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	v1 := r.Group("/api/v1", exigirChave())
	v1.GET("/clientes", ctcliente.Listar)
	v1.POST("/clientes", ctcliente.Criar)
	v1.PUT("/clientes/:id/token", ctcliente.SalvarToken)
	v1.POST("/clientes/:id/coleta", ctcoleta.Iniciar)
	v1.GET("/clientes/:id/coleta", ctcoleta.Situacao)
	v1.GET("/clientes/:id/registros", ctfrota.Registros)
	v1.GET("/clientes/:id/odometros", ctfrota.Odometros)
	v1.GET("/clientes/:id/premiacoes", ctfrota.Premiacoes)
}

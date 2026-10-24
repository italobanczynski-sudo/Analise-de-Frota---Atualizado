package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	ctauth "analisefrota/api/controller/auth"
	ctcliente "analisefrota/api/controller/cliente"
	ctcoleta "analisefrota/api/controller/coleta"
	ctfrota "analisefrota/api/controller/frota"
	ctusuario "analisefrota/api/controller/usuario"
	"analisefrota/api/web"
	"analisefrota/modules/usuario"
)

// GetRoutes registra todas as rotas do servidor
func GetRoutes(r *gin.Engine) {
	r.GET("/saude", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })

	// telas
	r.GET("/", web.Pagina("entrar.html"))
	r.GET("/convite", web.Pagina("convite.html"))
	r.GET("/plataforma", web.Pagina("plataforma.html"))
	r.StaticFS("/estatico", web.Estatico())

	// login (sem sessão)
	pub := r.Group("/api/v1/auth", exigirCabecalhoApp())
	pub.POST("/entrar", ctauth.Entrar)
	pub.POST("/convite", ctauth.AceitarConvite)
	pub.POST("/sair", ctauth.Sair)
	pub.POST("/codigo", ctauth.SolicitarCodigo)
	pub.POST("/codigo/confirmar", ctauth.ConfirmarCodigo)

	// qualquer pessoa logada
	v1 := r.Group("/api/v1", exigirCabecalhoApp(), autenticar())
	v1.GET("/auth/eu", ctauth.Eu)
	v1.GET("/clientes", ctcliente.Listar)
	v1.GET("/clientes/:id/coleta", ctcoleta.Situacao)
	v1.GET("/clientes/:id/registros", ctfrota.Registros)
	v1.GET("/clientes/:id/odometros", ctfrota.Odometros)
	v1.GET("/clientes/:id/premiacoes", ctfrota.Premiacoes)

	// colaborador e administrador
	colab := v1.Group("", exigirPapel(usuario.PapelAdmin, usuario.PapelColaborador))
	colab.POST("/clientes/:id/coleta", ctcoleta.Iniciar)

	// só administrador
	adm := v1.Group("", exigirPapel(usuario.PapelAdmin))
	adm.POST("/clientes", ctcliente.Criar)
	adm.PUT("/clientes/:id/token", ctcliente.SalvarToken)
	adm.GET("/usuarios", ctusuario.Listar)
	adm.POST("/usuarios", ctusuario.Criar)
	adm.POST("/usuarios/:id/convite", ctusuario.NovoConvite)
	adm.PUT("/usuarios/:id", ctusuario.Alterar)
}

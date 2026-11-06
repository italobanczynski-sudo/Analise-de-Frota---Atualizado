package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	ctauth "analisefrota/api/controller/auth"
	ctbase "analisefrota/api/controller/base"
	ctcliente "analisefrota/api/controller/cliente"
	ctcoleta "analisefrota/api/controller/coleta"
	ctfrota "analisefrota/api/controller/frota"
	ctmapa "analisefrota/api/controller/mapa"
	ctmensagem "analisefrota/api/controller/mensagem"
	ctusuario "analisefrota/api/controller/usuario"
	"analisefrota/api/web"
	"analisefrota/modules/usuario"
)

// GetRoutes registra todas as rotas do servidor
func GetRoutes(r *gin.Engine) {
	// público e sem credencial: o painel no GitHub Pages confere aqui se a plataforma está no ar antes de levar a pessoa para ela
	r.GET("/saude", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	// telas
	r.GET("/", web.Pagina("entrar.html"))
	r.GET("/convite", web.Pagina("convite.html"))
	r.GET("/plataforma", web.Pagina("plataforma.html"))
	r.StaticFS("/estatico", web.Estatico())
	// navegadores pedem /favicon.ico por conta própria: vai para o ícone da Gobrax
	r.GET("/favicon.ico", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "/estatico/icone.svg") })
	// painel completo (index.html): só com login; sem sessão volta para a tela de entrada
	r.GET("/painel", paginaLogada(), web.Painel())
	r.GET("/painel/xlsx.full.min.js", web.PainelXLSX())
	// fundo do mapa do painel, buscado pelo servidor no OpenStreetMap (só logado)
	r.GET("/mapa/:z/:x/:y", autenticar(), ctmapa.Imagem)

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
	v1.GET("/clientes/:id/trajeto", ctfrota.Trajeto)
	v1.GET("/clientes/:id/ao-vivo", ctfrota.AoVivo)
	v1.GET("/clientes/:id/mensagens", ctmensagem.Listar)
	v1.GET("/clientes/:id/mensagens/acompanhamento", ctmensagem.Acompanhamento)
	v1.GET("/clientes/:id/mensagens/enviadas", ctmensagem.Enviadas)
	v1.GET("/painel/dados", ctbase.Dados)
	v1.GET("/base/:colecao", ctbase.Listar)

	// colaborador e administrador
	colab := v1.Group("", exigirPapel(usuario.PapelAdmin, usuario.PapelColaborador))
	colab.POST("/clientes/:id/coleta", ctcoleta.Iniciar)
	// mensagens dos motoristas: quem manda pelo WhatsApp marca como enviada
	colab.PUT("/clientes/:id/mensagens/enviada", ctmensagem.Marcar)
	colab.DELETE("/clientes/:id/mensagens/enviada", ctmensagem.Desmarcar)
	// telefone do motorista para a fila de envio (o painel abre a conversa no WhatsApp; não há envio automático)
	colab.PUT("/clientes/:id/mensagens/telefone", ctmensagem.Telefone)
	// base do painel: o leitor só lê (como o Leitor do Artifact); colaborador e admin gravam
	colab.PUT("/base/:colecao/:doc", ctbase.Gravar)
	colab.POST("/base/:colecao", ctbase.Criar)
	colab.DELETE("/base/:colecao/:doc", ctbase.Apagar)

	// só administrador
	adm := v1.Group("", exigirPapel(usuario.PapelAdmin))
	adm.POST("/clientes", ctcliente.Criar)
	adm.PUT("/clientes/:id/token", ctcliente.SalvarToken)
	adm.GET("/usuarios", ctusuario.Listar)
	adm.POST("/usuarios", ctusuario.Criar)
	adm.POST("/usuarios/:id/convite", ctusuario.NovoConvite)
	adm.PUT("/usuarios/:id", ctusuario.Alterar)
}

package usuario

import (
	"github.com/gin-gonic/gin"

	"analisefrota/api/controller"
	mwusuario "analisefrota/api/middleware/usuario"
	"analisefrota/common/utils"
)

// Listar: GET /api/v1/usuarios
func Listar(c *gin.Context) {
	controller.Response(c, mwusuario.Usuario_t{}.Listar(c.Request.Context()))
}

// Criar: POST /api/v1/usuarios {"nome", "email", "papel"} → devolve o link de convite
func Criar(c *gin.Context) {
	var req CriarRequest
	if err := c.ShouldBindJSON(&req); err != nil || utils.ValidateStruct(req) != nil {
		controller.RetFail(c, "informe nome, um e-mail válido e o perfil (admin, colaborador ou leitor)")
		return
	}
	controller.Response(c, mwusuario.Usuario_t{Nome: req.Nome, Email: req.Email, Papel: req.Papel, QuemPede: controller.UsuarioLogado(c).ID}.Criar(c.Request.Context()))
}

// NovoConvite: POST /api/v1/usuarios/:id/convite → devolve um link novo (o anterior deixa de valer)
func NovoConvite(c *gin.Context) {
	id, ok := controller.IDParam(c)
	if !ok {
		return
	}
	controller.Response(c, mwusuario.Usuario_t{ID: id, QuemPede: controller.UsuarioLogado(c).ID}.NovoConvite(c.Request.Context()))
}

// Alterar: PUT /api/v1/usuarios/:id {"papel", "ativo"}
func Alterar(c *gin.Context) {
	id, ok := controller.IDParam(c)
	if !ok {
		return
	}
	var req AlterarRequest
	if err := c.ShouldBindJSON(&req); err != nil || utils.ValidateStruct(req) != nil {
		controller.RetFail(c, "informe o perfil (admin, colaborador ou leitor) e se a pessoa está ativa")
		return
	}
	controller.Response(c, mwusuario.Usuario_t{ID: id, Papel: req.Papel, Ativo: *req.Ativo, QuemPede: controller.UsuarioLogado(c).ID}.Alterar(c.Request.Context()))
}

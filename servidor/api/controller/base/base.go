// Package base recebe os pedidos do painel (index.html) para a base de documentos e os dados da API.
package base

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"analisefrota/api/controller"
	mwbase "analisefrota/api/middleware/base"
	modbase "analisefrota/modules/base"
)

// Listar: GET /api/v1/base/:colecao
func Listar(c *gin.Context) {
	controller.Response(c, mwbase.Base_t{Colecao: c.Param("colecao")}.Listar(c.Request.Context()))
}

// Gravar: PUT /api/v1/base/:colecao/:doc (corpo = o documento)
func Gravar(c *gin.Context) {
	dados, ok := corpo(c)
	if !ok {
		return
	}
	controller.Response(c, mwbase.Base_t{Colecao: c.Param("colecao"), ID: c.Param("doc"), Documento: dados, QuemPede: controller.UsuarioLogado(c).ID}.Gravar(c.Request.Context()))
}

// Criar: POST /api/v1/base/:colecao (corpo = o documento; o id é gerado)
func Criar(c *gin.Context) {
	dados, ok := corpo(c)
	if !ok {
		return
	}
	controller.Response(c, mwbase.Base_t{Colecao: c.Param("colecao"), Documento: dados, QuemPede: controller.UsuarioLogado(c).ID}.Criar(c.Request.Context()))
}

// Apagar: DELETE /api/v1/base/:colecao/:doc
func Apagar(c *gin.Context) {
	controller.Response(c, mwbase.Base_t{Colecao: c.Param("colecao"), ID: c.Param("doc")}.Apagar(c.Request.Context()))
}

// Dados: GET /api/v1/painel/dados (o que o servidor buscou na API Gobrax, para o painel)
func Dados(c *gin.Context) {
	controller.Response(c, mwbase.Base_t{}.Dados(c.Request.Context()))
}

// corpo lê o documento com limite de tamanho (o módulo confere se é um objeto JSON)
func corpo(c *gin.Context) ([]byte, bool) {
	dados, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, modbase.TamanhoMaximo+1))
	if err != nil {
		controller.RetFail(c, modbase.ErrGrande.Error())
		return nil, false
	}
	return dados, true
}

// Package middleware tem os casos de uso: orquestram os módulos e montam a resposta.
package middleware

import (
	"errors"
	"net/http"

	"github.com/sirupsen/logrus"

	"analisefrota/modules/base"
	"analisefrota/modules/cliente"
	"analisefrota/modules/coleta"
	"analisefrota/modules/diario"
	"analisefrota/modules/premiacao"
	"analisefrota/modules/sessao"
	"analisefrota/modules/usuario"
)

type Response_t struct {
	Status int
	Data   any
	Erro   string
}

func RetOkData(data any) *Response_t { return &Response_t{Status: http.StatusOK, Data: data} }

func RetCriado(data any) *Response_t { return &Response_t{Status: http.StatusCreated, Data: data} }

func RetAceito(data any) *Response_t { return &Response_t{Status: http.StatusAccepted, Data: data} }

func RetFail(status int, msg string) *Response_t { return &Response_t{Status: status, Erro: msg} }

// erros que o usuário pode corrigir: a mensagem vai como está
var (
	errosDeEntrada = []error{
		cliente.ErrNomeVazio, cliente.ErrNomeLongo, cliente.ErrHubspotID, cliente.ErrTokenVazio, cliente.ErrTokenEstranho, cliente.ErrSemToken,
		diario.ErrDiaInvalido, diario.ErrPeriodoInvertido, diario.ErrPeriodoLongo, premiacao.ErrMesInvalido, coleta.ErrDiasColeta,
		usuario.ErrNome, usuario.ErrEmail, usuario.ErrPapel, usuario.ErrSenhaFraca, usuario.ErrSenhaLonga, usuario.ErrConviteInvalido,
		usuario.ErrCodigoInvalido, base.ErrColecao, base.ErrID, base.ErrDados, base.ErrGrande,
	}
	errosNaoAutenticado = []error{usuario.ErrLogin, sessao.ErrSessaoInvalida, ErrMuitasTentativas}
	errosProibido       = []error{usuario.ErrInativo, ErrSemPermissao}
	errosNaoEncontrado  = []error{cliente.ErrNaoEncontrado, usuario.ErrNaoEncontrado}
	errosConflito       = []error{cliente.ErrNomeRepetido, coleta.ErrEmAndamento, usuario.ErrEmailRepetido, usuario.ErrUltimoAdmin}

	ErrSemPermissao     = errors.New("seu perfil não permite esta ação; fale com um administrador")
	ErrMuitasTentativas = errors.New("muitas tentativas de entrar; espere 15 minutos e tente de novo")
)

// RetErro traduz o erro de um módulo na resposta certa; erro inesperado vai para o log e não para o usuário
func RetErro(err error) *Response_t {
	for _, grupo := range []struct {
		lista  []error
		status int
	}{{errosDeEntrada, http.StatusBadRequest}, {errosNaoAutenticado, http.StatusUnauthorized}, {errosProibido, http.StatusForbidden},
		{errosNaoEncontrado, http.StatusNotFound}, {errosConflito, http.StatusConflict}} {
		for _, e := range grupo.lista {
			if errors.Is(err, e) {
				return RetFail(grupo.status, err.Error())
			}
		}
	}
	logrus.Errorf("erro inesperado: %v", err)
	return RetFail(http.StatusInternalServerError, "erro interno; veja o log do servidor")
}

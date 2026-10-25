package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"analisefrota/adapters/email"
	"analisefrota/api/middleware"
	"analisefrota/environment"
	"analisefrota/modules/usuario"
)

// resposta igual para e-mail cadastrado ou não: a tela não revela quem tem conta
const msgCodigoEnviado = "Se este e-mail estiver cadastrado, enviamos um código de 6 dígitos. Ele vale por 15 minutos."

// SolicitarCodigo manda por e-mail o código de primeiro acesso ou de troca de senha
func (m Auth_t) SolicitarCodigo(ctx context.Context) *middleware.Response_t {
	c, err := usuario.NewUsuario().SolicitarCodigo(ctx, m.Email)
	switch {
	case errors.Is(err, usuario.ErrAguarde):
		return middleware.RetFail(http.StatusTooManyRequests, err.Error())
	case errors.Is(err, usuario.ErrNaoEncontrado), errors.Is(err, usuario.ErrInativo):
		return middleware.RetOkData(map[string]string{"mensagem": msgCodigoEnviado})
	case err != nil:
		return middleware.RetErro(err)
	}
	if err := email.NewEmail().Enviar(c.Usuario.Email, "Seu código de acesso: "+c.Codigo, textoCodigo(c)); err != nil {
		logrus.Errorf("email: código para o usuário %d não saiu: %v", c.Usuario.ID, err)
		return middleware.RetFail(http.StatusServiceUnavailable, "não conseguimos enviar o e-mail agora; tente de novo em alguns minutos")
	}
	return middleware.RetOkData(map[string]string{"mensagem": msgCodigoEnviado})
}

// ConfirmarCodigo confere o código, grava a senha nova e já deixa a pessoa logada
func (m Auth_t) ConfirmarCodigo(ctx context.Context) (*middleware.Response_t, *Entrada_t) {
	agora := time.Now()
	if bloqueado(m.Email, m.Origem, agora) {
		return middleware.RetErro(middleware.ErrMuitasTentativas), nil
	}
	u, err := usuario.NewUsuario().UsarCodigo(ctx, m.Email, m.Codigo, m.Senha)
	if err != nil {
		if errors.Is(err, usuario.ErrCodigoInvalido) {
			registrarErro(m.Email, m.Origem, agora)
		}
		return middleware.RetErro(err), nil
	}
	limparErros(m.Email, m.Origem)
	return abrirSessaoComSenhaNova(ctx, u)
}

func textoCodigo(c usuario.Codigo_t) string {
	primeiro := strings.Fields(c.Usuario.Nome)
	nome := "Olá"
	if len(primeiro) > 0 {
		nome = "Olá, " + primeiro[0]
	}
	return fmt.Sprintf(`%s.

Seu código de acesso à Retenção Gobrax é:

    %s %s

Digite o código na tela de entrada e crie sua senha. Ele vale por 15 minutos e só pode ser usado uma vez.

Se você não pediu este código, ignore este e-mail: sua conta continua como estava.

%s
`, nome, c.Codigo[:3], c.Codigo[3:], strings.TrimRight(environment.Cfg.Servidor.URLPublica, "/"))
}

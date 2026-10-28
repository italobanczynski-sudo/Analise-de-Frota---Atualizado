// Package mensagem tem os casos de uso das mensagens prontas dos motoristas: listar, marcar como enviada e acompanhar.
package mensagem

import (
	"context"
	"time"

	"analisefrota/api/middleware"
	"analisefrota/common/utils"
	modcliente "analisefrota/modules/cliente"
	modmensagem "analisefrota/modules/mensagem"
)

// Mensagem_t reúne o que os casos de uso recebem do controller
type Mensagem_t struct {
	ClienteID int64
	UsuarioID int64
	Dia       string // AAAA-MM-DD; vazio = ontem
	Motorista string
}

func (m Mensagem_t) dia() string {
	if m.Dia != "" {
		return m.Dia
	}
	return utils.SomarDias(utils.DiaBR(time.Now()), -1)
}

// Listar devolve a mensagem pronta de cada motorista que rodou no dia e se já foi marcada como enviada
func (m Mensagem_t) Listar(ctx context.Context) *middleware.Response_t {
	if _, err := modcliente.NewCliente().Buscar(ctx, m.ClienteID); err != nil {
		return middleware.RetErro(err)
	}
	r, err := modmensagem.NewMensagem().Listar(ctx, m.ClienteID, m.dia())
	if err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetOkData(r)
}

func (m Mensagem_t) Marcar(ctx context.Context) *middleware.Response_t {
	if _, err := modcliente.NewCliente().Buscar(ctx, m.ClienteID); err != nil {
		return middleware.RetErro(err)
	}
	r, err := modmensagem.NewMensagem().Marcar(ctx, m.ClienteID, m.dia(), m.Motorista, m.UsuarioID)
	if err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetOkData(r)
}

func (m Mensagem_t) Desmarcar(ctx context.Context) *middleware.Response_t {
	if err := modmensagem.NewMensagem().Desmarcar(ctx, m.ClienteID, m.dia(), m.Motorista); err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetOkData(map[string]bool{"desmarcada": true})
}

// Acompanhamento devolve os dias com mensagem enviada ao motorista e se ele melhorou depois no ponto apontado
func (m Mensagem_t) Acompanhamento(ctx context.Context) *middleware.Response_t {
	if _, err := modcliente.NewCliente().Buscar(ctx, m.ClienteID); err != nil {
		return middleware.RetErro(err)
	}
	r, err := modmensagem.NewMensagem().Acompanhamento(ctx, m.ClienteID, m.Motorista)
	if err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetOkData(r)
}

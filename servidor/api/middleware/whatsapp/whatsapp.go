// Package whatsapp tem os casos de uso do WhatsApp dos motoristas: cadastro dos telefones, prévia e envio.
package whatsapp

import (
	"context"
	"errors"
	"net/http"
	"sort"
	"time"

	"github.com/sirupsen/logrus"

	"analisefrota/adapters/twilio"
	"analisefrota/api/middleware"
	"analisefrota/common/utils"
	modcliente "analisefrota/modules/cliente"
	modcontato "analisefrota/modules/contato"
	"analisefrota/modules/diario"
	modwhatsapp "analisefrota/modules/whatsapp"
)

// dias de histórico de envios na tela
const diasHistorico = 14

// Painel devolve a configuração, os telefones cadastrados, os motoristas recentes e os últimos envios
func (m WhatsApp_t) Painel(ctx context.Context) *middleware.Response_t {
	if _, err := modcliente.NewCliente().Buscar(ctx, m.ClienteID); err != nil {
		return middleware.RetErro(err)
	}
	contatos, err := modcontato.NewContato().Listar(ctx, m.ClienteID)
	if err != nil {
		return middleware.RetErro(err)
	}
	hoje := utils.DiaBR(time.Now())
	regs, err := diario.NewDiario().Listar(ctx, m.ClienteID, utils.SomarDias(hoje, -30), hoje)
	if err != nil {
		return middleware.RetErro(err)
	}
	vistos := map[string]bool{}
	motoristas := []string{}
	for _, r := range regs {
		if r.Motorista != "" && !vistos[r.Motorista] {
			vistos[r.Motorista] = true
			motoristas = append(motoristas, r.Motorista)
		}
	}
	sort.Strings(motoristas)
	svc := modwhatsapp.NewWhatsApp()
	envios, err := svc.Historico(ctx, m.ClienteID, diasHistorico)
	if err != nil {
		return middleware.RetErro(err)
	}
	cfg, err := svc.Config(ctx)
	if err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetOkData(Painel_t{Configurado: cfg.Configurado, Automatico: cfg.Automatico, Hora: cfg.Hora,
		Template: modwhatsapp.TextoTemplate, Contatos: contatos, Motoristas: motoristas, Envios: envios})
}

// Config devolve a conexão com a Twilio (sem o Auth Token)
func (m WhatsApp_t) Config(ctx context.Context) *middleware.Response_t {
	cfg, err := modwhatsapp.NewWhatsApp().Config(ctx)
	if err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetOkData(cfg)
}

// SalvarConfig grava a conexão com a Twilio (o Auth Token vai cifrado; vazio mantém o salvo)
func (m WhatsApp_t) SalvarConfig(ctx context.Context) *middleware.Response_t {
	cfg, err := modwhatsapp.NewWhatsApp().SalvarConfig(ctx, m.NovaConfig, m.UsuarioID)
	if err != nil {
		return middleware.RetErro(err)
	}
	logrus.Infof("whatsapp: conexão com a Twilio alterada pelo usuário %d", m.UsuarioID)
	return middleware.RetOkData(cfg)
}

// Previa monta as mensagens do dia sem enviar
func (m WhatsApp_t) Previa(ctx context.Context) *middleware.Response_t {
	if _, err := modcliente.NewCliente().Buscar(ctx, m.ClienteID); err != nil {
		return middleware.RetErro(err)
	}
	p, err := modwhatsapp.NewWhatsApp().Previa(ctx, m.ClienteID, m.dia())
	if err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetOkData(p)
}

// Enviar manda agora as mensagens do dia (quem já recebeu não recebe de novo)
func (m WhatsApp_t) Enviar(ctx context.Context) *middleware.Response_t {
	if _, err := modcliente.NewCliente().Buscar(ctx, m.ClienteID); err != nil {
		return middleware.RetErro(err)
	}
	r, err := modwhatsapp.NewWhatsApp().Enviar(ctx, m.ClienteID, m.dia())
	if errors.Is(err, twilio.ErrCredenciais) { // configuração do servidor: o administrador precisa ver a causa
		return middleware.RetFail(http.StatusBadGateway, err.Error())
	}
	if err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetOkData(r)
}

func (m WhatsApp_t) SalvarContato(ctx context.Context) *middleware.Response_t {
	if _, err := modcliente.NewCliente().Buscar(ctx, m.ClienteID); err != nil {
		return middleware.RetErro(err)
	}
	c, err := modcontato.NewContato().Salvar(ctx, m.ClienteID, m.Motorista, m.Telefone, m.Autorizado, m.UsuarioID)
	if err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetOkData(c)
}

func (m WhatsApp_t) AtivarContato(ctx context.Context) *middleware.Response_t {
	if err := modcontato.NewContato().Ativar(ctx, m.ClienteID, m.Motorista, m.Ativo); err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetOkData(map[string]bool{"ativo": m.Ativo})
}

// RemoverContato apaga o telefone (o motorista deixa de receber e o número sai do banco)
func (m WhatsApp_t) RemoverContato(ctx context.Context) *middleware.Response_t {
	if err := modcontato.NewContato().Remover(ctx, m.ClienteID, m.Motorista); err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetOkData(map[string]bool{"removido": true})
}

func (m WhatsApp_t) dia() string {
	if m.Dia != "" {
		return m.Dia
	}
	return utils.SomarDias(utils.DiaBR(time.Now()), -1)
}

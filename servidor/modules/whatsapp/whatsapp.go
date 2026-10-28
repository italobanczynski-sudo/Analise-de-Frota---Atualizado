// Package whatsapp manda a cada motorista, de manhã, a análise individual do dia anterior e o que melhorar.
// Lê o dia a dia (diario), a nota do mês (premiacao) e os telefones autorizados (contato); envia pelo template da Twilio.
// Cada motorista recebe no máximo uma mensagem por dia de resultado.
package whatsapp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"analisefrota/adapters/twilio"
	"analisefrota/common/utils"
	"analisefrota/environment"
	"analisefrota/modules/cliente"
	"analisefrota/modules/contato"
	"analisefrota/modules/diario"
	"analisefrota/modules/premiacao"
	"analisefrota/modules/whatsapp/whatsapp_repo"
)

type Service_I interface {
	// Previa monta as mensagens do dia sem enviar nada
	Previa(ctx context.Context, clienteID int64, dia string) ([]Previa_t, error)
	// Enviar manda as mensagens do dia para quem tem WhatsApp ativo e ainda não recebeu
	Enviar(ctx context.Context, clienteID int64, dia string) (Resultado_t, error)
	// EnviarTodos manda o resultado de ontem (dia de Brasília) para todos os clientes
	EnviarTodos(ctx context.Context) map[string]error
	// EnviarAgendado roda de hora em hora: só envia se o automático estiver ligado e for a hora marcada (Brasília)
	EnviarAgendado(ctx context.Context) map[string]error
	Historico(ctx context.Context, clienteID int64, dias int) ([]Envio_t, error)
	// EnviarTeste manda a mensagem de um motorista para um número qualquer e espera a Twilio dizer se entregou
	EnviarTeste(ctx context.Context, clienteID int64, dia, motorista, telefone string) (Teste_t, error)
	// Config e SalvarConfig: a conexão com a Twilio que o administrador preenche na aba WhatsApp
	Config(ctx context.Context) (Config_t, error)
	SalvarConfig(ctx context.Context, n NovaConfig_t, usuarioID int64) (Config_t, error)
}

// Deps_t são as dependências do envio; os testes trocam cada uma por um mock
type Deps_t struct {
	Repo      whatsapp_repo.Repository_I
	Clientes  cliente.Service_I
	Contatos  contato.Service_I
	Diario    diario.Service_I
	Premiacao premiacao.Service_I
	// Twilio monta o envio para a configuração que vale no momento
	Twilio func(cfg environment.WhatsAppConfig_t) twilio.WhatsApp_I
	Env    environment.WhatsAppConfig_t // usada enquanto o administrador não salvar a conexão na tela
	Chave  []byte                       // cifra o Auth Token salvo na tela
	Agora  func() time.Time
}

type whatsapp_t struct{ d Deps_t }

func NewWhatsApp() Service_I {
	return &whatsapp_t{d: Deps_t{
		Repo: whatsapp_repo.NewRepository(), Clientes: cliente.NewCliente(), Contatos: contato.NewContato(),
		Diario: diario.NewDiario(), Premiacao: premiacao.NewPremiacao(), Twilio: twilio.NewWhatsApp,
		Env: environment.Cfg.WhatsApp, Chave: environment.Cfg.Seguranca.ChaveTokensBytes(), Agora: time.Now,
	}}
}

// NewWhatsAppCom recebe as dependências (usado nos testes)
func NewWhatsAppCom(d Deps_t) Service_I { return &whatsapp_t{d: d} }

var ErrDiaFuturo = errors.New("o resultado só pode ser de um dia que já terminou (até ontem)")

func (w *whatsapp_t) resumos(ctx context.Context, clienteID int64, dia string) ([]Resumo_t, error) {
	if dia >= utils.DiaBR(w.d.Agora()) {
		return nil, ErrDiaFuturo
	}
	regs, err := w.d.Diario.Listar(ctx, clienteID, utils.SomarDias(dia, -DiasComparacao), dia)
	if err != nil {
		return nil, err
	}
	notas, err := w.notas(ctx, clienteID, dia)
	if err != nil {
		return nil, err
	}
	return MontarResumos(dia, regs, notas), nil
}

// notas: Nota Gobrax do mês do dia e do mês anterior, por motorista (premiação da API)
func (w *whatsapp_t) notas(ctx context.Context, clienteID int64, dia string) (map[string]Nota_t, error) {
	mes := dia[:7]
	atual, err := w.d.Premiacao.Listar(ctx, clienteID, mes)
	if err != nil {
		return nil, err
	}
	anterior, err := w.d.Premiacao.Listar(ctx, clienteID, MesAnterior(mes))
	if err != nil {
		return nil, err
	}
	parcial := mes == utils.DiaBR(w.d.Agora())[:7]
	out := map[string]Nota_t{}
	for _, p := range atual {
		out[ChaveMotorista(p.Motorista)] = Nota_t{Mes: mes, Atual: p.Nota, Parcial: parcial}
	}
	for _, p := range anterior {
		k := ChaveMotorista(p.Motorista)
		if n, ok := out[k]; ok {
			n.Anterior = p.Nota
			out[k] = n
		}
	}
	return out, nil
}

func (w *whatsapp_t) Previa(ctx context.Context, clienteID int64, dia string) ([]Previa_t, error) {
	res, err := w.resumos(ctx, clienteID, dia)
	if err != nil {
		return nil, err
	}
	dest, err := w.d.Contatos.Destinos(ctx, clienteID)
	if err != nil {
		return nil, err
	}
	tem := map[string]bool{}
	for _, d := range dest {
		tem[d.Motorista] = true
	}
	out := make([]Previa_t, len(res))
	for i, r := range res {
		out[i] = Previa_t{Resumo_t: r, Texto: Texto(r.Variaveis), TemWhatsApp: tem[r.Motorista]}
	}
	return out, nil
}

func (w *whatsapp_t) Enviar(ctx context.Context, clienteID int64, dia string) (Resultado_t, error) {
	cfg, err := w.efetiva(ctx)
	if err != nil {
		return Resultado_t{Dia: dia, Erros: []string{}}, err
	}
	envio := w.d.Twilio(cfg)
	out := Resultado_t{Dia: dia, Erros: []string{}, Teste: envio.Teste()}
	res, err := w.resumos(ctx, clienteID, dia)
	if err != nil {
		return out, err
	}
	dest, err := w.d.Contatos.Destinos(ctx, clienteID)
	if err != nil {
		return out, err
	}
	telefone := map[string]string{}
	for _, d := range dest {
		telefone[d.Motorista] = d.Telefone
	}
	for _, r := range res {
		tel, ok := telefone[r.Motorista]
		if !ok {
			out.SemWhatsApp++
			continue
		}
		livre, err := w.d.Repo.Reservar(ctx, clienteID, r.Motorista, dia)
		if err != nil {
			return out, err
		}
		if !livre {
			out.JaEnviadas++
			continue
		}
		sid, err := envio.Enviar(ctx, tel, r.Variaveis, Texto(r.Variaveis))
		if err != nil {
			_ = w.d.Repo.Concluir(ctx, clienteID, r.Motorista, dia, SituacaoErro, "", err.Error())
			out.Erros = append(out.Erros, r.Motorista+": "+err.Error())
			if errors.Is(err, twilio.ErrCredenciais) { // ninguém mais vai conseguir: para aqui
				return out, err
			}
			continue
		}
		if err := w.d.Repo.Concluir(ctx, clienteID, r.Motorista, dia, SituacaoEnviada, sid, ""); err != nil {
			return out, err
		}
		out.Enviadas++
	}
	return out, nil
}

func (w *whatsapp_t) EnviarAgendado(ctx context.Context) map[string]error {
	cfg, err := w.efetiva(ctx)
	if err != nil {
		logrus.Warnf("whatsapp: %v", err)
		return map[string]error{"(configuração)": err}
	}
	if !cfg.Automatico || w.d.Agora().In(utils.FusoBR()).Hour() != cfg.Hora {
		return nil
	}
	return w.EnviarTodos(ctx)
}

func (w *whatsapp_t) EnviarTodos(ctx context.Context) map[string]error {
	erros := map[string]error{}
	clientes, err := w.d.Clientes.Listar(ctx)
	if err != nil {
		erros["(lista de clientes)"] = err
		return erros
	}
	ontem := utils.SomarDias(utils.DiaBR(w.d.Agora()), -1)
	for _, c := range clientes {
		r, err := w.Enviar(ctx, c.ID, ontem)
		if err != nil {
			erros[c.Nome] = err
			logrus.Warnf("whatsapp: %s: %v", c.Nome, err)
			if errors.Is(err, twilio.ErrCredenciais) {
				return erros
			}
			continue
		}
		if r.Enviadas > 0 || len(r.Erros) > 0 {
			logrus.Infof("whatsapp: %s, dia %s: %d enviada(s), %d com erro, %d sem WhatsApp", c.Nome, ontem, r.Enviadas, len(r.Erros), r.SemWhatsApp)
		}
		if len(r.Erros) > 0 {
			erros[c.Nome] = fmt.Errorf("%d mensagem(ns) com erro", len(r.Erros))
		}
	}
	return erros
}

func (w *whatsapp_t) Historico(ctx context.Context, clienteID int64, dias int) ([]Envio_t, error) {
	desde := utils.SomarDias(utils.DiaBR(w.d.Agora()), -dias)
	rows, err := w.d.Repo.Listar(ctx, clienteID, desde)
	if err != nil {
		return nil, err
	}
	out := make([]Envio_t, len(rows))
	for i, r := range rows {
		out[i] = Envio_t{Motorista: r.Motorista, Dia: r.Dia, Situacao: r.Situacao, Erro: r.Erro, CriadoEm: r.CriadoEm}
	}
	return out, nil
}

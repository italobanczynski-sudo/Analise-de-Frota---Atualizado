// Package coleta busca na API Gobrax os dados de um cliente e entrega cada parte ao módulo dono:
// dia a dia (diario), odômetro (odometro) e premiação (premiacao). Guarda a situação da última busca.
package coleta

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/sirupsen/logrus"

	"analisefrota/adapters/gobraxapi"
	"analisefrota/common/utils"
	"analisefrota/environment"
	"analisefrota/modules/cliente"
	"analisefrota/modules/coleta/coleta_repo"
	"analisefrota/modules/diario"
	"analisefrota/modules/odometro"
	"analisefrota/modules/premiacao"
)

type Service_I interface {
	// Executar busca um cliente: da última busca até hoje (na primeira vez, os últimos N dias)
	Executar(ctx context.Context, clienteID int64) (Resultado_t, error)
	// ExecutarTodos busca, um de cada vez, todos os clientes que têm token
	ExecutarTodos(ctx context.Context) map[string]error
	Situacao(ctx context.Context, clienteID int64) (Situacao_t, error)
	// MarcarInterrompidas: ao subir, nenhuma busca está rodando de verdade; as que ficaram "rodando" foram interrompidas
	MarcarInterrompidas(ctx context.Context) (int64, error)
}

// Deps_t são as dependências da coleta; os testes trocam cada uma por um mock
type Deps_t struct {
	Repo      coleta_repo.Repository_I
	API       gobraxapi.Client_I
	Clientes  cliente.Service_I
	Diario    diario.Service_I
	Odometro  odometro.Service_I
	Premiacao premiacao.Service_I
	Agora     func() time.Time
	Dias      int
}

type coleta_t struct{ d Deps_t }

func NewColeta() Service_I {
	return &coleta_t{d: Deps_t{
		Repo: coleta_repo.NewRepository(), API: gobraxapi.NewClient(), Clientes: cliente.NewCliente(),
		Diario: diario.NewDiario(), Odometro: odometro.NewOdometro(), Premiacao: premiacao.NewPremiacao(),
		Agora: time.Now, Dias: environment.Cfg.Coleta.Dias,
	}}
}

// NewColetaCom recebe as dependências (usado nos testes)
func NewColetaCom(d Deps_t) Service_I { return &coleta_t{d: d} }

func (c *coleta_t) Executar(ctx context.Context, clienteID int64) (Resultado_t, error) {
	if err := validarDias(c.d.Dias); err != nil {
		return Resultado_t{}, err
	}
	if !trancar(clienteID) {
		return Resultado_t{}, ErrEmAndamento
	}
	defer liberar(clienteID)

	agora := c.d.Agora()
	sit, err := c.d.Repo.Buscar(ctx, clienteID)
	if err != nil {
		return Resultado_t{}, err
	}
	hoje := utils.DiaBR(agora)
	ini := utils.SomarDias(hoje, -(c.d.Dias - 1))
	if sit.UltimoDia.Valid && sit.UltimoDia.String > ini {
		ini = utils.SomarDias(sit.UltimoDia.String, -1) // refaz o último dia buscado, que pode ter vindo parcial
	}
	if err := c.d.Repo.MarcarInicio(ctx, clienteID, agora); err != nil {
		return Resultado_t{}, err
	}

	res, ultimoDia, err := c.buscar(ctx, clienteID, ini, hoje, agora)
	fim := c.d.Agora()
	if err != nil {
		_ = c.d.Repo.MarcarFim(ctx, clienteID, SituacaoErro, err.Error(), res.Avisos, nil, fim)
		return res, err
	}
	return res, c.d.Repo.MarcarFim(ctx, clienteID, SituacaoOk, resumo(res), res.Avisos, &ultimoDia, fim)
}

// buscar devolve o resultado e até que dia a busca ficou completa
func (c *coleta_t) buscar(ctx context.Context, clienteID int64, ini, hoje string, agora time.Time) (Resultado_t, string, error) {
	res := Resultado_t{Performance: true, Avisos: []string{}}
	token, err := c.d.Clientes.Token(ctx, clienteID)
	if err != nil {
		return res, "", err
	}

	fe, fp := &falhas_t{nome: "estatísticas (km e litros)"}, &falhas_t{nome: "performance"}
	for _, dia := range utils.DiasEntre(ini, hoje) {
		jIni, jFim, _ := utils.JanelaDiaBR(dia)
		var estat []gobraxapi.EstatisticaRec_t
		estatFalhou, estatPediu := false, !fe.desligado
		if estatPediu {
			estat, err = c.d.API.Estatisticas(ctx, token, jIni, jFim)
			switch {
			case err == nil || errors.Is(err, gobraxapi.ErrSemDados):
				fe.acertou()
			case errors.Is(err, gobraxapi.ErrTokenInvalido):
				return res, "", err
			default:
				fe.errou(dia, err)
				estatFalhou = true
			}
		}
		var perf []gobraxapi.PerformanceRec_t
		perfFalhou, perfPediu := false, !fp.desligado
		if perfPediu {
			perf, err = c.d.API.Performance(ctx, token, jIni, jFim)
			switch {
			case err == nil || errors.Is(err, gobraxapi.ErrSemDados):
				fp.acertou()
			case errors.Is(err, gobraxapi.ErrTokenInvalido):
				return res, "", err
			default:
				fp.errou(dia, err)
				perf, perfFalhou = nil, true
			}
		}
		// limite de uso: para aqui; o que já veio fica salvo e este dia em diante volta na próxima busca
		if fe.limite || fp.limite {
			break
		}
		if fe.desligado && fp.desligado && !fp.semAcesso {
			res.Avisos = append(res.Avisos, append(fe.avisos(), fp.avisos()...)...)
			return res, "", errors.New("estatísticas e performance pararam de responder; tente de novo mais tarde")
		}
		// dado que faltou por falha ou limite: não troca o dia salvo por um dia pela metade (a próxima busca tenta de novo)
		if estatFalhou || !estatPediu && fe.faltou(dia, false) {
			continue
		}
		if (perfFalhou || !perfPediu) && fp.faltou(dia, perfPediu) {
			continue
		}
		regs := c.d.Diario.MontarDia(clienteID, dia, estat, perf)
		if len(regs) == 0 {
			continue
		}
		if err := c.d.Diario.SalvarDia(ctx, clienteID, dia, regs); err != nil {
			return res, "", err
		}
		res.DiasComDados++
		if len(regs) > res.Veiculos {
			res.Veiculos = len(regs)
		}
	}
	res.Performance = !fp.desligado
	res.Avisos = append(res.Avisos, append(fe.avisos(), fp.avisos()...)...)

	// odômetro: o último de cada veículo nos últimos 30 dias
	if odo, err := c.d.API.Odometro(ctx, token, agora.Add(-30*24*time.Hour), agora); err == nil {
		if res.Odometros, err = c.d.Odometro.Substituir(ctx, clienteID, odo); err != nil {
			return res, "", err
		}
	} else if errors.Is(err, gobraxapi.ErrTokenInvalido) {
		return res, "", err
	} else {
		res.Avisos = append(res.Avisos, "odômetro não veio: "+err.Error())
	}

	// premiação: do mês do início ao mês atual
	if prem, err := c.d.API.Premiacao(ctx, token, utils.MesAPI(ini), utils.MesAPI(hoje)); err == nil {
		if res.MesesPremio, err = c.d.Premiacao.Salvar(ctx, clienteID, prem); err != nil {
			return res, "", err
		}
	} else {
		res.Avisos = append(res.Avisos, "premiação não veio: "+err.Error())
	}

	// dias perdidos voltam na próxima busca: o "último dia completo" fica antes do primeiro perdido
	ultimo := hoje
	perdidos := append(append([]string{}, fe.perdidos...), fp.perdidos...)
	if len(perdidos) > 0 {
		sort.Strings(perdidos)
		ultimo = utils.SomarDias(perdidos[0], -1)
	}
	return res, ultimo, nil
}

func (c *coleta_t) ExecutarTodos(ctx context.Context) map[string]error {
	out := map[string]error{}
	clis, err := c.d.Clientes.Listar(ctx)
	if err != nil {
		out["(lista de clientes)"] = err
		return out
	}
	for _, cl := range clis {
		if !cl.TemToken {
			continue
		}
		res, err := c.Executar(ctx, cl.ID)
		out[cl.Nome] = err
		if err != nil {
			logrus.Warnf("coleta: %s: %v", cl.Nome, err)
			continue
		}
		logrus.Infof("coleta: %s: %s", cl.Nome, resumo(res))
		for _, a := range res.Avisos {
			logrus.Warnf("coleta: %s: %s", cl.Nome, a)
		}
	}
	return out
}

func (c *coleta_t) MarcarInterrompidas(ctx context.Context) (int64, error) {
	return c.d.Repo.MarcarInterrompidas(ctx, c.d.Agora())
}

func (c *coleta_t) Situacao(ctx context.Context, clienteID int64) (Situacao_t, error) {
	r, err := c.d.Repo.Buscar(ctx, clienteID)
	if err != nil {
		return Situacao_t{}, err
	}
	s := Situacao_t{ClienteID: clienteID, IniciadaEm: r.IniciadaEm, TerminadaEm: r.TerminadaEm, Situacao: r.Situacao, Mensagem: r.Mensagem, Avisos: r.Avisos}
	if r.UltimoDia.Valid {
		s.UltimoDia = r.UltimoDia.String
	}
	if s.Situacao == "" {
		s.Situacao = SituacaoNunca
	}
	if s.Avisos == nil {
		s.Avisos = []string{}
	}
	return s, nil
}

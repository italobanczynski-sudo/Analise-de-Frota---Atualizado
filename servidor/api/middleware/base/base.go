// Package base tem os casos de uso do painel (index.html): a base de documentos e os dados da API Gobrax.
package base

import (
	"context"
	"time"

	"analisefrota/api/middleware"
	"analisefrota/common/utils"
	modbase "analisefrota/modules/base"
	modcliente "analisefrota/modules/cliente"
	"analisefrota/modules/diario"
	"analisefrota/modules/odometro"
	"analisefrota/modules/premiacao"
)

// Base_t reúne o que os casos de uso da base recebem do controller
type Base_t struct {
	Colecao   string
	ID        string
	Documento []byte
	QuemPede  int64
}

func (m Base_t) Listar(ctx context.Context) *middleware.Response_t {
	docs, err := modbase.NewBase().Listar(ctx, m.Colecao)
	if err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetOkData(docs)
}

func (m Base_t) Gravar(ctx context.Context) *middleware.Response_t {
	if err := modbase.NewBase().Gravar(ctx, m.Colecao, m.ID, m.Documento, m.QuemPede); err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetOkData(map[string]string{"id": m.ID})
}

func (m Base_t) Criar(ctx context.Context) *middleware.Response_t {
	id, err := modbase.NewBase().Criar(ctx, m.Colecao, m.Documento, m.QuemPede)
	if err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetCriado(map[string]string{"id": id})
}

func (m Base_t) Apagar(ctx context.Context) *middleware.Response_t {
	if err := modbase.NewBase().Apagar(ctx, m.Colecao, m.ID); err != nil {
		return middleware.RetErro(err)
	}
	return middleware.RetOkData(map[string]bool{"apagado": true})
}

// DiasPainel é quanto o painel recebe de dia a dia de cada cliente (o limite de consulta do módulo diario é 400)
const DiasPainel = 365

// DadosCliente_t é o que o servidor buscou na API Gobrax para um cliente, no formato que o painel monta em planilhas
type DadosCliente_t struct {
	ID         int64                   `json:"id"`
	Nome       string                  `json:"nome"`
	HubspotID  string                  `json:"hubspot_id"`
	Registros  []diario.Registro_t     `json:"registros"`
	Odometros  []odometro.Odometro_t   `json:"odometros"`
	Premiacoes []premiacao.Premiacao_t `json:"premiacoes"`
}

// Dados devolve, de cada cliente, o dia a dia do último ano, o odômetro e a premiação
func (m Base_t) Dados(ctx context.Context) *middleware.Response_t {
	clis, err := modcliente.NewCliente().Listar(ctx)
	if err != nil {
		return middleware.RetErro(err)
	}
	hoje := utils.DiaBR(time.Now())
	ini := utils.SomarDias(hoje, -(DiasPainel - 1))
	dia, odo, prem := diario.NewDiario(), odometro.NewOdometro(), premiacao.NewPremiacao()
	out := make([]DadosCliente_t, 0, len(clis))
	for _, c := range clis {
		d := DadosCliente_t{ID: c.ID, Nome: c.Nome, HubspotID: c.HubspotID}
		if d.Registros, err = dia.Listar(ctx, c.ID, ini, hoje); err != nil {
			return middleware.RetErro(err)
		}
		if d.Odometros, err = odo.Listar(ctx, c.ID); err != nil {
			return middleware.RetErro(err)
		}
		if d.Premiacoes, err = prem.Listar(ctx, c.ID, ""); err != nil {
			return middleware.RetErro(err)
		}
		out = append(out, d)
	}
	return middleware.RetOkData(out)
}

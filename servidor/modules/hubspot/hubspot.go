// Package hubspot monta a leitura de um cliente no HubSpot (perfil, saúde CHS, ligações, e-mails, reuniões,
// notas, chamados e contatos) para o cartão "Cliente no HubSpot" do painel. Só leitura; dados pessoais saem mascarados.
// Cada leitura vale alguns minutos em memória, para o painel poder atualizar sozinho sem estourar o limite do HubSpot.
package hubspot

import (
	"context"
	"errors"
	"sync"
	"time"

	"analisefrota/adapters/hubspotapi"
	"analisefrota/environment"
)

var (
	ErrDesligado     = errors.New("o HubSpot não está configurado no servidor (falta HUBSPOT_TOKEN)")
	ErrNaoEncontrado = errors.New("este cliente não foi encontrado no HubSpot pelo nome; preencha o ID do HubSpot no cadastro do cliente")
)

type Service_I interface {
	// Resumo lê o cliente no HubSpot. Sem hubspotID, procura a empresa pelo nome. atualizar ignora a leitura guardada.
	Resumo(ctx context.Context, nome, hubspotID string, atualizar bool) (Resumo_t, error)
}

type guardado_t struct {
	r  Resumo_t
	em time.Time
}

type hubspot_t struct {
	api      hubspotapi.Client_I
	ligado   bool
	validade time.Duration
	agora    func() time.Time

	mu    sync.Mutex // uma leitura por vez: evita duas buscas iguais e respeita o limite do HubSpot
	cache map[string]guardado_t
}

var (
	localSvc Service_I
	svcOnce  sync.Once
)

// NewHubspot devolve o serviço único (a leitura guardada vale para todos os usuários)
func NewHubspot() Service_I {
	svcOnce.Do(func() {
		cfg := environment.Cfg.Hubspot
		var api hubspotapi.Client_I
		if cfg.Configurado() {
			api = hubspotapi.NewClient()
		}
		localSvc = NewHubspotCom(api, time.Duration(cfg.CacheMin)*time.Minute, time.Now)
	})
	return localSvc
}

// NewHubspotCom recebe o cliente da API, a validade da leitura e o relógio (usado nos testes). api nil = desligado.
func NewHubspotCom(api hubspotapi.Client_I, validade time.Duration, agora func() time.Time) Service_I {
	return &hubspot_t{api: api, ligado: api != nil, validade: validade, agora: agora, cache: map[string]guardado_t{}}
}

func (h *hubspot_t) Resumo(ctx context.Context, nome, hubspotID string, atualizar bool) (Resumo_t, error) {
	if !h.ligado {
		return Resumo_t{}, ErrDesligado
	}
	chave := chaveNome(nome) + "|" + hubspotID
	h.mu.Lock()
	defer h.mu.Unlock()
	if g, ok := h.cache[chave]; ok && !atualizar && h.agora().Sub(g.em) < h.validade {
		return g.r, nil
	}
	r, err := h.ler(ctx, nome, hubspotID)
	if err != nil {
		return Resumo_t{}, err
	}
	h.cache[chave] = guardado_t{r: r, em: h.agora()}
	return r, nil
}

func (h *hubspot_t) empresa(ctx context.Context, nome, hubspotID string) (string, error) {
	if hubspotID != "" {
		return hubspotID, nil
	}
	lista, err := h.api.BuscarEmpresas(ctx, textoBusca(nome))
	if err != nil {
		return "", err
	}
	if len(lista) == 0 {
		return "", ErrNaoEncontrado
	}
	k := chaveNome(nome)
	for _, o := range lista {
		if chaveNome(o.Properties["name"]) == k {
			return o.ID, nil
		}
	}
	return lista[0].ID, nil
}

func (h *hubspot_t) ler(ctx context.Context, nome, hubspotID string) (Resumo_t, error) {
	id, err := h.empresa(ctx, nome, hubspotID)
	if err != nil {
		return Resumo_t{}, err
	}
	comp, err := h.api.Objeto(ctx, "companies", id, props["companies"])
	if err != nil {
		return Resumo_t{}, err
	}
	tipos := []string{"calls", "emails", "meetings", "notes", "tickets", "contacts"}
	ids, reg := map[string][]string{}, map[string][]hubspotapi.Objeto_t{}
	semDetalhe := []string{}
	for _, t := range tipos {
		ids[t], err = h.api.Associados(ctx, "companies", id, t)
		if errors.Is(err, hubspotapi.ErrSemEscopo) {
			ids[t] = []string{}
		} else if err != nil {
			return Resumo_t{}, err
		}
		// sem o escopo de leitura do tipo, o total (das associações) continua valendo; só o detalhe fica de fora
		reg[t], err = h.api.LerLote(ctx, t, ids[t], props[t])
		if errors.Is(err, hubspotapi.ErrSemEscopo) {
			semDetalhe = append(semDetalhe, t)
			reg[t] = []hubspotapi.Objeto_t{}
		} else if err != nil {
			return Resumo_t{}, err
		}
		limpar(reg[t])
	}
	for _, t := range []string{"calls", "emails", "meetings", "notes"} {
		maisRecentes(reg[t], "hs_timestamp")
	}
	maisRecentes(reg["tickets"], "createdate")
	maisRecentes(reg["contacts"], "notes_last_contacted")

	semDet := func(t string) bool {
		for _, x := range semDetalhe {
			if x == t {
				return true
			}
		}
		return false
	}
	bloco := func(t string) Bloco_t {
		total := len(ids[t])
		if len(reg[t]) > total {
			total = len(reg[t])
		}
		return Bloco_t{Total: total, Results: primeiros(reg[t], maxRegistros), SemDetalhe: semDet(t)}
	}
	var abertos *Bloco_t
	if !semDet("tickets") {
		ab := []hubspotapi.Objeto_t{}
		for _, o := range reg["tickets"] {
			if o.Properties["closed_date"] == "" {
				ab = append(ab, o)
			}
		}
		abertos = &Bloco_t{Total: len(ab), Results: primeiros(ab, maxRegistros)}
	}
	// quem mais fala com a Gobrax: ligações e e-mails de cada contato (até 10)
	porContato := []PorContato_t{}
	for _, o := range primeiros(reg["contacts"], maxContatos) {
		l, _ := h.api.Associados(ctx, "contacts", o.ID, "calls")
		e, _ := h.api.Associados(ctx, "contacts", o.ID, "emails")
		porContato = append(porContato, PorContato_t{ID: o.ID, L: len(l), E: len(e)})
	}
	compLimpo := limpar([]hubspotapi.Objeto_t{comp})[0].Properties
	return Resumo_t{
		Formato: formato, Gerado: h.agora().UTC().Format(time.RFC3339), Cliente: nome, ID: id, Comp: compLimpo,
		Calls: bloco("calls"), Emails: bloco("emails"), Meets: bloco("meetings"), Notas: bloco("notes"),
		Tickets: bloco("tickets"), Abertos: abertos, Contatos: bloco("contacts"), PorContato: porContato, SemDetalhe: semDetalhe,
	}, nil
}

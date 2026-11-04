package hubspot

import "analisefrota/adapters/hubspotapi"

// Resumo_t é a leitura do cliente no HubSpot, no mesmo formato do arquivo do script busca-de-dados/hubspot.mjs
// ("gobrax-hubspot-1"), que o painel já sabe mostrar
type Resumo_t struct {
	Formato    string            `json:"formato"`
	Gerado     string            `json:"gerado"`
	Cliente    string            `json:"cliente"`
	ID         string            `json:"id"`
	Comp       map[string]string `json:"comp"`
	Calls      Bloco_t           `json:"calls"`
	Emails     Bloco_t           `json:"emails"`
	Meets      Bloco_t           `json:"meets"`
	Notas      Bloco_t           `json:"notas"`
	Tickets    Bloco_t           `json:"tickets"`
	Abertos    *Bloco_t          `json:"abertos"` // null quando o token não lê chamados
	Contatos   Bloco_t           `json:"contatos"`
	PorContato []PorContato_t    `json:"porContato"`
	SemDetalhe []string          `json:"semDetalhe"`
}

// Bloco_t: total completo e até maxRegistros registros mais recentes
type Bloco_t struct {
	Total      int                   `json:"total"`
	Results    []hubspotapi.Objeto_t `json:"results"`
	SemDetalhe bool                  `json:"semDetalhe,omitempty"`
}

// PorContato_t: quantas ligações (l) e e-mails (e) cada contato teve com a Gobrax
type PorContato_t struct {
	ID string `json:"id"`
	L  int    `json:"l"`
	E  int    `json:"e"`
}

const (
	formato      = "gobrax-hubspot-1"
	maxRegistros = 200 // registros guardados por tipo (os totais são sempre completos)
	maxContatos  = 10
)

// Mesmos campos que o painel mostra (index.html, cartão Cliente no HubSpot). CNPJ, e-mail e telefone não são pedidos.
var props = map[string][]string{
	"companies": {"name", "nome_fantasia", "city", "state", "estados_onde_opera", "tipo_de_veiculo", "frota_total", "frota_contratada", "veiculos_instalados", "veiculos_offline",
		"nota_do_chs", "chs_score", "chs_aderencia", "chs_nota_engajamento", "chs_nota_adimplencia", "chs_nota_usuario", "chs_nota_motorista", "chs_nota_tickets_positivos", "tickets_suporte__chs_",
		"notes_last_contacted", "notes_last_updated", "hs_last_logged_call_date", "data_inicio_do_contrato_original", "data_de_termino_do_contrato", "ciclo_experiencia", "status_contrato", "status_financeiro",
		"lifecyclestage", "num_associated_contacts", "num_associated_deals", "createdate"},
	"calls":    {"hs_timestamp", "hs_call_title", "hs_call_status", "hs_call_direction", "hs_call_duration", "hs_call_body"},
	"notes":    {"hs_timestamp", "hs_note_body"},
	"emails":   {"hs_timestamp", "hs_email_subject", "hs_email_direction"},
	"meetings": {"hs_timestamp", "hs_meeting_title", "hs_meeting_outcome", "hs_meeting_body", "hs_internal_meeting_notes"},
	"tickets":  {"subject", "createdate", "closed_date", "hs_ticket_category", "hs_ticket_priority", "time_to_close"},
	"contacts": {"firstname", "lastname", "jobtitle", "notes_last_contacted"},
}

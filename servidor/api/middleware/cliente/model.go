package cliente

import (
	"time"

	modcliente "analisefrota/modules/cliente"
)

// Cliente_t reúne o que os casos de uso de cliente recebem do controller
type Cliente_t struct {
	ID        int64
	Nome      string
	HubspotID string
	Token     string
}

// ClienteComColeta é o cliente com a situação da última busca (resposta da lista)
type ClienteComColeta_t struct {
	modcliente.Cliente_t
	Coleta Coleta_t `json:"coleta"`
}

type Coleta_t struct {
	Situacao    string     `json:"situacao"`
	UltimoDia   string     `json:"ultimo_dia"`
	TerminadaEm *time.Time `json:"terminada_em"`
	Mensagem    string     `json:"mensagem"`
}

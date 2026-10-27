package whatsapp

import (
	modcontato "analisefrota/modules/contato"
	modwhatsapp "analisefrota/modules/whatsapp"
)

// WhatsApp_t reúne o que os casos de uso do WhatsApp dos motoristas recebem do controller
type WhatsApp_t struct {
	ClienteID  int64
	UsuarioID  int64
	Motorista  string
	Telefone   string
	Autorizado bool
	Ativo      bool
	Dia        string // AAAA-MM-DD; vazio = ontem
}

// Painel_t é a tela do WhatsApp de um cliente
type Painel_t struct {
	Configurado bool                   `json:"configurado"`
	Automatico  bool                   `json:"automatico"`
	Hora        int                    `json:"hora"`
	Template    string                 `json:"template"`
	Contatos    []modcontato.Contato_t `json:"contatos"`
	Motoristas  []string               `json:"motoristas"` // quem rodou nos últimos 30 dias, para escolher na lista
	Envios      []modwhatsapp.Envio_t  `json:"envios"`
}

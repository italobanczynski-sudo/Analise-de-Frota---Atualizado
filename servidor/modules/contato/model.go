package contato

import "time"

// Contato_t é o que sai para a tela: do telefone, só os 4 últimos dígitos
type Contato_t struct {
	Motorista     string    `json:"motorista"`
	TelefoneFinal string    `json:"telefone_final"`
	Ativo         bool      `json:"ativo"`
	AutorizadoEm  time.Time `json:"autorizado_em"`
}

// Destino_t é quem recebe a mensagem: telefone completo, decifrado só para o envio
type Destino_t struct {
	Motorista string
	Telefone  string // +55DDNNNNNNNNN
}

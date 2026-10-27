package whatsapp

import "time"

// Resumo_t é a análise de um motorista num dia: os números principais e as variáveis do template
type Resumo_t struct {
	Motorista string            `json:"motorista"`
	Dia       string            `json:"dia"`
	Km        *float64          `json:"km"`
	Kml       *float64          `json:"kml"`
	Nota      *float64          `json:"nota"`
	Evolucao  string            `json:"evolucao"`
	Dica      string            `json:"dica"`
	Variaveis map[string]string `json:"-"`
}

// Previa_t é o resumo com o texto pronto e se o motorista tem WhatsApp cadastrado e ativo
type Previa_t struct {
	Resumo_t
	Texto       string `json:"texto"`
	TemWhatsApp bool   `json:"tem_whatsapp"`
}

// Envio_t é uma mensagem do histórico
type Envio_t struct {
	Motorista string    `json:"motorista"`
	Dia       string    `json:"dia"`
	Situacao  string    `json:"situacao"`
	Erro      string    `json:"erro"`
	CriadoEm  time.Time `json:"criado_em"`
}

// Resultado_t conta o que aconteceu num envio
type Resultado_t struct {
	Dia         string   `json:"dia"`
	Enviadas    int      `json:"enviadas"`
	JaEnviadas  int      `json:"ja_enviadas"`
	SemWhatsApp int      `json:"sem_whatsapp"`
	Erros       []string `json:"erros"`
	Teste       bool     `json:"teste"`
}

const (
	SituacaoEnviando = "enviando"
	SituacaoEnviada  = "enviada"
	SituacaoErro     = "erro"
)

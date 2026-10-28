package mensagem

import "time"

// Resumo_t é a análise de um motorista num dia: os números principais e as variáveis da mensagem
type Resumo_t struct {
	Motorista string   `json:"motorista"`
	Dia       string   `json:"dia"`
	Km        *float64 `json:"km"`
	Kml       *float64 `json:"kml"`
	Nota      *float64 `json:"nota"`
	Evolucao  string   `json:"evolucao"`
	Dica      string   `json:"dica"`
	// DiaAnterior: o dia com que a análise foi comparada (o anterior em que o motorista rodou; vazio = nenhum)
	DiaAnterior string `json:"dia_anterior"`
	// ponto de atenção do dia: o percentual que mais piorou contra o dia anterior (Foco vazio = nenhum)
	Foco       string            `json:"foco"`
	FocoRotulo string            `json:"foco_rotulo"`
	FocoValor  *float64          `json:"foco_valor"`
	FocoMedia  *float64          `json:"foco_media"` // valor do ponto apontado no dia anterior
	Melhorou   int               `json:"melhorou"`
	Piorou     int               `json:"piorou"`
	Comparados int               `json:"comparados"`
	Variaveis  map[string]string `json:"-"`
}

// Mensagem_t é a análise com o texto pronto para copiar e mandar pelo WhatsApp
type Mensagem_t struct {
	Resumo_t
	Texto   string     `json:"texto"`
	Enviada *time.Time `json:"enviada"` // quando foi marcada como enviada (nil = ainda não)
}

// Acompanhamento_t é o histórico das mensagens de um motorista: quantos dias recebeu e se melhorou no que foi apontado
type Acompanhamento_t struct {
	Motorista   string        `json:"motorista"`
	DiasEnviada int           `json:"dias_enviada"`
	Envios      []Envio_t     `json:"envios"`
	Focos       []FocoTotal_t `json:"focos"` // cada ponto apontado: quantas vezes e em quantas ele melhorou depois
}

// Envio_t é um dia com mensagem enviada: o retrato da análise e o que aconteceu nos dias seguintes
type Envio_t struct {
	Dia        string    `json:"dia"`
	Foco       string    `json:"foco"`
	FocoRotulo string    `json:"foco_rotulo"`
	FocoValor  *float64  `json:"foco_valor"`
	FocoMedia  *float64  `json:"foco_media"`
	Depois     *float64  `json:"depois"`      // média do ponto apontado nos dias seguintes (até DiasDepois)
	DiasDepois int       `json:"dias_depois"` // quantos dias com dado entraram nessa média
	Resultado  string    `json:"resultado"`   // melhorou, nao_melhorou, aguardando, sem_ponto
	Melhorou   int       `json:"melhorou"`
	Piorou     int       `json:"piorou"`
	Comparados int       `json:"comparados"`
	Nota       *float64  `json:"nota"`
	MarcadoEm  time.Time `json:"marcado_em"`
}

type FocoTotal_t struct {
	Foco       string `json:"foco"`
	Rotulo     string `json:"rotulo"`
	Vezes      int    `json:"vezes"`
	Melhorou   int    `json:"melhorou"`
	Aguardando int    `json:"aguardando"`
}

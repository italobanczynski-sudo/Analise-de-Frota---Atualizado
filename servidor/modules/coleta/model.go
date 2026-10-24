package coleta

import "time"

// Situacao_t é como está a busca automática de um cliente
type Situacao_t struct {
	ClienteID   int64      `json:"cliente_id"`
	UltimoDia   string     `json:"ultimo_dia"` // último dia de Brasília buscado por completo
	IniciadaEm  *time.Time `json:"iniciada_em"`
	TerminadaEm *time.Time `json:"terminada_em"`
	Situacao    string     `json:"situacao"` // nunca, rodando, ok, erro
	Mensagem    string     `json:"mensagem"`
	Avisos      []string   `json:"avisos"`
}

// Resultado_t resume uma busca
type Resultado_t struct {
	DiasComDados int      `json:"dias_com_dados"`
	Veiculos     int      `json:"veiculos"`
	Performance  bool     `json:"performance"`
	Odometros    int      `json:"odometros"`
	MesesPremio  []string `json:"meses_premiacao"`
	Avisos       []string `json:"avisos"`
}

const (
	SituacaoNunca   = "nunca"
	SituacaoRodando = "rodando"
	SituacaoOk      = "ok"
	SituacaoErro    = "erro"
)

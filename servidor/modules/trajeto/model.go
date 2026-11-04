package trajeto

import "time"

// Passo entre os pontos do trajeto: a API só dá a última posição de cada janela
const Passo = 5 * time.Minute

// margemFinal: uma janela só é definitiva se já tinha acabado havia 10 min quando foi buscada
// (o equipamento manda a posição a cada 30 s e, numa área sem sinal, entrega o histórico depois)
const margemFinal = 10 * time.Minute

// orcamento: tempo máximo de chamadas por pedido; o resto fica para o próximo pedido (o servidor corta a resposta em 60 s)
const orcamento = 40 * time.Second

// DiasMax: até quantos dias para trás o trajeto pode ser pedido
const DiasMax = 93

// Ponto_t é uma posição do trajeto
type Ponto_t struct {
	Data string  `json:"data"` // como a API mandou: "AAAA-MM-DD HH:MM:SS" (UTC)
	Hora string  `json:"hora"` // hora de Brasília, "HH:MM:SS"
	Lat  float64 `json:"lat"`
	Lon  float64 `json:"lon"`
}

// Trajeto_t é a resposta: os pontos em ordem e quantas janelas de 5 min o pedido cobre.
// Completo = false quando o tempo do pedido acabou antes de todas as janelas: basta pedir de novo,
// as janelas já buscadas vêm do banco.
type Trajeto_t struct {
	Pontos   []Ponto_t `json:"pontos"`
	Janelas  int       `json:"janelas"`
	Faltam   int       `json:"faltam"`
	Completo bool      `json:"completo"`
}

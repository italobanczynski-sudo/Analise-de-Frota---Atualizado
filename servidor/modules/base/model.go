package base

import (
	"encoding/json"
	"time"
)

// Documento_t é um documento da base do painel: o id e os dados como o painel os gravou
type Documento_t struct {
	ID           string          `json:"id"`
	Dados        json.RawMessage `json:"dados"`
	AtualizadoEm time.Time       `json:"atualizado_em"`
}

// Coleções que o painel usa; qualquer outra é recusada
var Colecoes = []string{"clientes", "planilhas", "tratativas", "relacionamento", "hubspot", "config"}

// Limite de um documento: o painel divide as planilhas em partes de até 200 KB
const TamanhoMaximo = 1 << 20

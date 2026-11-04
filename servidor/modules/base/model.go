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

// Coleções que o painel usa; qualquer outra é recusada. Ao criar uma coleção nova no index.html, inclua aqui:
// sem isso o painel no servidor guarda só no navegador de quem digitou, sem avisar.
var Colecoes = []string{
	"clientes", "planilhas", "tratativas", "relacionamento", "hubspot", "config",
	// retenção: ficha, combinados, marcos, próximos passos, fases e resumo da reunião
	"fichas", "combinados", "marcos", "passos", "dias_excluidos", "premissas",
	// foto diária do risco, histórico do diesel, liberações de bloqueio e dados suspeitos tratados
	"risco_foto", "diesel_hist", "liberacoes", "suspeitos_trat",
}

// Limite de um documento: o painel divide as planilhas em partes de até 200 KB
const TamanhoMaximo = 1 << 20

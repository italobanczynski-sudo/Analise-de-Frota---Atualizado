package trajeto_repo

import "time"

// PosicaoRow_t é uma janela de 5 minutos já consultada na API (vazio = sem posição nela)
type PosicaoRow_t struct {
	ClienteID int64     `db:"cliente_id"`
	Placa     string    `db:"placa"`
	JanelaIni time.Time `db:"janela_ini"`
	Data      string    `db:"data"`
	Lat       *float64  `db:"lat"`
	Lon       *float64  `db:"lon"`
	Vazio     bool      `db:"vazio"`
	BuscadoEm time.Time `db:"buscado_em"`
}

// RastroRow_t é uma posição gravada pela frota ao vivo (data em UTC)
type RastroRow_t struct {
	ClienteID  int64     `db:"cliente_id"`
	Placa      string    `db:"placa"`
	Data       time.Time `db:"data"`
	Lat        float64   `db:"lat"`
	Lon        float64   `db:"lon"`
	Status     string    `db:"status"`
	Velocidade *float64  `db:"velocidade"`
}

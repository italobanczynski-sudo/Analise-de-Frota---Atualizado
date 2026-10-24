package odometro_repo

import "time"

type OdometroRow_t struct {
	ClienteID int64     `db:"cliente_id"`
	Placa     string    `db:"placa"`
	Odometro  *float64  `db:"odometro"`
	Coletado  string    `db:"coletado"`
	BuscadoEm time.Time `db:"buscado_em"`
}

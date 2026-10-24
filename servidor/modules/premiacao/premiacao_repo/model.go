package premiacao_repo

type PremiacaoRow_t struct {
	ClienteID int64    `db:"cliente_id"`
	Mes       string   `db:"mes"`
	Motorista string   `db:"motorista"`
	Nota      *float64 `db:"nota"`
	Km        *float64 `db:"km"`
	Premio    *float64 `db:"premio"`
}

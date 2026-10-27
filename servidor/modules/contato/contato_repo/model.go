package contato_repo

import "time"

type ContatoRow_t struct {
	ClienteID       int64     `db:"cliente_id"`
	Motorista       string    `db:"motorista"`
	TelefoneCifrado []byte    `db:"telefone_cifrado"`
	TelefoneFinal   string    `db:"telefone_final"`
	Ativo           bool      `db:"ativo"`
	AutorizadoEm    time.Time `db:"autorizado_em"`
	AutorizadoPor   *int64    `db:"autorizado_por"`
}

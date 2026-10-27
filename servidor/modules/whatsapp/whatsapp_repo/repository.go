package whatsapp_repo

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"analisefrota/adapters/dbPostgres"
)

type EnvioRow_t struct {
	ClienteID int64     `db:"cliente_id"`
	Motorista string    `db:"motorista"`
	Dia       string    `db:"dia"`
	Situacao  string    `db:"situacao"`
	TwilioSID string    `db:"twilio_sid"`
	Erro      string    `db:"erro"`
	CriadoEm  time.Time `db:"criado_em"`
}

type Repository_I interface {
	// Reservar marca a mensagem como "enviando"; devolve false se o dia do motorista já foi enviado ou está saindo.
	// Uma mensagem que deu erro pode ser reservada de novo.
	Reservar(ctx context.Context, clienteID int64, motorista, dia string) (bool, error)
	Concluir(ctx context.Context, clienteID int64, motorista, dia, situacao, sid, erro string) error
	Listar(ctx context.Context, clienteID int64, desde string) ([]EnvioRow_t, error)
	BuscarConfig(ctx context.Context) (ConfigRow_t, bool, error)
	SalvarConfig(ctx context.Context, row ConfigRow_t) error
}

type Repository_t struct {
	Db *sqlx.DB
}

func NewRepository() Repository_I {
	return &Repository_t{Db: dbPostgres.GetPgConn()}
}

const sqlReservar = `INSERT INTO envios_whatsapp (cliente_id, motorista, dia, situacao) VALUES ($1, $2, $3, 'enviando')
ON CONFLICT (cliente_id, motorista, dia) DO UPDATE SET situacao = 'enviando', erro = '', criado_em = now()
WHERE envios_whatsapp.situacao = 'erro'`

const sqlConcluir = `UPDATE envios_whatsapp SET situacao = $4, twilio_sid = $5, erro = $6 WHERE cliente_id = $1 AND motorista = $2 AND dia = $3`

const sqlListar = `SELECT cliente_id, motorista, dia::text AS dia, situacao, twilio_sid, erro, criado_em
FROM envios_whatsapp WHERE cliente_id = $1 AND dia >= $2 ORDER BY dia DESC, motorista`

func (r *Repository_t) Reservar(ctx context.Context, clienteID int64, motorista, dia string) (bool, error) {
	res, err := r.Db.ExecContext(ctx, sqlReservar, clienteID, motorista, dia)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n == 1, err
}

func (r *Repository_t) Concluir(ctx context.Context, clienteID int64, motorista, dia, situacao, sid, erro string) error {
	_, err := r.Db.ExecContext(ctx, sqlConcluir, clienteID, motorista, dia, situacao, sid, erro)
	return err
}

func (r *Repository_t) Listar(ctx context.Context, clienteID int64, desde string) ([]EnvioRow_t, error) {
	rows := []EnvioRow_t{}
	err := r.Db.SelectContext(ctx, &rows, sqlListar, clienteID, desde)
	return rows, err
}

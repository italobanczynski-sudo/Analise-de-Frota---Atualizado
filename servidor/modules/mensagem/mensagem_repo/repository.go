package mensagem_repo

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"

	"analisefrota/adapters/dbPostgres"
)

// EnviadaRow_t é uma mensagem marcada como enviada, com o retrato da análise do dia
type EnviadaRow_t struct {
	ClienteID  int64     `db:"cliente_id"`
	Motorista  string    `db:"motorista"`
	Dia        string    `db:"dia"`
	Foco       string    `db:"foco"`
	FocoValor  *float64  `db:"foco_valor"`
	FocoMedia  *float64  `db:"foco_media"`
	Melhorou   int       `db:"melhorou"`
	Piorou     int       `db:"piorou"`
	Comparados int       `db:"comparados"`
	Nota       *float64  `db:"nota"`
	MarcadoEm  time.Time `db:"marcado_em"`
	MarcadoPor *int64    `db:"marcado_por"`
}

type Repository_I interface {
	// Marcar grava (ou refaz) a mensagem do motorista no dia como enviada
	Marcar(ctx context.Context, row EnviadaRow_t) error
	// Desmarcar devolve false se não estava marcada
	Desmarcar(ctx context.Context, clienteID int64, motorista, dia string) (bool, error)
	ListarDia(ctx context.Context, clienteID int64, dia string) ([]EnviadaRow_t, error)
	ListarMotorista(ctx context.Context, clienteID int64, motorista string) ([]EnviadaRow_t, error)
}

type Repository_t struct {
	Db *sqlx.DB
}

func NewRepository() Repository_I {
	return &Repository_t{Db: dbPostgres.GetPgConn()}
}

const sqlColunas = `cliente_id, motorista, dia::text AS dia, foco, foco_valor, foco_media, melhorou, piorou, comparados, nota, marcado_em, marcado_por`

const sqlMarcar = `INSERT INTO mensagens_enviadas (cliente_id, motorista, dia, foco, foco_valor, foco_media, melhorou, piorou, comparados, nota, marcado_por)
VALUES (:cliente_id, :motorista, :dia, :foco, :foco_valor, :foco_media, :melhorou, :piorou, :comparados, :nota, :marcado_por)
ON CONFLICT (cliente_id, motorista, dia) DO UPDATE SET foco = EXCLUDED.foco, foco_valor = EXCLUDED.foco_valor, foco_media = EXCLUDED.foco_media,
	melhorou = EXCLUDED.melhorou, piorou = EXCLUDED.piorou, comparados = EXCLUDED.comparados, nota = EXCLUDED.nota,
	marcado_por = EXCLUDED.marcado_por, marcado_em = now()`

const sqlDesmarcar = `DELETE FROM mensagens_enviadas WHERE cliente_id = $1 AND motorista = $2 AND dia = $3`

const sqlListarDia = `SELECT ` + sqlColunas + ` FROM mensagens_enviadas WHERE cliente_id = $1 AND dia = $2`

// o motorista é comparado sem diferenciar maiúsculas nem espaços (o painel e a API escrevem de jeitos diferentes)
const sqlListarMotorista = `SELECT ` + sqlColunas + ` FROM mensagens_enviadas
WHERE cliente_id = $1 AND upper(regexp_replace(trim(motorista), '\s+', ' ', 'g')) = upper(regexp_replace(trim($2), '\s+', ' ', 'g'))
ORDER BY dia`

func (r *Repository_t) Marcar(ctx context.Context, row EnviadaRow_t) error {
	_, err := r.Db.NamedExecContext(ctx, sqlMarcar, row)
	return err
}

func (r *Repository_t) Desmarcar(ctx context.Context, clienteID int64, motorista, dia string) (bool, error) {
	res, err := r.Db.ExecContext(ctx, sqlDesmarcar, clienteID, motorista, dia)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (r *Repository_t) ListarDia(ctx context.Context, clienteID int64, dia string) ([]EnviadaRow_t, error) {
	rows := []EnviadaRow_t{}
	err := r.Db.SelectContext(ctx, &rows, sqlListarDia, clienteID, dia)
	return rows, err
}

func (r *Repository_t) ListarMotorista(ctx context.Context, clienteID int64, motorista string) ([]EnviadaRow_t, error) {
	rows := []EnviadaRow_t{}
	err := r.Db.SelectContext(ctx, &rows, sqlListarMotorista, clienteID, motorista)
	return rows, err
}

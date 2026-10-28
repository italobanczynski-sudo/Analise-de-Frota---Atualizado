package diario_repo

import (
	"context"

	"github.com/jmoiron/sqlx"

	"analisefrota/adapters/dbPostgres"
)

type Repository_I interface {
	// SubstituirDia apaga o dia do cliente e grava os registros novos, tudo numa transação
	SubstituirDia(ctx context.Context, clienteID int64, dia string, rows []RegistroRow_t) error
	Listar(ctx context.Context, clienteID int64, ini, fim string) ([]RegistroRow_t, error)
	// DiasComIndicadores: dias do período que já têm os indicadores de condução gravados
	DiasComIndicadores(ctx context.Context, clienteID int64, ini, fim string) ([]string, error)
}

type Repository_t struct {
	Db *sqlx.DB
}

func NewRepository() Repository_I {
	return &Repository_t{Db: dbPostgres.GetPgConn()}
}

func (r *Repository_t) SubstituirDia(ctx context.Context, clienteID int64, dia string, rows []RegistroRow_t) error {
	tx, err := r.Db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // sem efeito depois do Commit
	if _, err := tx.ExecContext(ctx, sqlApagarDia, clienteID, dia); err != nil {
		return err
	}
	for _, row := range rows {
		if _, err := tx.NamedExecContext(ctx, sqlInserir, row); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository_t) DiasComIndicadores(ctx context.Context, clienteID int64, ini, fim string) ([]string, error) {
	dias := []string{}
	err := r.Db.SelectContext(ctx, &dias, sqlDiasComIndicadores, clienteID, ini, fim)
	return dias, err
}

func (r *Repository_t) Listar(ctx context.Context, clienteID int64, ini, fim string) ([]RegistroRow_t, error) {
	rows := []RegistroRow_t{}
	err := r.Db.SelectContext(ctx, &rows, sqlListar, clienteID, ini, fim)
	return rows, err
}

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
	// DiasComIndicadores: dias do período que já têm os indicadores de condução (e as linhas dos motoristas) gravados
	DiasComIndicadores(ctx context.Context, clienteID int64, ini, fim string) ([]string, error)
	// SubstituirMotoristasDia apaga as linhas dos motoristas do dia e grava as novas, numa transação
	SubstituirMotoristasDia(ctx context.Context, clienteID int64, dia string, rows []RegistroRow_t) error
	ListarMotoristas(ctx context.Context, clienteID int64, ini, fim string) ([]RegistroRow_t, error)
}

type Repository_t struct {
	Db *sqlx.DB
}

func NewRepository() Repository_I {
	return &Repository_t{Db: dbPostgres.GetPgConn()}
}

func (r *Repository_t) SubstituirDia(ctx context.Context, clienteID int64, dia string, rows []RegistroRow_t) error {
	return r.substituir(ctx, sqlApagarDia, sqlInserir, clienteID, dia, rows)
}

func (r *Repository_t) SubstituirMotoristasDia(ctx context.Context, clienteID int64, dia string, rows []RegistroRow_t) error {
	return r.substituir(ctx, sqlApagarMotoristasDia, sqlInserirMotorista, clienteID, dia, rows)
}

// substituir apaga o dia do cliente (apagar) e grava as linhas (inserir), numa transação só
func (r *Repository_t) substituir(ctx context.Context, apagar, inserir string, clienteID int64, dia string, rows []RegistroRow_t) error {
	tx, err := r.Db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // sem efeito depois do Commit
	if _, err := tx.ExecContext(ctx, apagar, clienteID, dia); err != nil {
		return err
	}
	for _, row := range rows {
		if _, err := tx.NamedExecContext(ctx, inserir, row); err != nil {
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
	return r.listar(ctx, sqlListar, clienteID, ini, fim)
}

func (r *Repository_t) ListarMotoristas(ctx context.Context, clienteID int64, ini, fim string) ([]RegistroRow_t, error) {
	return r.listar(ctx, sqlListarMotoristas, clienteID, ini, fim)
}

func (r *Repository_t) listar(ctx context.Context, consulta string, clienteID int64, ini, fim string) ([]RegistroRow_t, error) {
	rows := []RegistroRow_t{}
	err := r.Db.SelectContext(ctx, &rows, consulta, clienteID, ini, fim)
	return rows, err
}

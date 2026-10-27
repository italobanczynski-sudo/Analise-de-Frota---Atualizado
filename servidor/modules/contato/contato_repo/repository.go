package contato_repo

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"

	"analisefrota/adapters/dbPostgres"
)

// ErrNaoEncontrado é traduzido pelo módulo em mensagem para o usuário
var ErrNaoEncontrado = errors.New("contato_repo: não encontrado")

type Repository_I interface {
	Listar(ctx context.Context, clienteID int64) ([]ContatoRow_t, error)
	ListarAtivos(ctx context.Context, clienteID int64) ([]ContatoRow_t, error)
	// Salvar cria ou troca o telefone do motorista (e o reativa)
	Salvar(ctx context.Context, row ContatoRow_t) error
	Ativar(ctx context.Context, clienteID int64, motorista string, ativo bool) error
	Remover(ctx context.Context, clienteID int64, motorista string) error
}

type Repository_t struct {
	Db *sqlx.DB
}

func NewRepository() Repository_I {
	return &Repository_t{Db: dbPostgres.GetPgConn()}
}

func (r *Repository_t) Listar(ctx context.Context, clienteID int64) ([]ContatoRow_t, error) {
	rows := []ContatoRow_t{}
	err := r.Db.SelectContext(ctx, &rows, sqlListar, clienteID)
	return rows, err
}

func (r *Repository_t) ListarAtivos(ctx context.Context, clienteID int64) ([]ContatoRow_t, error) {
	rows := []ContatoRow_t{}
	err := r.Db.SelectContext(ctx, &rows, sqlListarAtivos, clienteID)
	return rows, err
}

func (r *Repository_t) Salvar(ctx context.Context, row ContatoRow_t) error {
	_, err := r.Db.NamedExecContext(ctx, sqlSalvar, row)
	return err
}

func (r *Repository_t) Ativar(ctx context.Context, clienteID int64, motorista string, ativo bool) error {
	return umaLinha(r.Db.ExecContext(ctx, sqlAtivar, clienteID, motorista, ativo))
}

func (r *Repository_t) Remover(ctx context.Context, clienteID int64, motorista string) error {
	return umaLinha(r.Db.ExecContext(ctx, sqlRemover, clienteID, motorista))
}

type resultado interface{ RowsAffected() (int64, error) }

func umaLinha(res resultado, err error) error {
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrNaoEncontrado
	}
	return nil
}

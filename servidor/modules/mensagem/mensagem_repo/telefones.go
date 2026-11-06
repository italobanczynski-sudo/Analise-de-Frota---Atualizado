package mensagem_repo

import (
	"context"

	"github.com/jmoiron/sqlx"

	"analisefrota/adapters/dbPostgres"
)

// TelefoneRow_t é o telefone de um motorista do cliente (chave = nome sem diferença de maiúsculas e espaços)
type TelefoneRow_t struct {
	ClienteID     int64  `db:"cliente_id"`
	Chave         string `db:"chave"`
	Motorista     string `db:"motorista"`
	Telefone      string `db:"telefone"`
	AtualizadoPor *int64 `db:"atualizado_por"`
}

type Telefones_I interface {
	// Salvar grava (ou troca) o telefone do motorista
	Salvar(ctx context.Context, row TelefoneRow_t) error
	// Apagar tira o telefone do motorista (sem erro se não havia)
	Apagar(ctx context.Context, clienteID int64, chave string) error
	// Listar devolve os telefones do cliente por chave do motorista
	Listar(ctx context.Context, clienteID int64) (map[string]string, error)
}

type Telefones_t struct {
	Db *sqlx.DB
}

func NewTelefones() Telefones_I {
	return &Telefones_t{Db: dbPostgres.GetPgConn()}
}

const sqlSalvarTelefone = `INSERT INTO telefones_motoristas (cliente_id, chave, motorista, telefone, atualizado_por)
VALUES (:cliente_id, :chave, :motorista, :telefone, :atualizado_por)
ON CONFLICT (cliente_id, chave) DO UPDATE SET motorista = EXCLUDED.motorista, telefone = EXCLUDED.telefone,
	atualizado_por = EXCLUDED.atualizado_por, atualizado_em = now()`

func (r *Telefones_t) Salvar(ctx context.Context, row TelefoneRow_t) error {
	_, err := r.Db.NamedExecContext(ctx, sqlSalvarTelefone, row)
	return err
}

func (r *Telefones_t) Apagar(ctx context.Context, clienteID int64, chave string) error {
	_, err := r.Db.ExecContext(ctx, `DELETE FROM telefones_motoristas WHERE cliente_id = $1 AND chave = $2`, clienteID, chave)
	return err
}

func (r *Telefones_t) Listar(ctx context.Context, clienteID int64) (map[string]string, error) {
	rows := []TelefoneRow_t{}
	if err := r.Db.SelectContext(ctx, &rows, `SELECT cliente_id, chave, motorista, telefone, atualizado_por FROM telefones_motoristas WHERE cliente_id = $1`, clienteID); err != nil {
		return nil, err
	}
	out := make(map[string]string, len(rows))
	for _, t := range rows {
		out[t.Chave] = t.Telefone
	}
	return out, nil
}

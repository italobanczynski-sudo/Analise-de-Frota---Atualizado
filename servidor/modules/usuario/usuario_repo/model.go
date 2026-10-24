package usuario_repo

import (
	"database/sql"
	"time"
)

type CodigoRow_t struct {
	ID         int64  `db:"id"`
	CodigoHash []byte `db:"codigo_hash"`
	Tentativas int    `db:"tentativas"`
}

type UsuarioRow_t struct {
	ID           int64          `db:"id"`
	Nome         string         `db:"nome"`
	Email        string         `db:"email"`
	Papel        string         `db:"papel"`
	SenhaHash    sql.NullString `db:"senha_hash"`
	Ativo        bool           `db:"ativo"`
	CriadoEm     time.Time      `db:"criado_em"`
	UltimoAcesso *time.Time     `db:"ultimo_acesso"`
}

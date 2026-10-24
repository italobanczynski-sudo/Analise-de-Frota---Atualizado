package sessao_repo

import "time"

type SessaoRow_t struct {
	TokenHash []byte    `db:"token_hash"`
	UsuarioID int64     `db:"usuario_id"`
	ExpiraEm  time.Time `db:"expira_em"`
}

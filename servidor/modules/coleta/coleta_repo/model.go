package coleta_repo

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type SituacaoRow_t struct {
	ClienteID   int64          `db:"cliente_id"`
	UltimoDia   sql.NullString `db:"ultimo_dia"`
	IniciadaEm  *time.Time     `db:"iniciada_em"`
	TerminadaEm *time.Time     `db:"terminada_em"`
	Situacao    string         `db:"situacao"`
	Mensagem    string         `db:"mensagem"`
	Avisos      pq.StringArray `db:"avisos"`
}

package whatsapp_repo

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// ConfigRow_t é a conexão com a Twilio salva pelo administrador (linha única)
type ConfigRow_t struct {
	AccountSID       string    `db:"account_sid"`
	AuthTokenCifrado []byte    `db:"auth_token_cifrado"`
	Remetente        string    `db:"remetente"`
	TemplateSID      string    `db:"template_sid"`
	Automatico       bool      `db:"automatico"`
	Hora             int       `db:"hora"`
	AtualizadoEm     time.Time `db:"atualizado_em"`
	AtualizadoPor    *int64    `db:"atualizado_por"`
}

const sqlBuscarConfig = `SELECT account_sid, auth_token_cifrado, remetente, template_sid, automatico, hora, atualizado_em, atualizado_por
FROM whatsapp_config WHERE id = 1`

const sqlSalvarConfig = `INSERT INTO whatsapp_config (id, account_sid, auth_token_cifrado, remetente, template_sid, automatico, hora, atualizado_por)
VALUES (1, :account_sid, :auth_token_cifrado, :remetente, :template_sid, :automatico, :hora, :atualizado_por)
ON CONFLICT (id) DO UPDATE SET account_sid = EXCLUDED.account_sid, auth_token_cifrado = EXCLUDED.auth_token_cifrado,
	remetente = EXCLUDED.remetente, template_sid = EXCLUDED.template_sid, automatico = EXCLUDED.automatico,
	hora = EXCLUDED.hora, atualizado_por = EXCLUDED.atualizado_por, atualizado_em = now()`

// BuscarConfig devolve false quando o administrador ainda não salvou a conexão
func (r *Repository_t) BuscarConfig(ctx context.Context) (ConfigRow_t, bool, error) {
	var row ConfigRow_t
	err := r.Db.GetContext(ctx, &row, sqlBuscarConfig)
	if errors.Is(err, sql.ErrNoRows) {
		return row, false, nil
	}
	return row, err == nil, err
}

func (r *Repository_t) SalvarConfig(ctx context.Context, row ConfigRow_t) error {
	_, err := r.Db.NamedExecContext(ctx, sqlSalvarConfig, row)
	return err
}

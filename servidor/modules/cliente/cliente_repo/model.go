package cliente_repo

import "time"

type ClienteRow_t struct {
	ID        int64     `db:"id"`
	Nome      string    `db:"nome"`
	HubspotID string    `db:"hubspot_id"`
	TemToken  bool      `db:"tem_token"`
	CriadoEm  time.Time `db:"criado_em"`
}

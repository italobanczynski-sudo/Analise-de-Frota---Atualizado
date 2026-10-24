package cliente

import "time"

// Cliente_t é o que sai do módulo: o token nunca aparece, só se ele existe
type Cliente_t struct {
	ID        int64     `json:"id"`
	Nome      string    `json:"nome"`
	HubspotID string    `json:"hubspot_id"`
	TemToken  bool      `json:"tem_token"`
	CriadoEm  time.Time `json:"criado_em"`
}

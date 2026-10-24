package cliente

type CriarRequest struct {
	Nome      string `json:"nome" validate:"required,max=120"`
	HubspotID string `json:"hubspot_id" validate:"omitempty,numeric,max=20"`
	Token     string `json:"token" validate:"omitempty,max=4096"`
}

type TokenRequest struct {
	Token string `json:"token" validate:"required,max=4096"`
}

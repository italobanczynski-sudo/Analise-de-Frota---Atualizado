package whatsapp

type DiaRequest struct {
	Dia string `form:"dia" json:"dia" validate:"omitempty,datetime=2006-01-02"`
}

type ContatoRequest struct {
	Motorista  string `json:"motorista" validate:"required,max=120"`
	Telefone   string `json:"telefone" validate:"required,max=30"`
	Autorizado bool   `json:"autorizado"`
}

type AtivoRequest struct {
	Motorista string `json:"motorista" validate:"required,max=120"`
	Ativo     bool   `json:"ativo"`
}

type ConfigRequest struct {
	AccountSID  string `json:"account_sid" validate:"required,max=64"`
	AuthToken   string `json:"auth_token" validate:"max=128"` // vazio mantém o salvo
	Remetente   string `json:"remetente" validate:"required,max=40"`
	TemplateSID string `json:"template_sid" validate:"required,max=64"`
	Automatico  bool   `json:"automatico"`
	Hora        int    `json:"hora" validate:"min=0,max=23"`
}

type MotoristaRequest struct {
	Motorista string `form:"motorista" validate:"required,max=120"`
}

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

type MotoristaRequest struct {
	Motorista string `form:"motorista" validate:"required,max=120"`
}

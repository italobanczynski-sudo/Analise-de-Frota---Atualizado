package usuario

type CriarRequest struct {
	Nome  string `json:"nome" validate:"required,max=120"`
	Email string `json:"email" validate:"required,email,max=200"`
	Papel string `json:"papel" validate:"required,oneof=admin colaborador leitor"`
}

type AlterarRequest struct {
	Papel string `json:"papel" validate:"required,oneof=admin colaborador leitor"`
	Ativo *bool  `json:"ativo" validate:"required"`
}

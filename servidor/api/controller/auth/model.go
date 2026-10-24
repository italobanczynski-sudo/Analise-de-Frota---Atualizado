package auth

type EntrarRequest struct {
	Email string `json:"email" validate:"required,max=200"`
	Senha string `json:"senha" validate:"required,max=200"`
}

type CodigoRequest struct {
	Email string `json:"email" validate:"required,email,max=200"`
}

type ConfirmarCodigoRequest struct {
	Email  string `json:"email" validate:"required,email,max=200"`
	Codigo string `json:"codigo" validate:"required,max=12"`
	Senha  string `json:"senha" validate:"required,max=200"`
}

type ConviteRequest struct {
	Token string `json:"token" validate:"required,len=43"`
	Senha string `json:"senha" validate:"required,max=200"`
}

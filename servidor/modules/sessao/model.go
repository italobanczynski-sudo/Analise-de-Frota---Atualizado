package sessao

import "time"

// Sessao_t é um login aberto; o Token só existe na criação (vai para o cookie; no banco fica o hash)
type Sessao_t struct {
	Token     string
	UsuarioID int64
	ExpiraEm  time.Time
}

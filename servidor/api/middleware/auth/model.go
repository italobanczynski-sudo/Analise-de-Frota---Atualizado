package auth

import (
	"time"

	"analisefrota/modules/usuario"
)

// Auth_t reúne o que os casos de uso de login recebem do controller
type Auth_t struct {
	Email  string
	Senha  string
	Codigo string // código de 6 dígitos recebido por e-mail
	Token  string // token do convite ou do cookie de sessão
	Origem string // IP de quem tenta entrar (para a trava de tentativas)
}

// Entrada_t é o resultado de um login: a pessoa e o token do cookie
type Entrada_t struct {
	Usuario  usuario.Usuario_t
	Token    string
	ExpiraEm time.Time
}

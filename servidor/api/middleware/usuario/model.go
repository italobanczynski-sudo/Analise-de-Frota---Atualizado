package usuario

import (
	"strings"
	"time"

	"analisefrota/environment"
	modusuario "analisefrota/modules/usuario"
)

// Usuario_t reúne o que os casos de uso de usuário recebem do controller
type Usuario_t struct {
	ID       int64
	Nome     string
	Email    string
	Papel    string
	Ativo    bool
	QuemPede int64 // id de quem fez o pedido (um administrador)
}

// ComConvite_t é o usuário com o link para criar a senha (só aparece na hora em que o convite é gerado)
type ComConvite_t struct {
	Usuario  modusuario.Usuario_t `json:"usuario"`
	Link     string               `json:"link_convite"`
	ExpiraEm time.Time            `json:"expira_em"`
}

// linkConvite usa # para o token: o navegador não manda o que vem depois do # ao servidor (nem para logs)
func linkConvite(token string) string {
	return strings.TrimRight(environment.Cfg.Servidor.URLPublica, "/") + "/convite#" + token
}

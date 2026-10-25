package environment

import "strings"

type ServidorConfig_t struct {
	Porta string `envconfig:"PORTA" default:"8080"`
	// Endereço pelo qual as pessoas abrem a plataforma; vira o começo dos links de convite
	URLPublica string `envconfig:"URL_PUBLICA" default:"http://localhost:8080"`
	// Endereços que podem chamar a API pelo navegador (CORS), separados por vírgula
	OrigensPermitidas []string `envconfig:"ORIGENS_PERMITIDAS" default:"http://localhost:8080"`
	// Proxies na frente do servidor (nginx, Railway), em IP ou CIDR separados por vírgula. Só deles o servidor
	// aceita X-Forwarded-For: sem isso, atrás de um proxy todo mundo parece vir do mesmo IP e a trava de
	// tentativas de login bloquearia todos juntos. Vazio = o IP vem direto da conexão.
	ProxiesConfiaveis []string `envconfig:"PROXIES_CONFIAVEIS"`
}

// HTTPS indica se o cookie de login deve ir só por conexão segura
func (s ServidorConfig_t) HTTPS() bool { return strings.HasPrefix(s.URLPublica, "https://") }

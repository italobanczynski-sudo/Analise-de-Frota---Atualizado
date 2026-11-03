package environment

import "strings"

type HubspotConfig_t struct {
	// Token do app privado do HubSpot (pat-...) ou chave de acesso pessoal. Só leitura.
	// Vazio = o cartão "Cliente no HubSpot" fica desligado no servidor.
	Token string `envconfig:"HUBSPOT_TOKEN"`
	// Por quantos minutos a leitura de um cliente vale antes de buscar de novo no HubSpot
	CacheMin int `envconfig:"HUBSPOT_CACHE_MIN" default:"10"`
}

// Configurado indica se há token para consultar o HubSpot
func (h HubspotConfig_t) Configurado() bool { return strings.TrimSpace(h.Token) != "" }

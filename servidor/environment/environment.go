// Package environment é o único lugar do servidor que lê variáveis de ambiente.
// Localmente elas vêm do arquivo .env (fora do Git); em produção, do ambiente do contêiner.
package environment

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config_t struct {
	Servidor  ServidorConfig_t
	Database  DatabaseConfig_t
	Gobrax    GobraxConfig_t
	Seguranca SegurancaConfig_t
	Coleta    ColetaConfig_t
}

// Cfg é carregado uma vez por Load(), no início do main.
var Cfg *Config_t

func Load() {
	_ = godotenv.Load() // sem .env (produção) as variáveis já estão no ambiente
	Cfg = &Config_t{}
	for nome, alvo := range map[string]any{
		"servidor":  &Cfg.Servidor,
		"database":  &Cfg.Database,
		"gobrax":    &Cfg.Gobrax,
		"seguranca": &Cfg.Seguranca,
		"coleta":    &Cfg.Coleta,
	} {
		if err := envconfig.Process("", alvo); err != nil {
			log.Fatalf("environment: configuração de %s inválida: %v", nome, err)
		}
	}
	if err := Cfg.Seguranca.validar(); err != nil {
		log.Fatalf("environment: %v", err)
	}
}

package environment

type ServidorConfig_t struct {
	Porta string `envconfig:"PORTA" default:"8080"`
	// Endereços que podem chamar a API pelo navegador (CORS), separados por vírgula
	OrigensPermitidas []string `envconfig:"ORIGENS_PERMITIDAS" default:"http://localhost:8080"`
}

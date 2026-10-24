package environment

type GobraxConfig_t struct {
	APIBase string `envconfig:"GOBRAX_API_BASE" default:"https://gateway-v3.gobrax.com.br:8889/api/"`
	// Pausa entre chamadas: a API aceita rajada de 10 e depois 1 por segundo
	IntervaloMs int `envconfig:"GOBRAX_INTERVALO_MS" default:"120"`
}

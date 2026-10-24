package odometro

import "time"

// Odometro_t é o último odômetro de um veículo e quando a unidade eletrônica o mandou
type Odometro_t struct {
	Placa     string    `json:"placa"`
	Odometro  *float64  `json:"odometro"`
	Coletado  string    `json:"coletado"` // como a API manda ("AAAA-MM-DD HH:MM:SS")
	BuscadoEm time.Time `json:"buscado_em"`
}

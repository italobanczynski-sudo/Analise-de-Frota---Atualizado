package premiacao

// Premiacao_t é o resultado de um motorista num mês (AAAA-MM)
type Premiacao_t struct {
	Mes       string   `json:"mes"`
	Motorista string   `json:"motorista"`
	Nota      *float64 `json:"nota"`
	Km        *float64 `json:"km"`
	Premio    *float64 `json:"premio"`
}

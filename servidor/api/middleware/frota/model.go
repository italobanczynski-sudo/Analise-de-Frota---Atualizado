package frota

// Frota_t reúne o que as consultas de dados da frota recebem do controller
type Frota_t struct {
	ClienteID int64
	Inicio    string // AAAA-MM-DD
	Fim       string // AAAA-MM-DD
	Mes       string // AAAA-MM (premiação; vazio = todos)
}

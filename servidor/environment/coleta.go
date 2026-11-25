package environment

type ColetaConfig_t struct {
	Automatica bool `envconfig:"COLETA_AUTOMATICA" default:"true"`
	// De quanto em quanto tempo o servidor busca os dados de todos os clientes
	IntervaloMin int `envconfig:"COLETA_INTERVALO_MIN" default:"60"`
	// Quantos dias para trás na primeira busca de um cliente
	Dias int `envconfig:"COLETA_DIAS" default:"30"`
	// Até quantos dias para trás a busca vai completando o passado que falta no banco, poucos dias por vez
	HistoricoDias int `envconfig:"COLETA_HISTORICO_DIAS" default:"365"`
	// Quantos dias antigos entram em cada busca completa (de hora em hora)
	HistoricoPorBusca int `envconfig:"COLETA_HISTORICO_POR_BUSCA" default:"4"`
	// Entre uma busca completa e outra, de quanto em quanto tempo refaz só o dia de hoje (0 = desligado)
	HojeMin int `envconfig:"COLETA_HOJE_MIN" default:"15"`
}

type RastroConfig_t struct {
	// De quanto em quanto tempo o servidor grava a posição de todos os veículos (rastro da frota ao vivo); 0 = desligado
	IntervaloMin int `envconfig:"RASTRO_INTERVALO_MIN" default:"5"`
	// Quantos dias de rastro ficam guardados
	Dias int `envconfig:"RASTRO_DIAS" default:"7"`
}

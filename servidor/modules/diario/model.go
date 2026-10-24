package diario

// Registro_t é um veículo num dia de Brasília: estatísticas + indicadores de condução.
// Campo nil = a API não trouxe o dado. Motorista vazio = nenhum motorista vinculado no dia.
type Registro_t struct {
	ClienteID              int64    `json:"cliente_id"`
	Dia                    string   `json:"dia"`
	Placa                  string   `json:"placa"`
	Motorista              string   `json:"motorista"`
	Rodou                  bool     `json:"rodou"`
	Km                     *float64 `json:"km"`
	Litros                 *float64 `json:"litros"`
	Frenagens              *float64 `json:"frenagens"`
	FrenagensAlta          *float64 `json:"frenagens_alta"`
	VelocidadeMedia        *float64 `json:"velocidade_media"`
	Odometro               *float64 `json:"odometro"`
	Horas                  *float64 `json:"horas"`
	Nota                   *float64 `json:"nota"`
	FaixaVerde             *float64 `json:"faixa_verde"`
	InicioFaixaVerde       *float64 `json:"inicio_faixa_verde"`
	FinalFaixaVerde        *float64 `json:"final_faixa_verde"`
	EcoRoll                *float64 `json:"eco_roll"`
	PressaoIdeal           *float64 `json:"pressao_ideal"`
	PressaoAtencao         *float64 `json:"pressao_atencao"`
	Embalo                 *float64 `json:"embalo"`
	MotorOcioso            *float64 `json:"motor_ocioso"`
	PressaoAcelerador      *float64 `json:"pressao_acelerador"`
	ExcessoRotacao         *float64 `json:"excesso_rotacao"`
	TempoExcessoVelocidade *float64 `json:"tempo_excesso_velocidade"`
	PilotoAutomatico       *float64 `json:"piloto_automatico"`
	FreioMotor             *float64 `json:"freio_motor"`
}

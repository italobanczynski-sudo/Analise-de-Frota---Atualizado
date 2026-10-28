package diario_repo

type RegistroRow_t struct {
	ClienteID              int64    `db:"cliente_id"`
	Dia                    string   `db:"dia"`
	Placa                  string   `db:"placa"`
	Motorista              string   `db:"motorista"`
	Rodou                  bool     `db:"rodou"`
	Km                     *float64 `db:"km"`
	Litros                 *float64 `db:"litros"`
	Frenagens              *float64 `db:"frenagens"`
	FrenagensAlta          *float64 `db:"frenagens_alta"`
	VelocidadeMedia        *float64 `db:"velocidade_media"`
	Odometro               *float64 `db:"odometro"`
	Horas                  *float64 `db:"horas"`
	HorasMotor             *float64 `db:"horas_motor"`
	HorasPedal             *float64 `db:"horas_pedal"`
	Nota                   *float64 `db:"nota"`
	FaixaVerde             *float64 `db:"faixa_verde"`
	InicioFaixaVerde       *float64 `db:"inicio_faixa_verde"`
	FinalFaixaVerde        *float64 `db:"final_faixa_verde"`
	EcoRoll                *float64 `db:"eco_roll"`
	PressaoIdeal           *float64 `db:"pressao_ideal"`
	PressaoAtencao         *float64 `db:"pressao_atencao"`
	Embalo                 *float64 `db:"embalo"`
	MotorOcioso            *float64 `db:"motor_ocioso"`
	PressaoAcelerador      *float64 `db:"pressao_acelerador"`
	ExcessoRotacao         *float64 `db:"excesso_rotacao"`
	TempoExcessoVelocidade *float64 `db:"tempo_excesso_velocidade"`
	PilotoAutomatico       *float64 `db:"piloto_automatico"`
	FreioMotor             *float64 `db:"freio_motor"`
}

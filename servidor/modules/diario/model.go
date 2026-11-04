package diario

import (
	"time"

	"analisefrota/adapters/gobraxapi"
)

// Registro_t é um veículo num dia de Brasília: estatísticas + indicadores de condução. A mesma estrutura serve à
// linha de um motorista (registros_motoristas): aí os números são só do tempo do vínculo dele com o veículo.
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
	HorasMotor             *float64 `json:"horas_motor"` // ignição ligada: base do motor ligado parado
	HorasPedal             *float64 `json:"horas_pedal"` // pé no acelerador: base dos ranges do pedal
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

// Janela_t é um trecho do dia (UTC) em que um motorista estava vinculado ao veículo
type Janela_t struct {
	Ini time.Time
	Fim time.Time
}

// Pendente_t é um vínculo parcial: o motorista ficou só parte do dia no veículo e o dado dele vem da API
// consultada nos trechos dele (Janelas)
type Pendente_t struct {
	Placa     string
	Motorista string
	Janelas   []Janela_t
}

// PlanoMotoristas_t: as linhas prontas de quem ficou o dia todo no veículo (Inteiros) e os vínculos parciais a consultar
type PlanoMotoristas_t struct {
	Inteiros  []Registro_t
	Pendentes []Pendente_t
}

// Parte_t é o que a API devolveu para um trecho de um vínculo parcial
type Parte_t struct {
	Estat []gobraxapi.EstatisticaRec_t
	Perf  []gobraxapi.PerformanceRec_t
}

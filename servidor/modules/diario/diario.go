// Package diario guarda o dia a dia de cada veículo: estatísticas e indicadores de condução da API Gobrax.
package diario

import (
	"context"

	"analisefrota/adapters/gobraxapi"
	"analisefrota/common/utils"
	"analisefrota/modules/diario/diario_repo"
)

type Service_I interface {
	// MontarDia junta estatísticas e performance de um dia (mesmas regras do painel); não grava nada
	MontarDia(clienteID int64, dia string, estat []gobraxapi.EstatisticaRec_t, perf []gobraxapi.PerformanceRec_t) []Registro_t
	// SalvarDia substitui o dia inteiro do cliente pelos registros dados
	SalvarDia(ctx context.Context, clienteID int64, dia string, regs []Registro_t) error
	Listar(ctx context.Context, clienteID int64, ini, fim string) ([]Registro_t, error)
	// DiasComIndicadores devolve os dias do período (AAAA-MM-DD) que já têm a performance gravada
	DiasComIndicadores(ctx context.Context, clienteID int64, ini, fim string) (map[string]bool, error)
}

type diario_t struct {
	repo diario_repo.Repository_I
}

func NewDiario() Service_I {
	return &diario_t{repo: diario_repo.NewRepository()}
}

// NewDiarioCom recebe o repositório (usado nos testes)
func NewDiarioCom(repo diario_repo.Repository_I) Service_I {
	return &diario_t{repo: repo}
}

// colunas de indicadores: campo do registro ← nome do indicador na API
var indicadores = []struct {
	nome  string
	campo func(r *Registro_t) **float64
}{
	{"greenRange", func(r *Registro_t) **float64 { return &r.FaixaVerde }},
	{"extraEconomicRange", func(r *Registro_t) **float64 { return &r.InicioFaixaVerde }},
	{"economicRange", func(r *Registro_t) **float64 { return &r.FinalFaixaVerde }},
	{"ecoRoll", func(r *Registro_t) **float64 { return &r.EcoRoll }},
	{"pedalPressureOnLow", func(r *Registro_t) **float64 { return &r.PressaoIdeal }},
	{"pedalPressureOnMid", func(r *Registro_t) **float64 { return &r.PressaoAtencao }},
	{"leverage", func(r *Registro_t) **float64 { return &r.Embalo }},
	{"idle", func(r *Registro_t) **float64 { return &r.MotorOcioso }},
	{"pedalPressureOnHigh", func(r *Registro_t) **float64 { return &r.PressaoAcelerador }},
	{"powerRange", func(r *Registro_t) **float64 { return &r.ExcessoRotacao }},
	{"speeding", func(r *Registro_t) **float64 { return &r.TempoExcessoVelocidade }},
	{"cruiseControl", func(r *Registro_t) **float64 { return &r.PilotoAutomatico }},
	{"engineBrake", func(r *Registro_t) **float64 { return &r.FreioMotor }},
}

func (d *diario_t) MontarDia(clienteID int64, dia string, estat []gobraxapi.EstatisticaRec_t, perf []gobraxapi.PerformanceRec_t) []Registro_t {
	perfPor := map[string]*gobraxapi.PerformanceRec_t{}
	for i := range perf {
		k := utils.ChavePlaca(perf[i].VehicleIdentification)
		if _, ja := perfPor[k]; k != "" && !ja {
			perfPor[k] = &perf[i]
		}
	}
	// sem estatísticas (token sem acesso ou falha), o dia sai só da performance: sem km e sem litros
	soPerf := len(estat) == 0
	base := estat
	if soPerf {
		base = make([]gobraxapi.EstatisticaRec_t, len(perf))
		for i, p := range perf {
			base[i] = gobraxapi.EstatisticaRec_t{VehicleIdentification: p.VehicleIdentification}
		}
	}
	vistos := map[string]bool{}
	regs := []Registro_t{}
	for _, e := range base {
		placa := utils.Placa(e.VehicleIdentification)
		k := utils.ChavePlaca(placa)
		if k == "" || vistos[k] {
			continue
		}
		vistos[k] = true
		pk := perfPor[k]
		var mov *float64
		if pk != nil {
			if m, ok := pk.Indicators["movement"]; ok {
				mov = m.Duration.Valor
			}
		}
		rodou := valorOu0(e.TotalMileage.Valor) > 0
		if soPerf {
			rodou = valorOu0(mov) > 0
		}
		r := Registro_t{ClienteID: clienteID, Dia: dia, Placa: placa, Rodou: rodou, Odometro: arred(e.Odometer.Valor)}
		if !soPerf {
			km := utils.Arred2(valorOu0(e.TotalMileage.Valor))
			r.Km = &km
		}
		if rodou {
			r.Litros, r.Frenagens, r.FrenagensAlta, r.VelocidadeMedia = arred(e.TotalConsumption.Valor), arred(e.TotalBreaking.Valor), arred(e.TotalBreakingOnHighSpeed.Valor), arred(e.AverageSpeed.Valor)
			if pk != nil {
				r.Motorista = motoristaPrincipal(pk)
				r.Nota = notaGeral(pk)
				if m, ok := pk.Indicators["movement"]; ok && m.Duration.Valor != nil {
					h := utils.Arred2(*m.Duration.Valor / 60) // a API manda minutos
					r.Horas = &h
				}
				for _, ind := range indicadores {
					*ind.campo(&r) = percentual(pk.Indicators, ind.nome)
				}
			}
		}
		regs = append(regs, r)
	}
	return regs
}

func (d *diario_t) SalvarDia(ctx context.Context, clienteID int64, dia string, regs []Registro_t) error {
	if err := validarPeriodo(dia, dia); err != nil {
		return err
	}
	rows := make([]diario_repo.RegistroRow_t, len(regs))
	for i, r := range regs {
		rows[i] = paraRow(r)
	}
	return d.repo.SubstituirDia(ctx, clienteID, dia, rows)
}

func (d *diario_t) Listar(ctx context.Context, clienteID int64, ini, fim string) ([]Registro_t, error) {
	if err := validarPeriodo(ini, fim); err != nil {
		return nil, err
	}
	rows, err := d.repo.Listar(ctx, clienteID, ini, fim)
	if err != nil {
		return nil, err
	}
	out := make([]Registro_t, len(rows))
	for i, r := range rows {
		out[i] = deRow(r)
	}
	return out, nil
}

func (d *diario_t) DiasComIndicadores(ctx context.Context, clienteID int64, ini, fim string) (map[string]bool, error) {
	if err := validarPeriodo(ini, fim); err != nil {
		return nil, err
	}
	dias, err := d.repo.DiasComIndicadores(ctx, clienteID, ini, fim)
	if err != nil {
		return nil, err
	}
	out := make(map[string]bool, len(dias))
	for _, dia := range dias {
		out[dia] = true
	}
	return out, nil
}

func paraRow(r Registro_t) diario_repo.RegistroRow_t {
	return diario_repo.RegistroRow_t{
		ClienteID: r.ClienteID, Dia: r.Dia, Placa: r.Placa, Motorista: r.Motorista, Rodou: r.Rodou,
		Km: r.Km, Litros: r.Litros, Frenagens: r.Frenagens, FrenagensAlta: r.FrenagensAlta, VelocidadeMedia: r.VelocidadeMedia,
		Odometro: r.Odometro, Horas: r.Horas, Nota: r.Nota, FaixaVerde: r.FaixaVerde, InicioFaixaVerde: r.InicioFaixaVerde,
		FinalFaixaVerde: r.FinalFaixaVerde, EcoRoll: r.EcoRoll, PressaoIdeal: r.PressaoIdeal, PressaoAtencao: r.PressaoAtencao,
		Embalo: r.Embalo, MotorOcioso: r.MotorOcioso, PressaoAcelerador: r.PressaoAcelerador, ExcessoRotacao: r.ExcessoRotacao,
		TempoExcessoVelocidade: r.TempoExcessoVelocidade, PilotoAutomatico: r.PilotoAutomatico, FreioMotor: r.FreioMotor,
	}
}

func deRow(r diario_repo.RegistroRow_t) Registro_t {
	dia := r.Dia
	if len(dia) > 10 {
		dia = dia[:10]
	}
	return Registro_t{
		ClienteID: r.ClienteID, Dia: dia, Placa: r.Placa, Motorista: r.Motorista, Rodou: r.Rodou,
		Km: r.Km, Litros: r.Litros, Frenagens: r.Frenagens, FrenagensAlta: r.FrenagensAlta, VelocidadeMedia: r.VelocidadeMedia,
		Odometro: r.Odometro, Horas: r.Horas, Nota: r.Nota, FaixaVerde: r.FaixaVerde, InicioFaixaVerde: r.InicioFaixaVerde,
		FinalFaixaVerde: r.FinalFaixaVerde, EcoRoll: r.EcoRoll, PressaoIdeal: r.PressaoIdeal, PressaoAtencao: r.PressaoAtencao,
		Embalo: r.Embalo, MotorOcioso: r.MotorOcioso, PressaoAcelerador: r.PressaoAcelerador, ExcessoRotacao: r.ExcessoRotacao,
		TempoExcessoVelocidade: r.TempoExcessoVelocidade, PilotoAutomatico: r.PilotoAutomatico, FreioMotor: r.FreioMotor,
	}
}

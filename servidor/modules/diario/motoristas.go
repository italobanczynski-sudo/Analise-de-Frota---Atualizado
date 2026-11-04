package diario

import (
	"context"
	"strings"
	"time"

	"analisefrota/adapters/gobraxapi"
	"analisefrota/common/utils"
)

// Motorista não é veículo: o dado do condutor é só o do tempo em que ele estava vinculado ao veículo (como a análise
// por motorista da plataforma e o painel, index.html: vinculos, diaParaCSV e juntarJanelas). drivers[] da performance
// traz o início e o fim de cada vínculo, recortados na janela do dia. Um único motorista vinculado o dia todo usa a
// linha do próprio veículo; nos demais casos a coleta consulta estatísticas e performance só nos trechos daquele
// motorista (vehicleIdentification + início e fim do vínculo) e MontarMotorista junta os trechos.

// tolerância nas pontas do dia para considerar que o motorista ficou o dia todo no veículo (igual ao painel)
const folgaDiaTodo = time.Minute

// PlanoMotoristas monta, de um dia já montado (MontarDia), as linhas dos motoristas que ficaram o dia todo no veículo e
// a lista dos vínculos parciais que ainda precisam ser consultados. Não chama a API nem grava nada.
func (d *diario_t) PlanoMotoristas(dia string, regs []Registro_t, perf []gobraxapi.PerformanceRec_t) PlanoMotoristas_t {
	plano := PlanoMotoristas_t{Inteiros: []Registro_t{}, Pendentes: []Pendente_t{}}
	ji, jf, err := utils.JanelaDiaBR(dia)
	if err != nil {
		return plano
	}
	perfPor := map[string]*gobraxapi.PerformanceRec_t{}
	for i := range perf {
		k := utils.ChavePlaca(perf[i].VehicleIdentification)
		if _, ja := perfPor[k]; k != "" && !ja {
			perfPor[k] = &perf[i]
		}
	}
	for _, r := range regs {
		if !r.Rodou { // sem km, o dia fica só na planilha de veículos
			continue
		}
		p := perfPor[utils.ChavePlaca(r.Placa)]
		if p == nil {
			continue
		}
		lista := vinculos(p, ji, jf)
		if diaTodo(lista, ji, jf) {
			m := r
			m.Motorista = lista[0].Motorista
			plano.Inteiros = append(plano.Inteiros, m)
			continue
		}
		for _, v := range lista {
			plano.Pendentes = append(plano.Pendentes, Pendente_t{Placa: r.Placa, Motorista: v.Motorista, Janelas: v.Janelas})
		}
	}
	return plano
}

// vinculo_t são os trechos de um motorista no veículo, na ordem em que aparecem em drivers[]
type vinculo_t struct {
	Motorista string
	Janelas   []Janela_t
}

// vinculos agrupa drivers[] por nome, com cada trecho recortado na janela do dia. Início ou fim ilegível vale a ponta da janela.
func vinculos(p *gobraxapi.PerformanceRec_t, ji, jf time.Time) []vinculo_t {
	ordem := []string{}
	por := map[string][]Janela_t{}
	for _, d := range p.Drivers {
		nome := strings.TrimSpace(d.DriverName)
		if nome == "" {
			continue
		}
		a, b := ji, jf
		if t, ok := utils.InstanteAPI(d.StartDate); ok && t.After(ji) {
			a = t
		}
		if t, ok := utils.InstanteAPI(d.EndDate); ok && t.Before(jf) {
			b = t
		}
		if !b.After(a) {
			continue
		}
		if _, ja := por[nome]; !ja {
			ordem = append(ordem, nome)
		}
		por[nome] = append(por[nome], Janela_t{Ini: a, Fim: b})
	}
	out := make([]vinculo_t, 0, len(ordem))
	for _, n := range ordem {
		out = append(out, vinculo_t{Motorista: n, Janelas: por[n]})
	}
	return out
}

// diaTodo: um motorista só, com um trecho que cobre o dia inteiro (com folga de 1 min nas pontas)
func diaTodo(lista []vinculo_t, ji, jf time.Time) bool {
	if len(lista) != 1 {
		return false
	}
	for _, j := range lista[0].Janelas {
		if j.Ini.Sub(ji) <= folgaDiaTodo && jf.Sub(j.Fim) <= folgaDiaTodo {
			return true
		}
	}
	return false
}

// bases_t são as bases (em minutos) de cada percentual da performance: o tempo em movimento na maioria; a ignição
// ligada no motor ligado parado (movement.percentage é movimento ÷ ignição); o pé no acelerador nos três ranges do pedal
type bases_t struct{ mov, motor, pedal float64 }

var rangesPedal = []string{"pedalPressureOnLow", "pedalPressureOnMid", "pedalPressureOnHigh"}

// indicador devolve o indicador pelo nome; a API escreve "pedalPressureOnHig" (sem o h final) em parte das respostas
func indicador(ind map[string]gobraxapi.Indicador_t, nome string) (gobraxapi.Indicador_t, bool) {
	x, ok := ind[nome]
	if !ok && nome == "pedalPressureOnHigh" {
		x, ok = ind["pedalPressureOnHig"]
	}
	return x, ok
}

func basesDe(ind map[string]gobraxapi.Indicador_t) bases_t {
	var b bases_t
	if m, ok := ind["movement"]; ok {
		b.mov = valorOu0(m.Duration.Valor)
		if pc := valorOu0(m.Percentage.Valor); b.mov > 0 && pc > 0 {
			b.motor = b.mov / (pc / 100)
		}
	}
	for _, k := range rangesPedal {
		if x, ok := indicador(ind, k); ok {
			b.pedal += valorOu0(x.Duration.Valor)
		}
	}
	return b
}

// baseDo: a base que pondera cada indicador ao juntar trechos
func (b bases_t) baseDo(nome string) float64 {
	switch nome {
	case "idle":
		return b.motor
	case "pedalPressureOnLow", "pedalPressureOnMid", "pedalPressureOnHigh":
		return b.pedal
	}
	return b.mov
}

// MontarMotorista junta os trechos de um vínculo parcial (estatísticas e performance de cada trecho): km, litros e
// frenagens somados, velocidade média ponderada pelos km, maior odômetro e cada indicador como a média dos percentuais
// da API ponderada pela base dele em cada trecho. ok = false quando o motorista não rodou nesses trechos.
func (d *diario_t) MontarMotorista(clienteID int64, dia string, p Pendente_t, partes []Parte_t) (Registro_t, bool) {
	chave := utils.ChavePlaca(p.Placa)
	daPlaca := func(v string) bool { k := utils.ChavePlaca(v); return k == "" || k == chave }
	var km, litros, fren, frenAlta, kmVel, velSoma float64
	var odo *float64
	var tot bases_t
	soma, peso := map[string]float64{}, map[string]float64{}
	for _, pt := range partes {
		for _, e := range pt.Estat {
			if !daPlaca(e.VehicleIdentification) {
				continue
			}
			k := valorOu0(e.TotalMileage.Valor)
			km += k
			litros += valorOu0(e.TotalConsumption.Valor)
			fren += valorOu0(e.TotalBreaking.Valor)
			frenAlta += valorOu0(e.TotalBreakingOnHighSpeed.Valor)
			if v := e.AverageSpeed.Valor; v != nil && k > 0 {
				velSoma += *v * k
				kmVel += k
			}
			if o := e.Odometer.Valor; o != nil && (odo == nil || *o > *odo) {
				x := *o
				odo = &x
			}
		}
		for _, r := range pt.Perf {
			if !daPlaca(r.VehicleIdentification) {
				continue
			}
			b := basesDe(r.Indicators)
			if b.mov <= 0 {
				continue
			}
			tot.mov, tot.motor, tot.pedal = tot.mov+b.mov, tot.motor+b.motor, tot.pedal+b.pedal
			for _, ind := range indicadores {
				x, ok := indicador(r.Indicators, ind.nome)
				w := b.baseDo(ind.nome)
				if !ok || x.Percentage.Valor == nil || w <= 0 {
					continue
				}
				soma[ind.nome] += *x.Percentage.Valor * w
				peso[ind.nome] += w
			}
		}
	}
	if km <= 0 {
		return Registro_t{}, false
	}
	r := Registro_t{ClienteID: clienteID, Dia: dia, Placa: p.Placa, Motorista: p.Motorista, Rodou: true, Odometro: arred(odo)}
	r.Km, r.Litros, r.Frenagens, r.FrenagensAlta = ptr(km), ptr(litros), ptr(fren), ptr(frenAlta)
	if kmVel > 0 {
		r.VelocidadeMedia = ptr(velSoma / kmVel)
	}
	r.Horas, r.HorasMotor, r.HorasPedal = horasOuNil(tot.mov), horasOuNil(tot.motor), horasOuNil(tot.pedal) // a API manda minutos
	for _, ind := range indicadores {
		if peso[ind.nome] > 0 {
			*ind.campo(&r) = ptr(soma[ind.nome] / peso[ind.nome])
		}
	}
	return r, true
}

func (d *diario_t) SalvarMotoristasDia(ctx context.Context, clienteID int64, dia string, regs []Registro_t) error {
	if err := validarPeriodo(dia, dia); err != nil {
		return err
	}
	for _, r := range regs {
		if strings.TrimSpace(r.Motorista) == "" {
			return ErrSemMotorista
		}
	}
	return d.repo.SubstituirMotoristasDia(ctx, clienteID, dia, paraRows(regs))
}

func (d *diario_t) ListarMotoristas(ctx context.Context, clienteID int64, ini, fim string) ([]Registro_t, error) {
	if err := validarPeriodo(ini, fim); err != nil {
		return nil, err
	}
	rows, err := d.repo.ListarMotoristas(ctx, clienteID, ini, fim)
	if err != nil {
		return nil, err
	}
	out := make([]Registro_t, len(rows))
	for i, r := range rows {
		out[i] = deRow(r)
	}
	return out, nil
}

// ptr arredonda em 2 casas e devolve o ponteiro
func ptr(v float64) *float64 {
	x := utils.Arred2(v)
	return &x
}

func horasOuNil(minutos float64) *float64 {
	if minutos <= 0 {
		return nil
	}
	return ptr(minutos / 60)
}

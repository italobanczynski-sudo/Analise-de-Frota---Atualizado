package mensagem

import (
	"context"
	"sort"

	"analisefrota/common/utils"
	"analisefrota/modules/diario"
)

// DiasDepois: depois de uma mensagem, o ponto apontado é medido nos dias seguintes (até este tanto)
const DiasDepois = 7

const (
	ResultadoMelhorou    = "melhorou"
	ResultadoNaoMelhorou = "nao_melhorou"
	ResultadoAguardando  = "aguardando" // ainda não rodou depois da mensagem
	ResultadoSemPonto    = "sem_ponto"  // no dia, nenhum indicador piorou o bastante para ser apontado
)

func indicadorPorChave(chave string) (indicador_t, bool) {
	for _, ind := range indicadores {
		if ind.chave == chave {
			return ind, true
		}
	}
	return indicador_t{}, false
}

func (m *mensagem_t) Acompanhamento(ctx context.Context, clienteID int64, motorista string) (Acompanhamento_t, error) {
	rows, err := m.d.Repo.ListarMotorista(ctx, clienteID, motorista)
	if err != nil {
		return Acompanhamento_t{}, err
	}
	out := Acompanhamento_t{Motorista: motorista, DiasEnviada: len(rows), Envios: []Envio_t{}, Focos: []FocoTotal_t{}}
	if len(rows) == 0 {
		return out, nil
	}

	// dia a dia do motorista, do primeiro envio até DiasDepois depois do último (no máximo até ontem)
	ontem := utils.SomarDias(utils.DiaBR(m.d.Agora()), -1)
	ini, ultimo := rows[0].Dia, rows[0].Dia
	for _, e := range rows {
		if e.Dia < ini {
			ini = e.Dia
		}
		if e.Dia > ultimo {
			ultimo = e.Dia
		}
	}
	fim := utils.SomarDias(ultimo, DiasDepois)
	if fim > ontem {
		fim = ontem
	}
	if limite := utils.SomarDias(fim, -399); ini < limite { // o diário aceita até 400 dias por consulta
		ini = limite
	}
	var dele []diario.Registro_t
	if ini <= fim {
		regs, err := m.d.Diario.Listar(ctx, clienteID, ini, fim)
		if err != nil {
			return out, err
		}
		chave := ChaveMotorista(motorista)
		for _, r := range regs {
			if r.Rodou && ChaveMotorista(r.Motorista) == chave {
				dele = append(dele, r)
			}
		}
	}

	totais := map[string]*FocoTotal_t{}
	for _, e := range rows {
		env := Envio_t{Dia: e.Dia, Foco: e.Foco, FocoValor: e.FocoValor, FocoMedia: e.FocoMedia, Melhorou: e.Melhorou, Piorou: e.Piorou,
			Comparados: e.Comparados, Nota: e.Nota, MarcadoEm: e.MarcadoEm, Resultado: ResultadoSemPonto}
		if ind, ok := indicadorPorChave(e.Foco); ok {
			env.FocoRotulo = ind.rotulo
			ate := utils.SomarDias(e.Dia, DiasDepois)
			depois, dias := []diario.Registro_t{}, map[string]bool{}
			for _, r := range dele {
				if r.Dia > e.Dia && r.Dia <= ate {
					depois = append(depois, r)
					dias[r.Dia] = true
				}
			}
			env.Depois, env.DiasDepois = agregar(depois, ind), len(dias)
			env.Resultado = resultado(ind, env.FocoValor, env.Depois)

			t := totais[e.Foco]
			if t == nil {
				t = &FocoTotal_t{Foco: e.Foco, Rotulo: ind.rotulo}
				totais[e.Foco] = t
			}
			t.Vezes++
			switch env.Resultado {
			case ResultadoMelhorou:
				t.Melhorou++
			case ResultadoAguardando:
				t.Aguardando++
			}
		}
		out.Envios = append(out.Envios, env)
	}
	// o mais recente primeiro; os pontos mais apontados primeiro
	sort.Slice(out.Envios, func(i, j int) bool { return out.Envios[i].Dia > out.Envios[j].Dia })
	for _, t := range totais {
		out.Focos = append(out.Focos, *t)
	}
	sort.Slice(out.Focos, func(i, j int) bool {
		if out.Focos[i].Vezes != out.Focos[j].Vezes {
			return out.Focos[i].Vezes > out.Focos[j].Vezes
		}
		return out.Focos[i].Rotulo < out.Focos[j].Rotulo
	})
	return out, nil
}

// resultado: melhorou se, nos dias seguintes, o ponto apontado ficou melhor que no dia da mensagem (pelo menos o "estável" do indicador)
func resultado(ind indicador_t, noDia, depois *float64) string {
	switch {
	case depois == nil:
		return ResultadoAguardando
	case noDia == nil:
		return ResultadoNaoMelhorou
	case (*depois-*noDia)*float64(ind.melhor) >= ind.estavel:
		return ResultadoMelhorou
	}
	return ResultadoNaoMelhorou
}

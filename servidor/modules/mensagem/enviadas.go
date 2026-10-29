package mensagem

import (
	"context"
	"sort"
	"strings"
)

// rotuloIndicador: o nome da plataforma para a chave do diário (vazio = chave desconhecida)
func rotuloIndicador(chave string) string {
	if chave == kml.chave {
		return kml.rotulo
	}
	if ind, ok := indicadorPorChave(chave); ok {
		return ind.rotulo
	}
	return ""
}

// Enviadas junta as mensagens marcadas como enviadas por motorista. Em cada dia, os indicadores a melhorar são o
// ponto da dica e os que pioraram contra o dia anterior; cada um conta uma vez por dia.
func (m *mensagem_t) Enviadas(ctx context.Context, clienteID int64) ([]Enviadas_t, error) {
	rows, err := m.d.Repo.ListarCliente(ctx, clienteID)
	if err != nil {
		return nil, err
	}
	por, vezes := map[string]*Enviadas_t{}, map[string]map[string]*IndicadorVez_t{}
	ultimo := map[string]string{}
	for _, e := range rows {
		k := ChaveMotorista(e.Motorista)
		out := por[k]
		if out == nil {
			out = &Enviadas_t{Motorista: e.Motorista, Indicadores: []IndicadorVez_t{}, Envios: []EnviadaNota_t{}}
			por[k], vezes[k] = out, map[string]*IndicadorVez_t{}
		}
		if e.Dia >= ultimo[k] { // o nome como veio na mensagem mais recente
			out.Motorista, ultimo[k] = e.Motorista, e.Dia
		}
		env := EnviadaNota_t{Dia: e.Dia, Nota: e.Nota, Foco: e.Foco, FocoRotulo: rotuloIndicador(e.Foco), Indicadores: []string{}}
		chaves := append([]string{e.Foco}, strings.Split(e.Pioraram, ",")...)
		visto := map[string]bool{}
		for _, c := range chaves {
			rot := rotuloIndicador(c)
			if rot == "" || visto[c] {
				continue
			}
			visto[c] = true
			env.Indicadores = append(env.Indicadores, rot)
			v := vezes[k][c]
			if v == nil {
				v = &IndicadorVez_t{Chave: c, Rotulo: rot}
				vezes[k][c] = v
			}
			v.Vezes++
		}
		out.Envios = append(out.Envios, env)
		out.DiasEnviada++
	}

	lista := make([]Enviadas_t, 0, len(por))
	for k, out := range por {
		sort.Slice(out.Envios, func(i, j int) bool { return out.Envios[i].Dia > out.Envios[j].Dia })
		for _, v := range vezes[k] {
			out.Indicadores = append(out.Indicadores, *v)
		}
		sort.Slice(out.Indicadores, func(i, j int) bool {
			if out.Indicadores[i].Vezes != out.Indicadores[j].Vezes {
				return out.Indicadores[i].Vezes > out.Indicadores[j].Vezes
			}
			return out.Indicadores[i].Rotulo < out.Indicadores[j].Rotulo
		})
		lista = append(lista, *out)
	}
	// quem recebeu mais mensagens primeiro
	sort.Slice(lista, func(i, j int) bool {
		if lista[i].DiasEnviada != lista[j].DiasEnviada {
			return lista[i].DiasEnviada > lista[j].DiasEnviada
		}
		return lista[i].Motorista < lista[j].Motorista
	})
	return lista, nil
}

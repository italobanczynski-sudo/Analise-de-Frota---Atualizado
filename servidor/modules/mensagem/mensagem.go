// Package mensagem monta, para cada motorista que rodou num dia, a análise individual pronta para mandar
// pelo WhatsApp: km, média, início da faixa verde, embalo, motor ligado parado, piloto automático, pedal crítico
// (cada um contra a média dos 7 dias anteriores), nota do mês contra o mês anterior, evolução e dica do dia.
// Lê o dia a dia (diario) e a nota (premiacao); não envia nada nem guarda telefone: quem manda é a pessoa,
// que marca a mensagem como enviada. Cada marca guarda um retrato da análise para o acompanhamento do motorista.
package mensagem

import (
	"context"
	"errors"
	"strings"
	"time"

	"analisefrota/common/utils"
	"analisefrota/modules/diario"
	"analisefrota/modules/mensagem/mensagem_repo"
	"analisefrota/modules/premiacao"
)

type Service_I interface {
	// Listar monta as mensagens do dia (AAAA-MM-DD, até ontem) de cada motorista que rodou
	Listar(ctx context.Context, clienteID int64, dia string) ([]Mensagem_t, error)
	// Marcar registra que a mensagem do motorista no dia foi enviada (guarda o retrato da análise)
	Marcar(ctx context.Context, clienteID int64, dia, motorista string, usuarioID int64) (Mensagem_t, error)
	Desmarcar(ctx context.Context, clienteID int64, dia, motorista string) error
	// Acompanhamento: dias com mensagem enviada e se o motorista melhorou depois no ponto apontado
	Acompanhamento(ctx context.Context, clienteID int64, motorista string) (Acompanhamento_t, error)
	// Enviadas: por motorista, quantas mensagens foram enviadas, a nota em cada dia e os indicadores apontados
	Enviadas(ctx context.Context, clienteID int64) ([]Enviadas_t, error)
}

// Deps_t são as dependências; os testes trocam cada uma por um mock
type Deps_t struct {
	Repo      mensagem_repo.Repository_I
	Diario    diario.Service_I
	Premiacao premiacao.Service_I
	Agora     func() time.Time
}

type mensagem_t struct{ d Deps_t }

func NewMensagem() Service_I {
	return &mensagem_t{d: Deps_t{Repo: mensagem_repo.NewRepository(), Diario: diario.NewDiario(), Premiacao: premiacao.NewPremiacao(), Agora: time.Now}}
}

// NewMensagemCom recebe as dependências (usado nos testes)
func NewMensagemCom(d Deps_t) Service_I { return &mensagem_t{d: d} }

var (
	ErrDiaFuturo       = errors.New("a análise só pode ser de um dia que já terminou (até ontem)")
	ErrMotoristaSemDia = errors.New("este motorista não rodou neste dia; não há mensagem para marcar")
	ErrNaoMarcada      = errors.New("esta mensagem não estava marcada como enviada")
)

func (m *mensagem_t) Listar(ctx context.Context, clienteID int64, dia string) ([]Mensagem_t, error) {
	res, err := m.resumos(ctx, clienteID, dia)
	if err != nil {
		return nil, err
	}
	marcadas, err := m.d.Repo.ListarDia(ctx, clienteID, dia)
	if err != nil {
		return nil, err
	}
	quando := map[string]time.Time{}
	for _, e := range marcadas {
		quando[ChaveMotorista(e.Motorista)] = e.MarcadoEm
	}
	out := make([]Mensagem_t, len(res))
	for i, r := range res {
		out[i] = Mensagem_t{Resumo_t: r, Texto: Texto(r.Variaveis)}
		if t, ok := quando[ChaveMotorista(r.Motorista)]; ok {
			out[i].Enviada = &t
		}
	}
	return out, nil
}

func (m *mensagem_t) Marcar(ctx context.Context, clienteID int64, dia, motorista string, usuarioID int64) (Mensagem_t, error) {
	res, err := m.resumos(ctx, clienteID, dia)
	if err != nil {
		return Mensagem_t{}, err
	}
	for _, r := range res {
		if ChaveMotorista(r.Motorista) != ChaveMotorista(motorista) {
			continue
		}
		row := mensagem_repo.EnviadaRow_t{ClienteID: clienteID, Motorista: r.Motorista, Dia: dia, Foco: r.Foco, FocoValor: r.FocoValor,
			FocoMedia: r.FocoMedia, Melhorou: r.Melhorou, Piorou: r.Piorou, Comparados: r.Comparados, Nota: r.Nota,
			Pioraram: strings.Join(r.Pioraram, ",")}
		if usuarioID > 0 {
			row.MarcadoPor = &usuarioID
		}
		if err := m.d.Repo.Marcar(ctx, row); err != nil {
			return Mensagem_t{}, err
		}
		agora := m.d.Agora()
		return Mensagem_t{Resumo_t: r, Texto: Texto(r.Variaveis), Enviada: &agora}, nil
	}
	return Mensagem_t{}, ErrMotoristaSemDia
}

func (m *mensagem_t) Desmarcar(ctx context.Context, clienteID int64, dia, motorista string) error {
	ok, err := m.d.Repo.Desmarcar(ctx, clienteID, motorista, dia)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNaoMarcada
	}
	return nil
}

func (m *mensagem_t) resumos(ctx context.Context, clienteID int64, dia string) ([]Resumo_t, error) {
	if dia >= utils.DiaBR(m.d.Agora()) {
		return nil, ErrDiaFuturo
	}
	regs, err := m.d.Diario.Listar(ctx, clienteID, utils.SomarDias(dia, -DiasComparacao), dia)
	if err != nil {
		return nil, err
	}
	notas, err := m.notas(ctx, clienteID, dia)
	if err != nil {
		return nil, err
	}
	return MontarResumos(dia, regs, notas), nil
}

// notas: Nota Gobrax do mês do dia e do mês anterior, por motorista (premiação da API)
func (m *mensagem_t) notas(ctx context.Context, clienteID int64, dia string) (map[string]Nota_t, error) {
	mes := dia[:7]
	atual, err := m.d.Premiacao.Listar(ctx, clienteID, mes)
	if err != nil {
		return nil, err
	}
	anterior, err := m.d.Premiacao.Listar(ctx, clienteID, MesAnterior(mes))
	if err != nil {
		return nil, err
	}
	parcial := mes == utils.DiaBR(m.d.Agora())[:7]
	out := map[string]Nota_t{}
	for _, p := range atual {
		out[ChaveMotorista(p.Motorista)] = Nota_t{Mes: mes, Atual: p.Nota, Parcial: parcial}
	}
	for _, p := range anterior {
		k := ChaveMotorista(p.Motorista)
		if n, ok := out[k]; ok {
			n.Anterior = p.Nota
			out[k] = n
		}
	}
	return out, nil
}

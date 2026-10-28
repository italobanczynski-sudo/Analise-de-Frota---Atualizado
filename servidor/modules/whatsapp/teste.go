package whatsapp

import (
	"context"
	"errors"
	"strconv"
	"time"

	"analisefrota/adapters/twilio"
	"analisefrota/modules/contato"
)

// Teste_t é o resultado de um envio de teste: a situação final na Twilio e, se falhou, o motivo
type Teste_t struct {
	Situacao string `json:"situacao"` // delivered, read, sent, queued, failed, undelivered
	Entregue bool   `json:"entregue"`
	Codigo   int    `json:"codigo"`
	Motivo   string `json:"motivo"`
	Texto    string `json:"texto"`
}

var (
	ErrNaoConectado    = errors.New("a Twilio não está conectada: preencha e salve a Conexão com a Twilio")
	ErrMotoristaSemDia = errors.New("este motorista não rodou no dia escolhido: escolha outro motorista ou outro dia (veja quem aparece em Ver prévia)")
)

// esperas entre as consultas da situação na Twilio (cerca de 15 s no total)
var esperasStatus = []time.Duration{time.Second, 2 * time.Second, 3 * time.Second, 4 * time.Second, 5 * time.Second}

// EnviarTeste manda a mensagem do motorista no dia para um número qualquer (não salvo, não conta como envio do dia)
// e espera a Twilio dizer se entregou
func (w *whatsapp_t) EnviarTeste(ctx context.Context, clienteID int64, dia, motorista, telefone string) (Teste_t, error) {
	tel, err := contato.NormalizarTelefone(telefone)
	if err != nil {
		return Teste_t{}, err
	}
	cfg, err := w.efetiva(ctx)
	if err != nil {
		return Teste_t{}, err
	}
	if !cfg.Configurado() {
		return Teste_t{}, ErrNaoConectado
	}
	res, err := w.resumos(ctx, clienteID, dia)
	if err != nil {
		return Teste_t{}, err
	}
	var r *Resumo_t
	for i := range res {
		if ChaveMotorista(res[i].Motorista) == ChaveMotorista(motorista) {
			r = &res[i]
		}
	}
	if r == nil {
		return Teste_t{}, ErrMotoristaSemDia
	}
	envio := w.d.Twilio(cfg)
	out := Teste_t{Texto: Texto(r.Variaveis)}
	sid, err := envio.Enviar(ctx, tel, r.Variaveis, out.Texto)
	if err != nil {
		return out, err
	}
	out.Situacao = "queued"
	for _, espera := range esperasStatus {
		select {
		case <-ctx.Done():
			return out, nil
		case <-time.After(espera):
		}
		st, codigo, err := envio.Status(ctx, sid)
		if err != nil {
			break
		}
		out.Situacao, out.Codigo = st, codigo
		if st == "delivered" || st == "read" || st == "failed" || st == "undelivered" {
			break
		}
	}
	out.Entregue = out.Situacao == "delivered" || out.Situacao == "read"
	if out.Situacao == "failed" || out.Situacao == "undelivered" {
		out.Motivo = "A Twilio não conseguiu entregar (código " + itoa(out.Codigo) + ")" + twilio.Dica(out.Codigo)
	}
	return out, nil
}

func itoa(n int) string {
	if n == 0 {
		return "sem código"
	}
	return strconv.Itoa(n)
}

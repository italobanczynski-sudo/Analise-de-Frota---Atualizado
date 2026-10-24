package cliente

import (
	"context"

	"analisefrota/api/middleware"
	modcliente "analisefrota/modules/cliente"
	"analisefrota/modules/coleta"
)

// Listar devolve os clientes com a situação da última busca de cada um
func (m Cliente_t) Listar(ctx context.Context) *middleware.Response_t {
	clis, err := modcliente.NewCliente().Listar(ctx)
	if err != nil {
		return middleware.RetErro(err)
	}
	col := coleta.NewColeta()
	out := make([]ClienteComColeta_t, len(clis))
	for i, c := range clis {
		out[i] = ClienteComColeta_t{Cliente_t: c}
		if s, err := col.Situacao(ctx, c.ID); err == nil {
			out[i].Coleta = Coleta_t{Situacao: s.Situacao, UltimoDia: s.UltimoDia, TerminadaEm: s.TerminadaEm, Mensagem: s.Mensagem}
		}
	}
	return middleware.RetOkData(out)
}

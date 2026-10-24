package coleta

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"

	"analisefrota/api/middleware"
	modcliente "analisefrota/modules/cliente"
	modcoleta "analisefrota/modules/coleta"
)

// Iniciar começa a busca do cliente em segundo plano e responde na hora (202); a situação sai em GET .../coleta
func (m Coleta_t) Iniciar(ctx context.Context) *middleware.Response_t {
	c, err := modcliente.NewCliente().Buscar(ctx, m.ClienteID)
	if err != nil {
		return middleware.RetErro(err)
	}
	if !c.TemToken {
		return middleware.RetErro(modcliente.ErrSemToken)
	}
	col := modcoleta.NewColeta()
	if s, err := col.Situacao(ctx, m.ClienteID); err == nil && s.Situacao == modcoleta.SituacaoRodando && s.IniciadaEm != nil && time.Since(*s.IniciadaEm) < 30*time.Minute {
		return middleware.RetErro(modcoleta.ErrEmAndamento)
	}
	go func(id int64, nome string) {
		// a busca segue mesmo depois que a requisição termina
		bg, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		if _, err := col.Executar(bg, id); err != nil {
			logrus.Warnf("coleta manual: %s: %v", nome, err)
		}
	}(c.ID, c.Nome)
	return middleware.RetAceito(map[string]string{"situacao": modcoleta.SituacaoRodando})
}

// Servidor do painel Análise de Frota: busca os dados da API Gobrax, guarda no Postgres e os entrega pela API HTTP.
package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/sirupsen/logrus"

	"analisefrota/adapters/dbPostgres"
	applog "analisefrota/adapters/log"
	"analisefrota/adapters/migrations"
	"analisefrota/api"
	mwcliente "analisefrota/api/middleware/cliente"
	"analisefrota/background"
	"analisefrota/environment"
)

func main() {
	importar := flag.String("importar-clientes", "", "caminho do api/clientes.json do script antigo: cadastra os clientes e sai")
	tokens := flag.String("tokens", "", "caminho do api/.env com os GOBRAX_TOKEN_* (usado com -importar-clientes)")
	flag.Parse()

	environment.Load()
	applog.Init()
	if err := migrations.Executar(dbPostgres.GetPgConn()); err != nil {
		logrus.Fatalf("banco: %v", err)
	}

	if *importar != "" {
		feito, err := mwcliente.Importar(context.Background(), *importar, *tokens)
		if err != nil {
			logrus.Fatalf("importar: %v", err)
		}
		for _, f := range feito {
			fmt.Println(f)
		}
		return
	}

	background.Init()
	api.NewService().Start()
}

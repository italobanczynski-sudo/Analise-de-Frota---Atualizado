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
	mwusuario "analisefrota/api/middleware/usuario"
	"analisefrota/background"
	"analisefrota/environment"
)

func main() {
	importar := flag.String("importar-clientes", "", "caminho do api/clientes.json do script antigo: cadastra os clientes e sai")
	tokens := flag.String("tokens", "", "caminho do api/.env com os GOBRAX_TOKEN_* (usado com -importar-clientes)")
	adminNome := flag.String("criar-admin", "", "nome do administrador a criar (com -email): mostra o link de convite e sai")
	adminEmail := flag.String("email", "", "e-mail do administrador (usado com -criar-admin)")
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

	if *adminNome != "" {
		link, err := mwusuario.CriarAdmin(context.Background(), *adminNome, *adminEmail)
		if err != nil {
			logrus.Fatalf("criar administrador: %v", err)
		}
		fmt.Println("Administrador criado. Abra este link para criar a senha (vale uma vez só):")
		fmt.Println(link)
		return
	}

	background.Init()
	api.NewService().Start()
}

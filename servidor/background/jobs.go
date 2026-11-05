// Package background roda as tarefas automáticas: a busca de dados (logo que o servidor sobe e depois
// a cada COLETA_INTERVALO_MIN), o rastro da frota ao vivo (a cada RASTRO_INTERVALO_MIN) e a limpeza das sessões de
// login vencidas e do rastro antigo (a cada 6 horas).
package background

import (
	"context"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/sirupsen/logrus"

	"analisefrota/common/utils"
	"analisefrota/environment"
	modcliente "analisefrota/modules/cliente"
	"analisefrota/modules/coleta"
	"analisefrota/modules/diario"
	"analisefrota/modules/sessao"
	"analisefrota/modules/trajeto"
)

func Init() {
	s, err := gocron.NewScheduler()
	if err != nil {
		logrus.Fatalf("background: agendador: %v", err)
	}
	unica := gocron.WithSingletonMode(gocron.LimitModeReschedule) // uma rodada de cada vez

	// o servidor acabou de subir: busca que ficou "rodando" foi cortada no meio (reinício, hospedagem dormindo)
	if n, err := coleta.NewColeta().MarcarInterrompidas(context.Background()); err != nil {
		logrus.Warnf("background: marcar buscas interrompidas: %v", err)
	} else if n > 0 {
		logrus.Infof("background: %d busca(s) interrompida(s) por reinício marcada(s) como erro", n)
	}

	cfg := environment.Cfg.Coleta
	if cfg.Automatica {
		if _, err := s.NewJob(gocron.DurationJob(time.Duration(cfg.IntervaloMin)*time.Minute), gocron.NewTask(buscarTodos),
			gocron.WithStartAt(gocron.WithStartImmediately()), unica); err != nil {
			logrus.Fatalf("background: tarefa de busca: %v", err)
		}
		logrus.Infof("background: busca automática a cada %d min", cfg.IntervaloMin)
	} else {
		logrus.Info("background: busca automática desligada (COLETA_AUTOMATICA=false)")
	}

	if r := environment.Cfg.Rastro; r.IntervaloMin > 0 {
		if _, err := s.NewJob(gocron.DurationJob(time.Duration(r.IntervaloMin)*time.Minute), gocron.NewTask(rastrearTodos),
			gocron.WithStartAt(gocron.WithStartImmediately()), unica); err != nil {
			logrus.Fatalf("background: tarefa do rastro: %v", err)
		}
		logrus.Infof("background: rastro da frota a cada %d min (guardado por %d dias)", r.IntervaloMin, r.Dias)
	} else {
		logrus.Info("background: rastro da frota desligado (RASTRO_INTERVALO_MIN=0)")
	}

	if _, err := s.NewJob(gocron.DurationJob(6*time.Hour), gocron.NewTask(limparSessoes), gocron.WithStartAt(gocron.WithStartImmediately()), unica); err != nil {
		logrus.Fatalf("background: limpeza de sessões: %v", err)
	}
	s.Start()
}

func buscarTodos() {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Minute)
	defer cancel()
	coleta.NewColeta().ExecutarTodos(ctx)
}

func limparSessoes() {
	limparRastro()
	n, err := sessao.NewSessao().LimparVencidas(context.Background())
	if err != nil {
		logrus.Warnf("background: limpeza de sessões: %v", err)
	} else if n > 0 {
		logrus.Infof("background: %d sessão(ões) vencida(s) apagada(s)", n)
	}
}

// rastrearTodos grava a posição atual de cada veículo de cada cliente com token (as placas que rodaram nos últimos
// 30 dias). Um cliente com erro (token vencido, API fora) não impede os outros.
func rastrearTodos() {
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
	defer cancel()
	cli := modcliente.NewCliente()
	cs, err := cli.Listar(ctx)
	if err != nil {
		logrus.Warnf("background: rastro: clientes: %v", err)
		return
	}
	hoje := utils.DiaBR(time.Now())
	for _, c := range cs {
		if !c.TemToken {
			continue
		}
		token, err := cli.Token(ctx, c.ID)
		if err != nil {
			continue
		}
		regs, err := diario.NewDiario().Listar(ctx, c.ID, utils.SomarDias(hoje, -30), hoje)
		if err != nil {
			logrus.Warnf("background: rastro de %s: registros: %v", c.Nome, err)
			continue
		}
		vistas, placas := map[string]bool{}, []string{}
		for _, r := range regs {
			if p := utils.ChavePlaca(r.Placa); p != "" && !vistas[p] && len(placas) < trajeto.VivoPlacasMax {
				vistas[p] = true
				placas = append(placas, p)
			}
		}
		if len(placas) == 0 {
			continue
		}
		if _, err := trajeto.NewAoVivo().Situacao(ctx, c.ID, token, placas); err != nil {
			logrus.Warnf("background: rastro de %s: %v", c.Nome, err)
		}
	}
}

func limparRastro() {
	antes := time.Now().AddDate(0, 0, -environment.Cfg.Rastro.Dias)
	if n, err := trajeto.NewAoVivo().LimparRastro(context.Background(), antes); err != nil {
		logrus.Warnf("background: limpeza do rastro: %v", err)
	} else if n > 0 {
		logrus.Infof("background: %d ponto(s) de rastro antigo(s) apagado(s)", n)
	}
}

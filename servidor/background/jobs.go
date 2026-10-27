// Package background roda as tarefas automáticas: a busca de dados (logo que o servidor sobe e depois
// a cada COLETA_INTERVALO_MIN), a limpeza das sessões de login vencidas (a cada 6 horas) e o WhatsApp
// dos motoristas (na hora marcada na aba WhatsApp, horário de Brasília, com o resultado do dia anterior).
package background

import (
	"context"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/sirupsen/logrus"

	"analisefrota/environment"
	"analisefrota/modules/coleta"
	"analisefrota/modules/sessao"
	"analisefrota/modules/whatsapp"
)

func Init() {
	s, err := gocron.NewScheduler()
	if err != nil {
		logrus.Fatalf("background: agendador: %v", err)
	}
	unica := gocron.WithSingletonMode(gocron.LimitModeReschedule) // uma rodada de cada vez

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

	if _, err := s.NewJob(gocron.DurationJob(6*time.Hour), gocron.NewTask(limparSessoes), gocron.WithStartAt(gocron.WithStartImmediately()), unica); err != nil {
		logrus.Fatalf("background: limpeza de sessões: %v", err)
	}

	// toda hora cheia: o módulo confere se o envio automático está ligado e se é a hora marcada (a aba WhatsApp muda sem reiniciar)
	if _, err := s.NewJob(gocron.CronJob("0 * * * *", false), gocron.NewTask(enviarWhatsApp), unica); err != nil {
		logrus.Fatalf("background: tarefa do WhatsApp: %v", err)
	}
	s.Start()
}

func enviarWhatsApp() {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Minute)
	defer cancel()
	whatsapp.NewWhatsApp().EnviarAgendado(ctx)
}

func buscarTodos() {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Minute)
	defer cancel()
	coleta.NewColeta().ExecutarTodos(ctx)
}

func limparSessoes() {
	n, err := sessao.NewSessao().LimparVencidas(context.Background())
	if err != nil {
		logrus.Warnf("background: limpeza de sessões: %v", err)
	} else if n > 0 {
		logrus.Infof("background: %d sessão(ões) vencida(s) apagada(s)", n)
	}
}

// Package background roda as tarefas automáticas: a busca de dados (logo que o servidor sobe e depois
// a cada COLETA_INTERVALO_MIN), a limpeza das sessões de login vencidas (a cada 6 horas) e o WhatsApp
// dos motoristas (todo dia na WHATSAPP_HORA de Brasília, com o resultado do dia anterior).
package background

import (
	"context"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/sirupsen/logrus"

	"analisefrota/common/utils"
	"analisefrota/environment"
	"analisefrota/modules/coleta"
	"analisefrota/modules/sessao"
	"analisefrota/modules/whatsapp"
)

func Init() {
	s, err := gocron.NewScheduler(gocron.WithLocation(utils.FusoBR())) // horas marcadas em horário de Brasília
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

	wa := environment.Cfg.WhatsApp
	if wa.Automatico {
		if _, err := s.NewJob(gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(uint(wa.Hora), 0, 0))), gocron.NewTask(enviarWhatsApp), unica); err != nil {
			logrus.Fatalf("background: tarefa do WhatsApp: %v", err)
		}
		logrus.Infof("background: WhatsApp dos motoristas todo dia às %dh (Brasília)", wa.Hora)
	} else {
		logrus.Info("background: WhatsApp automático desligado (WHATSAPP_AUTOMATICO=false)")
	}
	s.Start()
}

func enviarWhatsApp() {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Minute)
	defer cancel()
	whatsapp.NewWhatsApp().EnviarTodos(ctx)
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

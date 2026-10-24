package environment

type EmailConfig_t struct {
	// Servidor SMTP que envia os códigos de acesso (Google Workspace, Office 365...).
	// Vazio = modo de teste: o código não é enviado, aparece só no log do servidor.
	SMTPHost      string `envconfig:"SMTP_HOST"`
	SMTPPorta     int    `envconfig:"SMTP_PORTA" default:"587"`
	SMTPUsuario   string `envconfig:"SMTP_USUARIO"`
	SMTPSenha     string `envconfig:"SMTP_SENHA"`
	SMTPRemetente string `envconfig:"SMTP_REMETENTE"`
}

// Configurado indica se há SMTP para enviar e-mail de verdade
func (e EmailConfig_t) Configurado() bool { return e.SMTPHost != "" && e.SMTPRemetente != "" }

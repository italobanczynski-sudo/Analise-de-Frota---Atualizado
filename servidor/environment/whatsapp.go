package environment

type WhatsAppConfig_t struct {
	// Conta Twilio. Use uma API Key própria do bot (SK...), que dá para revogar sem mexer no resto da conta.
	// Vazio = modo de teste: nada é enviado, o texto aparece só no log do servidor.
	TwilioAccountSID string `envconfig:"TWILIO_ACCOUNT_SID"`
	TwilioAPIKey     string `envconfig:"TWILIO_API_KEY"`
	TwilioAPISecret  string `envconfig:"TWILIO_API_SECRET"`
	// Número WhatsApp Business ligado à Twilio, no formato whatsapp:+55DDNNNNNNNNN
	Remetente string `envconfig:"WHATSAPP_REMETENTE"`
	// Content SID (HX...) do template aprovado pela Meta, criado em Twilio > Content Template Builder
	TemplateSID string `envconfig:"WHATSAPP_TEMPLATE_SID"`
	// Envio automático todo dia, na hora de Brasília abaixo (0 a 23). Desligado até tudo estar configurado.
	Automatico bool `envconfig:"WHATSAPP_AUTOMATICO" default:"false"`
	Hora       int  `envconfig:"WHATSAPP_HORA" default:"7"`
}

// Configurado indica se há conta, remetente e template para enviar de verdade
func (w WhatsAppConfig_t) Configurado() bool {
	return w.TwilioAccountSID != "" && w.TwilioAPIKey != "" && w.TwilioAPISecret != "" && w.Remetente != "" && w.TemplateSID != ""
}

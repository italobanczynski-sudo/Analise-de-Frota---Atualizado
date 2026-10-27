-- Conexão com a Twilio, preenchida pelo administrador na aba WhatsApp (linha única, id = 1).
-- O Auth Token fica cifrado (AES-256-GCM, mesma chave dos tokens) e nunca volta para a tela.
-- Sem esta linha, valem as variáveis de ambiente TWILIO_* e WHATSAPP_*.
CREATE TABLE whatsapp_config (
	id                 SMALLINT PRIMARY KEY CHECK (id = 1),
	account_sid        TEXT NOT NULL,
	auth_token_cifrado BYTEA NOT NULL,
	remetente          TEXT NOT NULL,
	template_sid       TEXT NOT NULL,
	automatico         BOOLEAN NOT NULL DEFAULT false,
	hora               SMALLINT NOT NULL DEFAULT 7 CHECK (hora BETWEEN 0 AND 23),
	atualizado_em      TIMESTAMPTZ NOT NULL DEFAULT now(),
	atualizado_por     BIGINT REFERENCES usuarios (id) ON DELETE SET NULL
);

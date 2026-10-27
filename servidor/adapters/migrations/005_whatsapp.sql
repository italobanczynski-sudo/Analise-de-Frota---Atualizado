-- WhatsApp dos motoristas: o resultado do dia anterior chega toda manhã.
-- O motorista é o nome como vem da API Gobrax (o mesmo de registros_diarios.motorista).
-- O telefone é dado pessoal: fica cifrado (AES-256-GCM, mesma chave dos tokens); em texto só os 4 últimos dígitos.
-- Só recebe quem tem autorização registrada (autorizado_em) e está ativo.
CREATE TABLE contatos_motoristas (
	cliente_id       BIGINT NOT NULL REFERENCES clientes (id) ON DELETE CASCADE,
	motorista        TEXT NOT NULL,
	telefone_cifrado BYTEA NOT NULL,
	telefone_final   CHAR(4) NOT NULL,
	ativo            BOOLEAN NOT NULL DEFAULT true,
	autorizado_em    TIMESTAMPTZ NOT NULL,
	autorizado_por   BIGINT REFERENCES usuarios (id) ON DELETE SET NULL,
	atualizado_em    TIMESTAMPTZ NOT NULL DEFAULT now(),
	PRIMARY KEY (cliente_id, motorista)
);

-- Uma mensagem por cliente, motorista e dia de resultado: a chave impede mandar o mesmo dia duas vezes.
-- O telefone não é guardado aqui.
CREATE TABLE envios_whatsapp (
	cliente_id  BIGINT NOT NULL REFERENCES clientes (id) ON DELETE CASCADE,
	motorista   TEXT NOT NULL,
	dia         DATE NOT NULL,
	situacao    TEXT NOT NULL CHECK (situacao IN ('enviando', 'enviada', 'erro')),
	twilio_sid  TEXT NOT NULL DEFAULT '',
	erro        TEXT NOT NULL DEFAULT '',
	criado_em   TIMESTAMPTZ NOT NULL DEFAULT now(),
	PRIMARY KEY (cliente_id, motorista, dia)
);
CREATE INDEX envios_whatsapp_dia ON envios_whatsapp (cliente_id, dia DESC);

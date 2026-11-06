-- Telefone de cada motorista para a fila de envio das Mensagens do dia: o painel abre a conversa do motorista no
-- WhatsApp já com o texto e a pessoa só aperta Enviar. Não há envio automático. Um telefone por motorista e cliente;
-- o motorista é casado pelo nome sem diferenciar maiúsculas nem espaços (chave = ChaveMotorista).
CREATE TABLE telefones_motoristas (
	cliente_id     BIGINT NOT NULL REFERENCES clientes (id) ON DELETE CASCADE,
	chave          TEXT NOT NULL,
	motorista      TEXT NOT NULL,
	telefone       TEXT NOT NULL,   -- só dígitos, com o código do país (ex.: 55 + DDD + número)
	atualizado_em  TIMESTAMPTZ NOT NULL DEFAULT now(),
	atualizado_por BIGINT REFERENCES usuarios (id) ON DELETE SET NULL,
	PRIMARY KEY (cliente_id, chave)
);

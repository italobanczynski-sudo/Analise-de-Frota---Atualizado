-- A integração com a Twilio foi cancelada: as mensagens agora são copiadas e mandadas à mão.
-- Some com o que ela guardava (telefones dos motoristas, token da Twilio e envios), se existir.
DROP TABLE IF EXISTS envios_whatsapp;
DROP TABLE IF EXISTS contatos_motoristas;
DROP TABLE IF EXISTS whatsapp_config;

-- Mensagens marcadas como enviadas: um retrato da análise do dia, para acompanhar se o motorista
-- melhorou no ponto apontado. Sem telefone e sem o texto: só os números da análise.
CREATE TABLE mensagens_enviadas (
	cliente_id  BIGINT NOT NULL REFERENCES clientes (id) ON DELETE CASCADE,
	motorista   TEXT NOT NULL,
	dia         DATE NOT NULL,
	foco        TEXT NOT NULL DEFAULT '',  -- indicador apontado (coluna do diário); vazio = nenhum piorou
	foco_valor  DOUBLE PRECISION,          -- valor do indicador no dia
	foco_media  DOUBLE PRECISION,          -- média do motorista nos 7 dias anteriores
	melhorou    INT NOT NULL DEFAULT 0,
	piorou      INT NOT NULL DEFAULT 0,
	comparados  INT NOT NULL DEFAULT 0,
	nota        DOUBLE PRECISION,
	marcado_em  TIMESTAMPTZ NOT NULL DEFAULT now(),
	marcado_por BIGINT REFERENCES usuarios (id) ON DELETE SET NULL,
	PRIMARY KEY (cliente_id, motorista, dia)
);
CREATE INDEX mensagens_enviadas_motorista ON mensagens_enviadas (cliente_id, motorista, dia);

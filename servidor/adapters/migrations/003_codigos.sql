-- Códigos de acesso enviados por e-mail (primeiro acesso e "esqueci minha senha").
-- No banco fica só o HMAC do código (com a chave do servidor); vale 15 min e aceita até 5 tentativas.
CREATE TABLE codigos_acesso (
	id          BIGSERIAL PRIMARY KEY,
	usuario_id  BIGINT NOT NULL REFERENCES usuarios (id) ON DELETE CASCADE,
	codigo_hash BYTEA NOT NULL,
	expira_em   TIMESTAMPTZ NOT NULL,
	tentativas  INT NOT NULL DEFAULT 0,
	usado_em    TIMESTAMPTZ,
	criado_em   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX codigos_acesso_usuario ON codigos_acesso (usuario_id, criado_em DESC);

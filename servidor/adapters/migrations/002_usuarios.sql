-- Pessoas que usam a plataforma. A senha só existe como hash bcrypt; sem senha = convite ainda não aceito.
CREATE TABLE usuarios (
	id            BIGSERIAL PRIMARY KEY,
	nome          TEXT NOT NULL,
	email         TEXT NOT NULL,
	papel         TEXT NOT NULL CHECK (papel IN ('admin', 'colaborador', 'leitor')),
	senha_hash    TEXT,
	ativo         BOOLEAN NOT NULL DEFAULT true,
	criado_em     TIMESTAMPTZ NOT NULL DEFAULT now(),
	atualizado_em TIMESTAMPTZ NOT NULL DEFAULT now(),
	ultimo_acesso TIMESTAMPTZ
);
CREATE UNIQUE INDEX usuarios_email ON usuarios (lower(email));

-- Convites: o link leva um token aleatório; aqui fica só o SHA-256 dele. Uso único e com prazo.
CREATE TABLE convites (
	token_hash BYTEA PRIMARY KEY,
	usuario_id BIGINT NOT NULL REFERENCES usuarios (id) ON DELETE CASCADE,
	expira_em  TIMESTAMPTZ NOT NULL,
	usado_em   TIMESTAMPTZ,
	criado_em  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX convites_usuario ON convites (usuario_id);

-- Sessões de login: o cookie leva um token aleatório; aqui fica só o SHA-256 dele.
CREATE TABLE sessoes (
	token_hash BYTEA PRIMARY KEY,
	usuario_id BIGINT NOT NULL REFERENCES usuarios (id) ON DELETE CASCADE,
	expira_em  TIMESTAMPTZ NOT NULL,
	criado_em  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX sessoes_usuario ON sessoes (usuario_id);

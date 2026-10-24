-- Clientes da Gobrax acompanhados no painel. O token da API fica cifrado (AES-256-GCM); nunca em texto.
CREATE TABLE clientes (
	id            BIGSERIAL PRIMARY KEY,
	nome          TEXT NOT NULL UNIQUE,
	hubspot_id    TEXT NOT NULL DEFAULT '',
	token_cifrado BYTEA,
	criado_em     TIMESTAMPTZ NOT NULL DEFAULT now(),
	atualizado_em TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Um registro por cliente, dia (de Brasília) e veículo: estatísticas + performance da API.
-- Só o nome do motorista é guardado; o CPF que a API devolve é descartado antes de chegar aqui.
CREATE TABLE registros_diarios (
	cliente_id               BIGINT NOT NULL REFERENCES clientes (id) ON DELETE CASCADE,
	dia                      DATE NOT NULL,
	placa                    TEXT NOT NULL,
	motorista                TEXT NOT NULL DEFAULT '',
	rodou                    BOOLEAN NOT NULL DEFAULT false,
	km                       DOUBLE PRECISION,
	litros                   DOUBLE PRECISION,
	frenagens                DOUBLE PRECISION,
	frenagens_alta           DOUBLE PRECISION,
	velocidade_media         DOUBLE PRECISION,
	odometro                 DOUBLE PRECISION,
	horas                    DOUBLE PRECISION,
	nota                     DOUBLE PRECISION,
	faixa_verde              DOUBLE PRECISION,
	inicio_faixa_verde       DOUBLE PRECISION,
	final_faixa_verde        DOUBLE PRECISION,
	eco_roll                 DOUBLE PRECISION,
	pressao_ideal            DOUBLE PRECISION,
	pressao_atencao          DOUBLE PRECISION,
	embalo                   DOUBLE PRECISION,
	motor_ocioso             DOUBLE PRECISION,
	pressao_acelerador       DOUBLE PRECISION,
	excesso_rotacao          DOUBLE PRECISION,
	tempo_excesso_velocidade DOUBLE PRECISION,
	piloto_automatico        DOUBLE PRECISION,
	freio_motor              DOUBLE PRECISION,
	PRIMARY KEY (cliente_id, dia, placa)
);
CREATE INDEX registros_diarios_dia ON registros_diarios (cliente_id, dia);

-- Último odômetro de cada veículo (substituído a cada busca)
CREATE TABLE odometros (
	cliente_id BIGINT NOT NULL REFERENCES clientes (id) ON DELETE CASCADE,
	placa      TEXT NOT NULL,
	odometro   DOUBLE PRECISION,
	coletado   TEXT NOT NULL DEFAULT '',  -- como a API manda ("AAAA-MM-DD HH:MM:SS")
	buscado_em TIMESTAMPTZ NOT NULL DEFAULT now(),
	PRIMARY KEY (cliente_id, placa)
);

-- Nota, km e bonificação de cada motorista por mês (CPF e id da API são descartados)
CREATE TABLE premiacoes (
	cliente_id BIGINT NOT NULL REFERENCES clientes (id) ON DELETE CASCADE,
	mes        CHAR(7) NOT NULL,  -- AAAA-MM
	motorista  TEXT NOT NULL,
	nota       DOUBLE PRECISION,
	km         DOUBLE PRECISION,
	premio     DOUBLE PRECISION,
	PRIMARY KEY (cliente_id, mes, motorista)
);

-- Situação da busca automática de cada cliente
CREATE TABLE coletas (
	cliente_id     BIGINT PRIMARY KEY REFERENCES clientes (id) ON DELETE CASCADE,
	ultimo_dia     DATE,
	iniciada_em    TIMESTAMPTZ,
	terminada_em   TIMESTAMPTZ,
	situacao       TEXT NOT NULL DEFAULT 'nunca',  -- nunca, rodando, ok, erro
	mensagem       TEXT NOT NULL DEFAULT '',
	avisos         TEXT[] NOT NULL DEFAULT '{}'
);

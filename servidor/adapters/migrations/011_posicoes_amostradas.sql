-- Trajeto do dia: a API de posições só devolve a ÚLTIMA posição de um intervalo, então o trajeto é montado
-- perguntando janelas de 5 minutos. Cada janela consultada fica guardada aqui (vazio = a API não tinha posição nela),
-- para não pedir de novo o que já era definitivo quando foi buscado.
CREATE TABLE posicoes_amostradas (
	cliente_id  BIGINT NOT NULL REFERENCES clientes (id) ON DELETE CASCADE,
	placa       TEXT NOT NULL,
	janela_ini  TIMESTAMPTZ NOT NULL,
	data        TEXT NOT NULL DEFAULT '',
	lat         DOUBLE PRECISION,
	lon         DOUBLE PRECISION,
	vazio       BOOLEAN NOT NULL DEFAULT false,
	buscado_em  TIMESTAMPTZ NOT NULL DEFAULT now(),
	PRIMARY KEY (cliente_id, placa, janela_ini)
);

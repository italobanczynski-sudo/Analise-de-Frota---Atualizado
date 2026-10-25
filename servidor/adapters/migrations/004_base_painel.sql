-- Base do painel (index.html): os documentos que ele guardava na base compartilhada do Artifact
-- (clientes com grupo e observação, tratativas, relacionamento, HubSpot, configurações, planilhas importadas).
-- Os dados da API Gobrax não ficam aqui: o painel os recebe montados a partir de registros_diarios, odometros e premiacoes.
CREATE TABLE base_documentos (
	colecao        TEXT NOT NULL,
	id             TEXT NOT NULL,
	dados          JSONB NOT NULL,
	atualizado_em  TIMESTAMPTZ NOT NULL DEFAULT now(),
	atualizado_por BIGINT REFERENCES usuarios (id) ON DELETE SET NULL,
	PRIMARY KEY (colecao, id)
);

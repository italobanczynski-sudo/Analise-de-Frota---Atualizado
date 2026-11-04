-- Um registro por cliente, dia (de Brasília), veículo e motorista: o dado do condutor só no tempo em que ele estava
-- vinculado ao veículo, como a análise por motorista da plataforma (e o painel). Quem ficou o dia todo no veículo
-- repete a linha do veículo; os vínculos parciais vêm da API consultada só no trecho daquele motorista.
-- Mesmas medidas de registros_diarios. Só o nome do motorista é guardado; o CPF é descartado antes de chegar aqui.
CREATE TABLE registros_motoristas (
	cliente_id               BIGINT NOT NULL REFERENCES clientes (id) ON DELETE CASCADE,
	dia                      DATE NOT NULL,
	placa                    TEXT NOT NULL,
	motorista                TEXT NOT NULL,
	rodou                    BOOLEAN NOT NULL DEFAULT false,
	km                       DOUBLE PRECISION,
	litros                   DOUBLE PRECISION,
	frenagens                DOUBLE PRECISION,
	frenagens_alta           DOUBLE PRECISION,
	velocidade_media         DOUBLE PRECISION,
	odometro                 DOUBLE PRECISION,
	horas                    DOUBLE PRECISION,
	horas_motor              DOUBLE PRECISION,
	horas_pedal              DOUBLE PRECISION,
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
	PRIMARY KEY (cliente_id, dia, placa, motorista)
);
CREATE INDEX registros_motoristas_dia ON registros_motoristas (cliente_id, dia);

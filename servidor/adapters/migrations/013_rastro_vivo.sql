-- Rastro da frota ao vivo: a cada RASTRO_INTERVALO_MIN o servidor grava a última posição de cada veículo e a situação
-- (movimento, ligado = motor ligado parado, desligado, sem_sinal). O mapa da Visão geral desenha o rastro do dia com
-- os pontos de ignição ligada. Fica guardado RASTRO_DIAS dias.
CREATE TABLE rastro_vivo (
	cliente_id  BIGINT NOT NULL REFERENCES clientes (id) ON DELETE CASCADE,
	placa       TEXT NOT NULL,
	data        TIMESTAMPTZ NOT NULL,
	lat         DOUBLE PRECISION NOT NULL,
	lon         DOUBLE PRECISION NOT NULL,
	status      TEXT NOT NULL,
	velocidade  DOUBLE PRECISION,
	PRIMARY KEY (cliente_id, placa, data)
);
CREATE INDEX rastro_vivo_cliente_data ON rastro_vivo (cliente_id, data);

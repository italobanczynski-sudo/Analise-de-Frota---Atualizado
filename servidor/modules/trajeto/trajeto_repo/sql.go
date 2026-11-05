package trajeto_repo

const sqlListar = `SELECT cliente_id, placa, janela_ini, data, lat, lon, vazio, buscado_em
	FROM posicoes_amostradas
	WHERE cliente_id = $1 AND placa = $2 AND janela_ini >= $3 AND janela_ini < $4
	ORDER BY janela_ini`

const sqlSalvar = `INSERT INTO posicoes_amostradas (cliente_id, placa, janela_ini, data, lat, lon, vazio, buscado_em)
	VALUES (:cliente_id, :placa, :janela_ini, :data, :lat, :lon, :vazio, :buscado_em)
	ON CONFLICT (cliente_id, placa, janela_ini) DO UPDATE
	SET data = EXCLUDED.data, lat = EXCLUDED.lat, lon = EXCLUDED.lon, vazio = EXCLUDED.vazio, buscado_em = EXCLUDED.buscado_em`

const sqlSalvarRastro = `INSERT INTO rastro_vivo (cliente_id, placa, data, lat, lon, status, velocidade)
	VALUES (:cliente_id, :placa, :data, :lat, :lon, :status, :velocidade)
	ON CONFLICT (cliente_id, placa, data) DO UPDATE SET status = EXCLUDED.status, velocidade = EXCLUDED.velocidade`

// só ignição ligada (em movimento ou motor ligado parado): desligado e sem sinal não fazem rastro
const sqlListarRastro = `SELECT cliente_id, placa, data, lat, lon, status, velocidade
	FROM rastro_vivo
	WHERE cliente_id = $1 AND placa = ANY($2) AND data >= $3 AND status IN ('movimento', 'ligado')
	ORDER BY placa, data`

const sqlLimparRastro = `DELETE FROM rastro_vivo WHERE data < $1`

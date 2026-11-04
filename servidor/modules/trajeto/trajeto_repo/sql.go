package trajeto_repo

const sqlListar = `SELECT cliente_id, placa, janela_ini, data, lat, lon, vazio, buscado_em
	FROM posicoes_amostradas
	WHERE cliente_id = $1 AND placa = $2 AND janela_ini >= $3 AND janela_ini < $4
	ORDER BY janela_ini`

const sqlSalvar = `INSERT INTO posicoes_amostradas (cliente_id, placa, janela_ini, data, lat, lon, vazio, buscado_em)
	VALUES (:cliente_id, :placa, :janela_ini, :data, :lat, :lon, :vazio, :buscado_em)
	ON CONFLICT (cliente_id, placa, janela_ini) DO UPDATE
	SET data = EXCLUDED.data, lat = EXCLUDED.lat, lon = EXCLUDED.lon, vazio = EXCLUDED.vazio, buscado_em = EXCLUDED.buscado_em`

package diario_repo

const sqlApagarDia = `DELETE FROM registros_diarios WHERE cliente_id = $1 AND dia = $2`

const sqlInserir = `INSERT INTO registros_diarios (
	cliente_id, dia, placa, motorista, rodou, km, litros, frenagens, frenagens_alta, velocidade_media, odometro, horas, nota,
	faixa_verde, inicio_faixa_verde, final_faixa_verde, eco_roll, pressao_ideal, pressao_atencao, embalo, motor_ocioso,
	pressao_acelerador, excesso_rotacao, tempo_excesso_velocidade, piloto_automatico, freio_motor
) VALUES (
	:cliente_id, :dia, :placa, :motorista, :rodou, :km, :litros, :frenagens, :frenagens_alta, :velocidade_media, :odometro, :horas, :nota,
	:faixa_verde, :inicio_faixa_verde, :final_faixa_verde, :eco_roll, :pressao_ideal, :pressao_atencao, :embalo, :motor_ocioso,
	:pressao_acelerador, :excesso_rotacao, :tempo_excesso_velocidade, :piloto_automatico, :freio_motor
)`

const sqlListar = `SELECT
	cliente_id, to_char(dia, 'YYYY-MM-DD') AS dia, placa, motorista, rodou, km, litros, frenagens, frenagens_alta, velocidade_media,
	odometro, horas, nota, faixa_verde, inicio_faixa_verde, final_faixa_verde, eco_roll, pressao_ideal, pressao_atencao, embalo,
	motor_ocioso, pressao_acelerador, excesso_rotacao, tempo_excesso_velocidade, piloto_automatico, freio_motor
FROM registros_diarios
WHERE cliente_id = $1 AND dia BETWEEN $2 AND $3
ORDER BY dia, placa`

// um dia com performance tem ao menos um veículo com indicador de condução
const sqlDiasComIndicadores = `SELECT DISTINCT to_char(dia, 'YYYY-MM-DD') FROM registros_diarios
WHERE cliente_id = $1 AND dia BETWEEN $2 AND $3
	AND (faixa_verde IS NOT NULL OR inicio_faixa_verde IS NOT NULL OR embalo IS NOT NULL OR motor_ocioso IS NOT NULL)`

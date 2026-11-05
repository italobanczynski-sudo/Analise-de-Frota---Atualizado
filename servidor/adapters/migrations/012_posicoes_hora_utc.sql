-- A hora de lastPosition.date vem no horário de Brasília, mas era guardada como se fosse UTC (3 h a menos).
-- O adaptador passou a converter para UTC; aqui as posições já guardadas ganham as 3 h que faltavam.
UPDATE posicoes_amostradas
SET data = to_char(data::timestamp + interval '3 hours', 'YYYY-MM-DD HH24:MI:SS')
WHERE data ~ '^[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}$';

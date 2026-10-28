-- Base de cada percentual da performance, para juntar veículos e dias como a plataforma Gobrax:
-- horas_motor = tempo de ignição ligada (base do motor ligado parado);
-- horas_pedal = tempo com o pé no acelerador (base dos ranges do pedal). A base dos demais é horas (em movimento).
ALTER TABLE registros_diarios ADD COLUMN horas_motor DOUBLE PRECISION;
ALTER TABLE registros_diarios ADD COLUMN horas_pedal DOUBLE PRECISION;

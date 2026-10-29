-- Indicadores que pioraram contra o dia anterior na mensagem enviada (os marcados com ⚠️ no texto):
-- chaves do diário separadas por vírgula. Mensagens marcadas antes desta coluna ficam só com o foco.
ALTER TABLE mensagens_enviadas ADD COLUMN pioraram TEXT NOT NULL DEFAULT '';

// Regras do robô sem WhatsApp nem rede (testadas em regras.test.mjs)

// Mensagens a enviar: não enviadas e com telefone cadastrado
export const pendentes = (lista) => (lista || []).filter((m) => !m.enviada && m.telefone);

// Os dados do dia já chegaram por completo? (a busca da API roda de hora em hora no servidor)
export function dadosProntos(coleta, dia) {
  if (!coleta) return false;
  if (coleta.situacao === 'rodando' || coleta.situacao === 'aguardando') return false;
  return !!coleta.ultimo_dia && coleta.ultimo_dia >= dia;
}

// Ontem no horário de Brasília (AAAA-MM-DD)
export function ontemBR(agora = new Date()) {
  const hoje = new Intl.DateTimeFormat('en-CA', { timeZone: 'America/Sao_Paulo' }).format(agora);   // AAAA-MM-DD
  const d = new Date(hoje + 'T12:00:00Z');
  d.setUTCDate(d.getUTCDate() - 1);
  return d.toISOString().slice(0, 10);
}

// Telefone no log: só os 4 últimos dígitos
export const mascarar = (t) => (t ? '••••' + String(t).slice(-4) : 'sem telefone');

// Pausa entre mensagens, em ms, sorteada entre min e max segundos
export const pausaAleatoria = (min, max, sorteio = Math.random) => Math.round((min + sorteio() * Math.max(0, max - min)) * 1000);

// Robô do WhatsApp: todo dia de manhã manda as Mensagens do dia (análise de ontem) a cada motorista com telefone
// cadastrado e marca como enviada na plataforma. Conecta ao WhatsApp como um aparelho a mais (Baileys, sem navegador).
//
//   node robo.mjs parear      → mostra o QR Code: no celular, WhatsApp › Dispositivos conectados › Conectar
//   node robo.mjs teste       → mostra o que seria enviado hoje, sem enviar nem marcar nada
//   node robo.mjs             → envia (é o que o agendador roda às 08:00)
//
// ATENÇÃO: automatizar o WhatsApp é contra os Termos do WhatsApp e o número pode ser bloqueado.
// Use um número só para isso. As pausas aleatórias entre mensagens diminuem o risco, não o eliminam.
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { pendentes, telefonesRepetidos, dadosProntos, ontemBR, mascarar, pausaAleatoria } from './regras.mjs';

const AQUI = path.dirname(fileURLToPath(import.meta.url));
const MODO = process.argv[2] || 'enviar';

// ---------- configuração (.env desta pasta; nunca vai para o Git) ----------
function lerEnv() {
  const arq = path.join(AQUI, '.env');
  if (!fs.existsSync(arq)) throw new Error('Falta o arquivo robo-whatsapp/.env (copie o .env.example e preencha).');
  const env = {};
  for (const linha of fs.readFileSync(arq, 'utf8').split(/\r?\n/)) {
    const m = linha.match(/^\s*([A-Z_]+)\s*=\s*(.*)\s*$/);
    if (m) env[m[1]] = m[2].replace(/^["']|["']$/g, '');
  }
  return env;
}
const ENV = lerEnv();
const CFG = {
  servidor: (ENV.SERVIDOR_URL || 'https://analise-de-frota.onrender.com').replace(/\/$/, ''),
  email: ENV.ROBO_EMAIL || '',
  senha: ENV.ROBO_SENHA || '',
  pausaMin: Number(ENV.PAUSA_MIN_SEG || 25),
  pausaMax: Number(ENV.PAUSA_MAX_SEG || 60),
  limiteHora: Number(ENV.LIMITE_HORA || 12),       // até que hora espera os dados de ontem chegarem
  esperaMin: Number(ENV.ESPERA_DADOS_MIN || 20),   // de quanto em quanto tempo confere de novo
  maxPorDia: Number(ENV.MAX_POR_DIA || 200),
  clientes: (ENV.CLIENTES || '').split(',').map((s) => s.trim().toUpperCase()).filter(Boolean)  // vazio = todos
};

// ---------- log (um arquivo por dia em robo-whatsapp/logs; telefones mascarados) ----------
const pastaLog = path.join(AQUI, 'logs');
fs.mkdirSync(pastaLog, { recursive: true });
const arqLog = path.join(pastaLog, new Date().toISOString().slice(0, 10) + '.log');
function log(...partes) {
  const l = `[${new Date().toLocaleTimeString('pt-BR')}] ${partes.join(' ')}`;
  console.log(l);
  fs.appendFileSync(arqLog, l + '\n');
}
const dormir = (ms) => new Promise((r) => setTimeout(r, ms));

// ---------- plataforma ----------
let cookie = '';
async function chamar(metodo, caminho, corpo) {
  for (let tentativa = 1; ; tentativa++) {
    try {
      const r = await fetch(CFG.servidor + caminho, {
        method: metodo,
        headers: { 'X-Requisicao': 'plataforma', 'Content-Type': 'application/json', ...(cookie ? { Cookie: cookie } : {}) },
        body: corpo ? JSON.stringify(corpo) : undefined,
        signal: AbortSignal.timeout(90000)   // o servidor pode estar acordando
      });
      const sc = r.headers.get('set-cookie');
      if (sc) cookie = sc.split(';')[0];
      const j = await r.json().catch(() => ({}));
      if (!r.ok || j.ok === false) throw Object.assign(new Error(j.erro || 'HTTP ' + r.status), { status: r.status });
      return j.dados;
    } catch (e) {
      if (e.status && e.status < 500) throw e;   // erro de entrada ou login: tentar de novo não resolve
      if (tentativa >= 4) throw e;
      log(`servidor não respondeu (${e.message}); nova tentativa em 30 s`);
      await dormir(30000);
    }
  }
}
async function entrar() {
  if (!CFG.email || !CFG.senha) throw new Error('Preencha ROBO_EMAIL e ROBO_SENHA no .env.');
  await chamar('POST', '/api/v1/auth/entrar', { email: CFG.email, senha: CFG.senha });
  const eu = await chamar('GET', '/api/v1/auth/eu');
  if (eu && eu.papel === 'leitor') throw new Error('O usuário do robô é leitor: ele precisa ser colaborador para ver os telefones e marcar as mensagens.');
}

// ---------- WhatsApp (Baileys: fala direto com o WhatsApp, sem Chrome) ----------
// A sessão (as chaves do aparelho conectado) fica numa pasta fora do projeto: no Windows em %LOCALAPPDATA%,
// no Linux (máquina virtual) em ~/.local/share. Quem tiver essa pasta manda mensagem pelo número do robô.
const PASTA_DADOS = process.platform === 'win32'
  ? (process.env.LOCALAPPDATA || AQUI)
  : (process.env.XDG_DATA_HOME || path.join(process.env.HOME || AQUI, '.local', 'share'));
const SESSAO = ENV.PASTA_SESSAO || path.join(PASTA_DADOS, 'robo-whatsapp-gobrax', 'sessao');
async function abrirWhatsApp() {
  const baileys = await import('baileys');
  const { default: pino } = await import('pino');
  const { makeWASocket, useMultiFileAuthState, DisconnectReason, fetchLatestBaileysVersion, Browsers } = baileys;
  fs.mkdirSync(SESSAO, { recursive: true });
  const { state, saveCreds } = await useMultiFileAuthState(SESSAO);
  let version;
  try { ({ version } = await fetchLatestBaileysVersion()); } catch { /* fica a versão embutida */ }
  const limite = MODO === 'parear' ? 6 : 3;   // no pareamento, tempo para pegar o celular e ler o código
  return new Promise((ok, falha) => {
    let fim = false, sock = null, quedas = 0;
    const encerrar = (erro) => { if (fim) return; fim = true; clearTimeout(relogio); try { sock?.end(undefined); } catch { /* já fechado */ } falha(erro); };
    const relogio = setTimeout(() => encerrar(new Error(`O WhatsApp não conectou em ${limite} minutos.`)), limite * 60000);
    const conectar = () => {
      sock = makeWASocket({
        auth: state, version, logger: pino({ level: 'silent' }), browser: Browsers.ubuntu('Robo Gobrax'),
        markOnlineOnConnect: false, syncFullHistory: false
      });
      sock.ev.on('creds.update', saveCreds);
      sock.ev.on('connection.update', async (u) => {
        if (fim) return;
        if (u.qr) {
          if (MODO !== 'parear') { encerrar(new Error('O WhatsApp pediu o QR Code de novo: rode "npm run parear".')); return; }
          const { default: qrcode } = await import('qrcode-terminal');
          console.log('\nNo celular do número do robô: WhatsApp › Dispositivos conectados › Conectar um dispositivo, e leia o código:\n');
          qrcode.generate(u.qr, { small: true });
          if (process.env.QR_BRUTO) console.log('QR:' + u.qr);   // para o vm/parear-na-tela.mjs mostrar o código no navegador
        }
        if (u.connection === 'open') {
          fim = true; clearTimeout(relogio);
          const s = sock;
          ok({
            // mesmo formato do whatsapp-web.js: { _serialized } ou null se o número não tem WhatsApp
            getNumberId: async (tel) => {
              const [r] = (await s.onWhatsApp(String(tel).replace(/\D/g, '') + '@s.whatsapp.net')) || [];
              return r && r.exists ? { _serialized: r.jid } : null;
            },
            sendMessage: (jid, texto) => s.sendMessage(jid, { text: texto }),
            fechar: async () => { try { s.end(undefined); } catch { /* já fechado */ } }
          });
        }
        if (u.connection === 'close') {
          const codigo = u.lastDisconnect?.error?.output?.statusCode;
          if (codigo === DisconnectReason.loggedOut) {
            encerrar(new Error('O WhatsApp desconectou o robô (aparelho removido no celular): rode "npm run parear".'));
          } else if (++quedas > 5) {
            encerrar(new Error('O WhatsApp fechou a conexão várias vezes seguidas (código ' + codigo + ').'));
          } else {
            conectar();   // 515 logo depois de ler o QR Code é normal: o WhatsApp pede para reconectar
          }
        }
      });
    };
    conectar();
  });
}
const abrirWhatsAppComRetentativa = abrirWhatsApp;

// ---------- envio ----------
async function enviarDoCliente(wa, c, dia, conta) {
  const lista = await chamar('GET', `/api/v1/clientes/${encodeURIComponent(c.id)}/mensagens?dia=${dia}`);
  const repetidos = telefonesRepetidos(lista);
  const fila = pendentes(lista).filter((m) => {
    if (!repetidos.has(m.telefone)) return true;
    const outros = lista.filter((x) => x.telefone === m.telefone && x.motorista !== m.motorista).map((x) => x.motorista).join(', ');
    log(`  ${m.motorista} (${mascarar(m.telefone)}): mesmo telefone de ${outros}; não enviada, corrija o cadastro`);
    conta.erros.push({ cliente: c.nome, motorista: m.motorista, motivo: `mesmo telefone de ${outros}; corrija o cadastro em Mensagens do dia` });
    return false;
  });
  const semTel = lista.filter((m) => !m.enviada && !m.telefone).length;
  conta.semTelefone += semTel;
  log(`${c.nome}: ${lista.length} mensagens, ${fila.length} para enviar${semTel ? `, ${semTel} sem telefone (não vão)` : ''}`);
  for (const m of fila) {
    if (conta.n >= CFG.maxPorDia) {
      log(`limite de ${CFG.maxPorDia} mensagens no dia atingido`);
      fila.slice(fila.indexOf(m)).forEach((x) => conta.pendentes.push({ cliente: c.nome, motorista: x.motorista, motivo: 'limite de mensagens do dia' }));
      return;
    }
    if (MODO === 'teste') { log(`  [teste] mandaria para ${m.motorista} (${mascarar(m.telefone)})`); continue; }
    try {
      const id = await wa.getNumberId(m.telefone);
      if (!id) {
        log(`  ${m.motorista} (${mascarar(m.telefone)}): este número não tem WhatsApp; confira o cadastro`);
        conta.erros.push({ cliente: c.nome, motorista: m.motorista, motivo: 'o número não tem WhatsApp; confira o cadastro' });
        continue;
      }
      await wa.sendMessage(id._serialized, m.texto);
      await chamar('PUT', `/api/v1/clientes/${encodeURIComponent(c.id)}/mensagens/enviada`, { dia, motorista: m.motorista });
      conta.n++;
      conta.enviadas.push({ cliente: c.nome, motorista: m.motorista });
      log(`  enviada para ${m.motorista} (${mascarar(m.telefone)})`);
    } catch (e) {
      log(`  ERRO com ${m.motorista}: ${e.message}`);
      conta.erros.push({ cliente: c.nome, motorista: m.motorista, motivo: String(e.message || e).slice(0, 200) });
    }
    await dormir(pausaAleatoria(CFG.pausaMin, CFG.pausaMax));
  }
}

// Teste de envio de verdade, sem tocar nos motoristas: manda UMA mensagem de exemplo (a análise de ontem do primeiro
// motorista que houver) para o TELEFONE_TESTE do .env e não marca nada como enviado.
async function testeEnvio(dia) {
  const tel = String(ENV.TELEFONE_TESTE || '').replace(/\D/g, '');
  if (!tel) throw new Error('Para o teste de envio, preencha TELEFONE_TESTE no .env com o SEU número (DDD + número).');
  const numero = tel.length <= 11 ? '55' + tel : tel;
  let exemplo = null;
  for (const c of await chamar('GET', '/api/v1/clientes')) {
    const lista = await chamar('GET', `/api/v1/clientes/${encodeURIComponent(c.id)}/mensagens?dia=${dia}`).catch(() => []);
    if (lista.length) { exemplo = { cliente: c.nome, m: lista[0] }; break; }
  }
  const texto = '🧪 TESTE DO ROBÔ (só para você, nenhum motorista recebeu)\n\n' + (exemplo ? exemplo.m.texto : 'Sem análise de ontem para mostrar; o envio funcionou.');
  const wa = await abrirWhatsAppComRetentativa();
  try {
    const id = await wa.getNumberId(numero);
    if (!id) throw new Error(`o número de teste (${mascarar(numero)}) não tem WhatsApp`);
    await wa.sendMessage(id._serialized, texto);
    log(`teste enviado para ${mascarar(numero)}${exemplo ? ` com a análise de ${exemplo.m.motorista} (${exemplo.cliente})` : ''}. Nada foi marcado como enviado.`);
  } finally {
    await dormir(8000);
    await wa.fechar();
  }
}

// Manda a análise de ontem a UM condutor (pelo nome), mesmo que já marcada como enviada, e marca de novo.
//   node robo.mjs enviar-um "NOME DO CONDUTOR"
async function enviarUm(dia, nome) {
  const sem = (s) => String(s || '').normalize('NFD').replace(/[̀-ͯ]/g, '').toUpperCase().replace(/\s+/g, ' ').trim();
  if (!sem(nome)) throw new Error('Informe o nome: node robo.mjs enviar-um "NOME DO CONDUTOR"');
  const achados = [];
  for (const c of await chamar('GET', '/api/v1/clientes')) {
    const lista = await chamar('GET', `/api/v1/clientes/${encodeURIComponent(c.id)}/mensagens?dia=${dia}`).catch(() => []);
    lista.filter((m) => sem(m.motorista) === sem(nome)).forEach((m) => achados.push({ c, m }));
  }
  if (!achados.length) throw new Error(`${nome} não tem análise de ${dia} (não rodou nesse dia ou o nome está diferente).`);
  if (achados.length > 1) throw new Error(`${nome} aparece em mais de um cliente: ${achados.map((a) => a.c.nome).join(', ')}.`);
  const { c, m } = achados[0];
  if (!m.telefone) throw new Error(`${m.motorista} (${c.nome}) não tem telefone cadastrado.`);
  const wa = await abrirWhatsAppComRetentativa();
  try {
    const id = await wa.getNumberId(m.telefone);
    if (!id) throw new Error(`o número de ${m.motorista} (${mascarar(m.telefone)}) não tem WhatsApp`);
    await wa.sendMessage(id._serialized, m.texto);
    await chamar('PUT', `/api/v1/clientes/${encodeURIComponent(c.id)}/mensagens/enviada`, { dia, motorista: m.motorista });
    log(`enviada para ${m.motorista} (${c.nome}, ${mascarar(m.telefone)}): análise de ${dia}`);
  } finally {
    await dormir(8000);
    await wa.fechar();
  }
}

async function principal() {
  log(`== robô do WhatsApp · modo ${MODO}`);
  if (MODO === 'parear') {
    const wa = await abrirWhatsAppComRetentativa();
    log('WhatsApp conectado. A sessão ficou guardada em ' + SESSAO + '.');
    await dormir(10000);   // termina de sincronizar antes de fechar
    await wa.fechar();
    return;
  }
  const dia = ontemBR();
  await entrar();
  if (MODO === 'teste-envio') return testeEnvio(dia);
  if (MODO === 'enviar-um') return enviarUm(dia, process.argv[3]);
  const wa = MODO === 'teste' ? null : await abrirWhatsAppComRetentativa();
  const feitos = new Set(), conta = { n: 0, enviadas: [], erros: [], pendentes: [], semTelefone: 0 };
  try {
    for (;;) {
      const clientes = (await chamar('GET', '/api/v1/clientes'))
        .filter((c) => !CFG.clientes.length || CFG.clientes.includes(String(c.nome).toUpperCase()));
      const esperando = [];
      for (const c of clientes) {
        if (feitos.has(c.id)) continue;
        if (!dadosProntos(c.coleta, dia)) { esperando.push(c.nome); continue; }
        await enviarDoCliente(wa, c, dia, conta);
        feitos.add(c.id);
      }
      if (!esperando.length) break;
      const hora = Number(new Date().toLocaleString('en-US', { timeZone: 'America/Sao_Paulo', hour: 'numeric', hour12: false }));
      if (hora >= CFG.limiteHora || MODO === 'teste') {
        log(`dados de ${dia} ainda não chegaram para: ${esperando.join(', ')}. Ficam sem envio hoje.`);
        esperando.forEach((nome) => conta.pendentes.push({ cliente: nome, motorista: '', motivo: `os dados de ${dia.slice(8, 10)}/${dia.slice(5, 7)} não chegaram até ${CFG.limiteHora}h` }));
        break;
      }
      log(`dados de ${dia} ainda não chegaram para: ${esperando.join(', ')}. Confiro de novo em ${CFG.esperaMin} min.`);
      await dormir(CFG.esperaMin * 60000);
    }
  } finally {
    if (wa) { await dormir(8000); await wa.fechar(); }   // dá tempo da última mensagem sair
  }
  log(`== fim: ${conta.n} mensagem(ns) enviada(s) referentes a ${dia}${conta.erros.length ? `, ${conta.erros.length} com erro` : ''}${conta.pendentes.length ? `, ${conta.pendentes.length} pendente(s)` : ''}`);
  if (MODO === 'enviar') await registrarResumo(dia, conta, '');
}

// Resumo da execução na plataforma (coleção robo_envios, um documento por dia analisado): o sino do painel mostra
// quantas foram enviadas e o que deu erro ou ficou pendente. Só nomes, sem telefone.
async function registrarResumo(dia, conta, erroGeral) {
  const doc = {
    dia, executado_em: new Date().toISOString(), enviadas: conta.n, erros: conta.erros, pendentes: conta.pendentes,
    sem_telefone: conta.semTelefone, erro_geral: erroGeral || ''
  };
  try { if (!cookie) await entrar(); await chamar('PUT', `/api/v1/base/robo_envios/${dia}`, doc); log('resumo registrado na plataforma'); }
  catch (e) { log('não consegui registrar o resumo na plataforma: ' + e.message); }
}

principal().then(() => process.exit(0)).catch(async (e) => {
  log('ERRO: ' + e.message);
  // falha geral (WhatsApp desconectado, servidor fora…): também vira alerta no sino, se der para falar com a plataforma
  if (MODO === 'enviar') await registrarResumo(ontemBR(), { n: 0, erros: [], pendentes: [], semTelefone: 0 }, String(e.message || e).slice(0, 300));
  process.exit(1);
});

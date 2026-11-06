// Robô do WhatsApp: todo dia de manhã manda as Mensagens do dia (análise de ontem) a cada motorista com telefone
// cadastrado e marca como enviada na plataforma. Roda NESTE computador, com o WhatsApp Web aberto por baixo.
//
//   node robo.mjs parear      → mostra o QR Code: no celular, WhatsApp › Dispositivos conectados › Conectar
//   node robo.mjs teste       → mostra o que seria enviado hoje, sem enviar nem marcar nada
//   node robo.mjs             → envia (é o que o agendador do Windows roda às 08:00)
//
// ATENÇÃO: automatizar o WhatsApp Web é contra os Termos do WhatsApp e o número pode ser bloqueado.
// Use um número só para isso. As pausas aleatórias entre mensagens diminuem o risco, não o eliminam.
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { pendentes, dadosProntos, ontemBR, mascarar, pausaAleatoria } from './regras.mjs';

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
  chrome: ENV.CHROME || 'C:/Program Files/Google/Chrome/Application/chrome.exe',
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

// ---------- WhatsApp Web ----------
// O Chrome é aberto por aqui (e não pelo whatsapp-web.js) para liberar o armazenamento persistente: sem ele o
// WhatsApp Web nega o "storage bucket", desconecta (post_logout) e recarrega sem parar. A sessão fica no perfil
// do Chrome numa pasta SEM espaços no caminho (com espaços, como a do projeto, o WhatsApp Web também falha).
const PERFIL = ENV.PERFIL_CHROME || path.join(process.env.LOCALAPPDATA || AQUI, 'robo-whatsapp-gobrax', 'perfil');
async function abrirWhatsApp() {
  const { default: puppeteer } = await import('puppeteer');
  const { default: wweb } = await import('whatsapp-web.js');
  if (/\s/.test(PERFIL)) throw new Error('O caminho do perfil do Chrome não pode ter espaços: ' + PERFIL + ' (ajuste PERFIL_CHROME no .env).');
  const navegador = await puppeteer.launch({ executablePath: CFG.chrome, headless: MODO !== 'parear', defaultViewport: null, userDataDir: PERFIL });
  await navegador.defaultBrowserContext().overridePermissions('https://web.whatsapp.com', ['persistent-storage', 'notifications']);
  const cliente = new wweb.Client({ puppeteer: { browserWSEndpoint: navegador.wsEndpoint() } });
  cliente.fechar = async () => { await cliente.destroy().catch(() => {}); await navegador.close().catch(() => {}); };
  const pronto = new Promise((ok, falha) => {
    cliente.on('qr', async (qr) => {
      if (MODO !== 'parear') { falha(new Error('O WhatsApp pediu o QR Code de novo: rode "node robo.mjs parear".')); return; }
      const { default: qrcode } = await import('qrcode-terminal');
      console.log('\nNo celular do número do robô: WhatsApp › Dispositivos conectados › Conectar um dispositivo, e leia o código:\n');
      qrcode.generate(qr, { small: true });
    });
    cliente.on('ready', ok);
    cliente.on('auth_failure', (m) => falha(new Error('Falha no login do WhatsApp: ' + m)));
    const limite = MODO === 'parear' ? 6 : 3;   // no pareamento, tempo para pegar o celular e ler o código
    setTimeout(() => falha(new Error(`O WhatsApp Web não conectou em ${limite} minutos.`)), limite * 60000);
  });
  pronto.catch(() => {});   // a falha é tratada no await abaixo
  try {
    await cliente.initialize();
    await pronto;
  } catch (e) {
    await cliente.fechar();   // fecha o Chrome: senão a próxima tentativa acha o perfil em uso
    throw e;
  }
  return cliente;
}
// O WhatsApp Web às vezes recarrega a página enquanto abre ("Execution context was destroyed"): tenta de novo
async function abrirWhatsAppComRetentativa() {
  for (let tentativa = 1; ; tentativa++) {
    try {
      return await abrirWhatsApp();
    } catch (e) {
      const passageiro = /context was destroyed|Target closed|Navigation|detached/i.test(e.message);
      if (!passageiro || tentativa >= 4) throw e;
      log(`o WhatsApp Web recarregou ao abrir (${e.message}); tentativa ${tentativa + 1} de 4`);
      await dormir(5000);
    }
  }
}

// ---------- envio ----------
async function enviarDoCliente(wa, c, dia, conta) {
  const lista = await chamar('GET', `/api/v1/clientes/${encodeURIComponent(c.id)}/mensagens?dia=${dia}`);
  const fila = pendentes(lista);
  const semTel = lista.filter((m) => !m.enviada && !m.telefone).length;
  log(`${c.nome}: ${lista.length} mensagens, ${fila.length} para enviar${semTel ? `, ${semTel} sem telefone (não vão)` : ''}`);
  for (const m of fila) {
    if (conta.n >= CFG.maxPorDia) { log(`limite de ${CFG.maxPorDia} mensagens no dia atingido`); return; }
    if (MODO === 'teste') { log(`  [teste] mandaria para ${m.motorista} (${mascarar(m.telefone)})`); continue; }
    try {
      const id = await wa.getNumberId(m.telefone);
      if (!id) { log(`  ${m.motorista} (${mascarar(m.telefone)}): este número não tem WhatsApp; confira o cadastro`); continue; }
      await wa.sendMessage(id._serialized, m.texto);
      await chamar('PUT', `/api/v1/clientes/${encodeURIComponent(c.id)}/mensagens/enviada`, { dia, motorista: m.motorista });
      conta.n++;
      log(`  enviada para ${m.motorista} (${mascarar(m.telefone)})`);
    } catch (e) {
      log(`  ERRO com ${m.motorista}: ${e.message}`);
    }
    await dormir(pausaAleatoria(CFG.pausaMin, CFG.pausaMax));
  }
}

async function principal() {
  log(`== robô do WhatsApp · modo ${MODO}`);
  if (MODO === 'parear') {
    const wa = await abrirWhatsAppComRetentativa();
    log('WhatsApp conectado. A sessão ficou guardada em ' + PERFIL + '.');
    await dormir(10000);   // termina de sincronizar antes de fechar
    await wa.fechar();
    return;
  }
  const dia = ontemBR();
  await entrar();
  const wa = MODO === 'teste' ? null : await abrirWhatsAppComRetentativa();
  const feitos = new Set(), conta = { n: 0 };
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
      if (hora >= CFG.limiteHora || MODO === 'teste') { log(`dados de ${dia} ainda não chegaram para: ${esperando.join(', ')}. Ficam sem envio hoje.`); break; }
      log(`dados de ${dia} ainda não chegaram para: ${esperando.join(', ')}. Confiro de novo em ${CFG.esperaMin} min.`);
      await dormir(CFG.esperaMin * 60000);
    }
  } finally {
    if (wa) { await dormir(8000); await wa.fechar(); }   // dá tempo da última mensagem sair
  }
  log(`== fim: ${conta.n} mensagem(ns) enviada(s) referentes a ${dia}`);
}

principal().then(() => process.exit(0)).catch((e) => { log('ERRO: ' + e.message); process.exit(1); });

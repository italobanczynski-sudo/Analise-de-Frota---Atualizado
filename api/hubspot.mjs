// Puxa do HubSpot (API oficial, app privado) o perfil e o relacionamento de cada cliente listado em api/clientes.json.
// O token fica em api/.env como HUBSPOT_TOKEN (nunca no código nem em chat). Só leitura.
// Gera api/saida/<cliente>/hubspot.json, que o painel importa na Saúde do cliente (botão "Importar arquivo do HubSpot").
// Uso: node api/hubspot.mjs [--cliente "Nome do cliente"] [--trocar-token]
// Na primeira vez (ou com --trocar-token) o script pede o token no terminal, sem mostrá-lo, testa e guarda no api/.env.
// Escopos do app privado: crm.objects.companies.read, crm.objects.contacts.read, tickets, e-mails/ligações/reuniões (sales-email-read e crm.objects.*.read).
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const PASTA = path.dirname(fileURLToPath(import.meta.url));
const SAIDA = path.join(PASTA, 'saida');
const HUB = 'https://api.hubapi.com';
const arg = (nome, padrao) => { const i = process.argv.indexOf('--' + nome); return i > 0 && process.argv[i + 1] ? process.argv[i + 1] : padrao; };
const SO_CLIENTE = arg('cliente', '');

function lerEnv() {
  const f = path.join(PASTA, '.env');
  if (!fs.existsSync(f)) return {};
  const env = {};
  fs.readFileSync(f, 'utf8').split(/\r?\n/).forEach((l) => {
    const m = l.match(/^\s*([A-Za-z_][A-Za-z0-9_]*)\s*=\s*(.*)\s*$/);
    if (m && !l.trim().startsWith('#')) env[m[1]] = m[2].replace(/^(['"])(.*)\1$/, '$2');
  });
  return env;
}
const ENV = Object.assign(lerEnv(), process.env);
let TOKEN = String(ENV.HUBSPOT_TOKEN || '').replace(/^\s*bearer\s+/i, '').trim();
// Sem token no api/.env: pede no terminal (sem mostrar o que é digitado), testa e guarda no api/.env
async function pedirToken() {
  if (!process.stdin.isTTY) { console.error('Falta HUBSPOT_TOKEN no api/.env (token do app privado do HubSpot).'); process.exit(1); }
  process.stdout.write('Cole o token do app privado do HubSpot e aperte Enter (ele não aparece na tela): ');
  const t = await new Promise((ok) => {
    let buf = '';
    process.stdin.setRawMode(true); process.stdin.resume(); process.stdin.setEncoding('utf8');
    const ler = (ch) => {
      for (const c of ch) {
        if (c === '\r' || c === '\n') { process.stdin.setRawMode(false); process.stdin.pause(); process.stdin.removeListener('data', ler); process.stdout.write('\n'); ok(buf); return; }
        if (c === '\u0003') { process.stdout.write('\n'); process.exit(1); }   // Ctrl+C
        if (c === '\u007f' || c === '\b') buf = buf.slice(0, -1); else buf += c;
      }
    };
    process.stdin.on('data', ler);
  });
  const tok = String(t).replace(/^\s*(authorization\s*:\s*)?bearer\s+/i, '').replace(/["'\s]/g, '');
  if (!tok) { console.error('Nenhum token informado.'); process.exit(1); }
  const r = await fetch(HUB + '/crm/v3/objects/companies?limit=1', { headers: { Authorization: 'Bearer ' + tok } });
  if (!r.ok) { console.error(`O HubSpot recusou o token (${r.status}). Confira se é o token do app privado e se ele tem os escopos de leitura.`); process.exit(1); }
  const fEnv = path.join(PASTA, '.env');
  const atual = fs.existsSync(fEnv) ? fs.readFileSync(fEnv, 'utf8') : '';
  const novo = /^\s*HUBSPOT_TOKEN\s*=.*$/m.test(atual) ? atual.replace(/^\s*HUBSPOT_TOKEN\s*=.*$/m, 'HUBSPOT_TOKEN=' + tok) : atual.replace(/\s*$/, '') + (atual ? '\n' : '') + 'HUBSPOT_TOKEN=' + tok + '\n';
  fs.writeFileSync(fEnv, novo);
  console.log('Token válido e guardado em api/.env (fora do Git).');
  return tok;
}
const fClientes = path.join(PASTA, 'clientes.json');
if (!fs.existsSync(fClientes)) { console.error('Falta api/clientes.json.'); process.exit(1); }
const CLIENTES = JSON.parse(fs.readFileSync(fClientes, 'utf8'));

const nrm = (v) => String(v || '').normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase().replace(/\s+/g, ' ').trim();
const ckey = (v) => nrm(v).replace(/[.,/\-]/g, ' ').replace(/\b(ltda|eireli|epp|me|s a|sa)\b/g, ' ').replace(/\s+/g, ' ').trim();
const slugCli = (n) => ckey(n).replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '') || 'sem-cliente';
const esperar = (ms) => new Promise((r) => setTimeout(r, ms));

// Mesmos campos que o painel lê pelo conector (index.html, cartão Cliente no HubSpot)
const PROPS = {
  companies: ['name', 'cnpj_texto', 'nome_fantasia', 'city', 'state', 'estados_onde_opera', 'tipo_de_veiculo', 'frota_total', 'frota_contratada', 'veiculos_instalados', 'veiculos_offline',
    'nota_do_chs', 'chs_score', 'chs_aderencia', 'chs_nota_engajamento', 'chs_nota_adimplencia', 'chs_nota_usuario', 'chs_nota_motorista', 'chs_nota_tickets_positivos', 'tickets_suporte__chs_',
    'notes_last_contacted', 'notes_last_updated', 'hs_last_logged_call_date', 'data_inicio_do_contrato_original', 'data_de_termino_do_contrato', 'ciclo_experiencia', 'status_contrato', 'status_financeiro',
    'lifecyclestage', 'num_associated_contacts', 'num_associated_deals', 'createdate'],
  calls: ['hs_timestamp', 'hs_call_title', 'hs_call_status', 'hs_call_direction', 'hs_call_duration'],
  emails: ['hs_timestamp', 'hs_email_subject', 'hs_email_direction'],
  meetings: ['hs_timestamp', 'hs_meeting_title', 'hs_meeting_outcome'],
  tickets: ['subject', 'createdate', 'closed_date', 'hs_ticket_category', 'hs_ticket_priority', 'time_to_close'],
  contacts: ['firstname', 'lastname', 'jobtitle', 'email', 'phone', 'mobilephone', 'notes_last_contacted']
};
const MAX = 200;   // registros guardados por tipo (os totais são sempre completos)

async function hub(metodo, caminho, corpo) {
  for (let t = 0; ; t++) {
    const r = await fetch(HUB + caminho, { method: metodo, headers: { Authorization: 'Bearer ' + TOKEN, 'Content-Type': 'application/json' }, body: corpo ? JSON.stringify(corpo) : undefined });
    if (r.status === 429 && t < 8) { await esperar((Number(r.headers.get('Retry-After')) || 1) * 1000 + 200); continue; }
    if (!r.ok) {
      const txt = await r.text().catch(() => '');
      const e = new Error(`HubSpot ${r.status} em ${caminho.split('?')[0]}`);
      e.status = r.status; e.det = txt.slice(0, 300);
      throw e;
    }
    return r.status === 204 ? null : r.json();
  }
}
// Todos os ids associados (API de associações v4, paginada)
async function associados(de, id, para) {
  const ids = [];
  let after = '';
  do {
    const j = await hub('GET', `/crm/v4/objects/${de}/${id}/associations/${para}?limit=500${after ? '&after=' + encodeURIComponent(after) : ''}`);
    (j.results || []).forEach((x) => ids.push(String(x.toObjectId)));
    after = j.paging && j.paging.next && j.paging.next.after;
  } while (after);
  return ids;
}
async function lerLote(tipo, ids) {
  const out = [];
  for (let i = 0; i < ids.length; i += 100) {
    const j = await hub('POST', `/crm/v3/objects/${tipo}/batch/read`, { properties: PROPS[tipo], inputs: ids.slice(i, i + 100).map((id) => ({ id })) });
    (j.results || []).forEach((o) => out.push({ id: String(o.id), properties: o.properties || {} }));
  }
  return out;
}
const porData = (campo) => (a, b) => (String(a.properties[campo] || '') < String(b.properties[campo] || '') ? 1 : -1);

async function empresaId(c) {
  if (c.hubspot_id) return String(c.hubspot_id);
  const j = await hub('POST', '/crm/v3/objects/companies/search', { query: String(c.cliente).replace(/\b(ltda|eireli|epp|me|s\/?a)\b\.?/gi, '').trim().slice(0, 100), properties: ['name'], limit: 10 });
  const lista = j.results || [];
  const m = lista.find((o) => ckey(o.properties && o.properties.name) === ckey(c.cliente)) || lista[0];
  return m ? String(m.id) : null;
}

async function puxar(c) {
  const id = await empresaId(c);
  if (!id) { console.log(`  ${c.cliente}: não encontrado no HubSpot (preencha "hubspot_id" no api/clientes.json).`); return; }
  const comp = await hub('GET', `/crm/v3/objects/companies/${id}?properties=${PROPS.companies.join(',')}`);
  const tipos = ['calls', 'emails', 'meetings', 'tickets', 'contacts'];
  const ids = {};
  for (const t of tipos) { try { ids[t] = await associados('companies', id, t); } catch (e) { ids[t] = []; console.log(`  ${t}: sem acesso (${e.status}). Confira os escopos do app privado.`); } }
  const reg = {};
  for (const t of tipos) reg[t] = await lerLote(t, ids[t]);
  reg.calls.sort(porData('hs_timestamp')); reg.emails.sort(porData('hs_timestamp')); reg.meetings.sort(porData('hs_timestamp'));
  reg.tickets.sort(porData('createdate')); reg.contacts.sort(porData('notes_last_contacted'));
  const abertos = reg.tickets.filter((o) => !o.properties.closed_date);
  // quem mais fala com a Gobrax: ligações e e-mails de cada contato (até 10 contatos)
  const porContato = [];
  for (const o of reg.contacts.slice(0, 10)) {
    const [l, e] = await Promise.all([associados('contacts', o.id, 'calls').catch(() => []), associados('contacts', o.id, 'emails').catch(() => [])]);
    porContato.push({ id: o.id, l: l.length, e: e.length });
  }
  const bloco = (lista) => ({ total: lista.length, results: lista.slice(0, MAX) });
  const saida = {
    formato: 'gobrax-hubspot-1', gerado: new Date().toISOString(), cliente: c.cliente, id,
    comp: comp.properties || {}, calls: bloco(reg.calls), emails: bloco(reg.emails), meets: bloco(reg.meetings),
    tickets: bloco(reg.tickets), abertos: bloco(abertos), contatos: bloco(reg.contacts), porContato
  };
  const dir = path.join(SAIDA, slugCli(c.cliente));
  fs.mkdirSync(dir, { recursive: true });
  fs.writeFileSync(path.join(dir, 'hubspot.json'), JSON.stringify(saida));
  // resumo só com números (nada de nomes, e-mails ou telefones no terminal)
  console.log(`  ${c.cliente}: empresa ${id} · ${reg.calls.length} ligações · ${reg.emails.length} e-mails · ${reg.meetings.length} reuniões · ${reg.tickets.length} chamados (${abertos.length} abertos) · ${reg.contacts.length} contatos`);
  console.log(`  → ${path.relative(process.cwd(), path.join(dir, 'hubspot.json'))}`);
}

(async () => {
  if (!TOKEN || process.argv.includes('--trocar-token')) TOKEN = await pedirToken();
  const lista = CLIENTES.filter((c) => !SO_CLIENTE || ckey(c.cliente) === ckey(SO_CLIENTE));
  if (!lista.length) { console.error('Nenhum cliente no api/clientes.json com esse nome.'); process.exit(1); }
  console.log('HubSpot: ' + lista.length + ' cliente(s)');
  let falhas = 0;
  for (const c of lista) {
    try { await puxar(c); } catch (e) {
      falhas++;
      console.log(`  ${c.cliente}: ${e.message}${e.status === 401 ? ' (token inválido ou expirado)' : e.status === 403 ? ' (o app privado não tem o escopo necessário)' : ''}`);
    }
  }
  process.exit(falhas === lista.length ? 1 : 0);
})();

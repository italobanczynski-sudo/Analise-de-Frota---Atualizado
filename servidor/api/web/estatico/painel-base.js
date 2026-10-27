// Ponte entre o painel (index.html) e o servidor: oferece em window.claude.use('db') a mesma base compartilhada
// que o painel usa no Artifact, guardada no Postgres do servidor. O painel roda sem nenhuma mudança.
// - Coleções gravadas pelo painel (clientes, tratativas, planilhas importadas...): /api/v1/base/:colecao
// - Dados da API Gobrax que o servidor busca sozinho: /api/v1/painel/dados, montados aqui como as
//   planilhas de veículos e motoristas, o odômetro e a premiação que o próprio painel gravaria.
'use strict';
(function () {
  const ATUALIZA_MS = 30000, ATUALIZA_DADOS_MS = 10 * 60000;

  async function chamar(metodo, caminho, corpo) {
    const op = { method: metodo, credentials: 'same-origin', headers: { 'X-Requisicao': 'plataforma' } };
    if (corpo !== undefined) { op.headers['Content-Type'] = 'application/json'; op.body = JSON.stringify(corpo); }
    const r = await fetch(caminho, op);
    if (r.status === 401) { location.replace('/'); throw Object.assign(new Error('sessão encerrada'), { code: 'unauthenticated' }); }
    let j = null;
    try { j = await r.json(); } catch (e) { /* sem JSON */ }
    if (!r.ok || !j || !j.ok) {
      const code = r.status === 403 ? 'permission_denied' : r.status === 413 ? 'quota_exceeded' : 'erro';
      throw Object.assign(new Error((j && j.erro) || 'O servidor respondeu ' + r.status), { code });
    }
    return j.dados;
  }

  // ---------- mesmas regras de nome do painel (ckey, slugCli e idPlan do index.html) ----------
  const nrm = (v) => String(v || '').normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase().replace(/\s+/g, ' ').trim();
  const ckey = (v) => nrm(v).replace(/[.,/\-]/g, ' ').replace(/\b(ltda|eireli|epp|me|s a|sa)\b/g, ' ').replace(/\s+/g, ' ').trim();
  const slugCli = (n) => ckey(n).replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '') || 'sem-cliente';
  const idPlan = (alvo, cli, data, parte) => [alvo, slugCli(cli), data, 'p' + parte].join('~');
  const br = (iso) => iso.split('-').reverse().join('/');
  const num = (x) => (x === null || x === undefined || !Number.isFinite(Number(x)) ? '' : String(Math.round(Number(x) * 100) / 100).replace('.', ','));
  const escC = (s) => (/[;"\n\r]/.test(s) ? '"' + s.replace(/"/g, '""') + '"' : s);
  const PERF = ['faixa_verde', 'inicio_faixa_verde', 'final_faixa_verde', 'eco_roll', 'pressao_ideal', 'pressao_atencao', 'embalo', 'motor_ocioso',
    'pressao_acelerador', 'excesso_rotacao', 'tempo_excesso_velocidade', 'piloto_automatico', 'freio_motor'];
  const COLS = ['data', 'motorista', 'veiculo', 'km', 'litros', 'frenagens', 'frenagens_alta', 'velocidade_media', 'odometro', 'horas', 'nota'].concat(PERF, ['cliente', 'dias_periodo']);

  // um dia de um cliente → as duas planilhas, no mesmo layout de diaParaCSV do painel
  function linhaCSV(r, cli) {
    const rodou = !!r.rodou, so = (v) => (rodou ? num(v) : '');
    const nota = r.nota !== null && r.nota >= 0 && r.nota <= 100 ? num(r.nota) : '';
    return [br(r.dia), escC(r.motorista || r.placa), escC(r.placa), num(r.km || 0), so(r.litros), so(r.frenagens), so(r.frenagens_alta), so(r.velocidade_media), num(r.odometro),
      num(r.horas), nota].concat(PERF.map((k) => num(r[k])), [escC(cli), '1']).join(';');
  }
  function docsDoServidor(dados) {
    const out = new Map(), em = new Date().toISOString(), cab = COLS.join(';');
    dados.forEach((c) => {
      const porDia = new Map();
      (c.registros || []).forEach((r) => { const k = String(r.dia).slice(0, 10); (porDia.get(k) || porDia.set(k, []).get(k)).push(Object.assign({}, r, { dia: k })); });
      porDia.forEach((rs, dia) => {
        const veic = rs.map((r) => linhaCSV(r, c.nome));
        const mot = rs.filter((r) => r.rodou && r.motorista).map((r) => linhaCSV(r, c.nome));
        out.set(idPlan('veiculos', c.nome, dia, 1), { alvo: 'veiculos', cliente: c.nome, data: dia, colunas: COLS, texto: [cab].concat(veic).join('\n'), parte: 1, partes: 1, arquivo: 'API Gobrax · estatísticas', em });
        if (mot.length) out.set(idPlan('motoristas', c.nome, dia, 1), { alvo: 'motoristas', cliente: c.nome, data: dia, colunas: COLS, texto: [cab].concat(mot).join('\n'), parte: 1, partes: 1, arquivo: 'API Gobrax · motoristas', em });
      });
      if ((c.odometros || []).length) {
        out.set(idPlan('odometro', c.nome, 'atual', 1), { alvo: 'odometro', cliente: c.nome, data: em.slice(0, 10),
          veiculos: c.odometros.map((o) => ({ placa: o.placa, odometro: o.odometro, coletado: o.coletado })), em });
      }
      const porMes = new Map();
      (c.premiacoes || []).forEach((p) => (porMes.get(p.mes) || porMes.set(p.mes, []).get(p.mes)).push({ nome: p.motorista, nota: p.nota, km: p.km, premio: p.premio }));
      porMes.forEach((ms, mes) => out.set(idPlan('premiacao', c.nome, mes, 1), { alvo: 'premiacao', cliente: c.nome, data: mes, motoristas: ms, em }));
    });
    return out;
  }

  // ---------- coleções com "onSnapshot" por consulta periódica ----------
  let dadosServidor = null, dadosEm = 0;
  async function lerDadosServidor() {
    if (!dadosServidor || Date.now() - dadosEm > ATUALIZA_DADOS_MS) {
      try { dadosServidor = await chamar('GET', '/api/v1/painel/dados'); dadosEm = Date.now(); } catch (e) { if (!dadosServidor) throw e; }
    }
    return dadosServidor;
  }

  const colecoes = new Map();
  function colecao(nome) {
    if (colecoes.has(nome)) return colecoes.get(nome);
    const c = { nome, docs: new Map(), ouvintes: [], carregado: false, timer: null };
    c.avisar = () => {
      const docs = [...c.docs.entries()].map(([id, d]) => ({ id, exists: true, data: () => JSON.parse(JSON.stringify(d)) }));
      c.ouvintes.slice().forEach((o) => { try { o.ok({ docs, size: docs.length, empty: !docs.length }); } catch (e) { console.error(e); } });
    };
    c.carregar = async () => {
      const gravados = await chamar('GET', '/api/v1/base/' + nome);
      const m = new Map();
      if (nome === 'planilhas') {   // dados da API buscados pelo servidor; o que alguém importou à mão no mesmo dia vale mais
        docsDoServidor(await lerDadosServidor()).forEach((d, id) => m.set(id, d));
      }
      gravados.forEach((d) => m.set(d.id, d.dados));
      if (nome === 'clientes') {    // clientes cadastrados no servidor que o painel ainda não conhece
        const conhecidos = new Set([...m.values()].map((d) => ckey(d.nome)));
        (await lerDadosServidor()).forEach((s) => { if (!conhecidos.has(ckey(s.nome))) m.set('srv-' + s.id, { nome: s.nome, grupo: '', obs: '', criado: '' }); });
      }
      c.docs = m; c.carregado = true; c.avisar();
    };
    c.atualizar = () => c.carregar().catch((e) => c.ouvintes.forEach((o) => o.erro && o.erro(e)));
    colecoes.set(nome, c);
    return c;
  }
  function ouvir(c, ok, erro) {
    const o = { ok, erro };
    c.ouvintes.push(o);
    if (c.carregado) setTimeout(c.avisar, 0);
    else if (c.ouvintes.length === 1) c.atualizar();
    if (!c.timer) c.timer = setInterval(() => { if (!document.hidden) c.atualizar(); }, ATUALIZA_MS);
    return () => { c.ouvintes = c.ouvintes.filter((x) => x !== o); };
  }

  function refDoc(c, id) {
    return {
      id,
      async set(dados) { await chamar('PUT', '/api/v1/base/' + c.nome + '/' + encodeURIComponent(id), dados); c.docs.set(id, JSON.parse(JSON.stringify(dados))); c.avisar(); },
      async delete() { await chamar('DELETE', '/api/v1/base/' + c.nome + '/' + encodeURIComponent(id)); c.docs.delete(id); c.avisar(); },
      onSnapshot(ok, erro) {
        return ouvir(c, () => { const d = c.docs.get(id); ok({ id, exists: d !== undefined, data: () => (d === undefined ? undefined : JSON.parse(JSON.stringify(d))) }); }, erro);
      }
    };
  }
  const db = {
    collection(nome) {
      const c = colecao(nome);
      return {
        onSnapshot: (ok, erro) => ouvir(c, ok, erro),
        doc: (id) => refDoc(c, id),
        async add(dados) { const r = await chamar('POST', '/api/v1/base/' + nome, dados); c.docs.set(r.id, JSON.parse(JSON.stringify(dados))); c.avisar(); return refDoc(c, r.id); }
      };
    }
  };
  window.claude = {
    use: (nome) => (nome === 'db' ? Promise.resolve(db) : Promise.reject(Object.assign(new Error(nome + ' indisponível fora do Claude'), { code: 'unavailable' })))
  };

  // ---------- quem está logado: botão redondo com as iniciais no topo do painel; o menu (nome, Clientes e usuários, Sair) só abre no clique ----------
  function botaoConta(eu) {
    // o painel usa isto para mostrar a seção Usuários em Configurações (só para administradores)
    window.GBX_CONTA = { eu, chamar };
    dispatchEvent(new Event('gbx-conta'));
    const iniciais = String(eu.nome || '?').trim().split(/\s+/).map((p) => p[0]).slice(0, 2).join('').toUpperCase();
    const caixa = document.createElement('div');
    caixa.className = 'sino conta';
    const btn = document.createElement('button');
    btn.type = 'button'; btn.className = 'btn-tema'; btn.textContent = iniciais;
    btn.title = eu.nome; btn.setAttribute('aria-label', 'Sua conta: ' + eu.nome); btn.setAttribute('aria-haspopup', 'true'); btn.setAttribute('aria-expanded', 'false');
    Object.assign(btn.style, { fontWeight: '700', fontSize: '13px' });
    const menu = document.createElement('div');
    menu.className = 'sino-painel'; menu.hidden = true;
    Object.assign(menu.style, { left: 'auto', right: '0', width: '240px' });
    const nome = document.createElement('div');
    nome.className = 'sino-cab'; nome.textContent = eu.nome;
    Object.assign(nome.style, { fontWeight: '600' });
    const item = (texto, acao) => { const a = document.createElement('button'); a.type = 'button'; a.className = 'dm-op'; a.textContent = texto; a.addEventListener('click', acao); return a; };
    if (eu.papel === 'admin') menu.append(nome, item('Usuários', () => { abrir(false); location.hash = '#/config'; setTimeout(() => { const s = document.getElementById('usuarios'); if (s) s.scrollIntoView({ block: 'start' }); }, 300); }));
    else menu.append(nome);
    if (eu.papel !== 'leitor') menu.append(item('WhatsApp dos motoristas', () => location.assign('/plataforma#whatsapp')));
    menu.append(item('Clientes e coleta', () => location.assign('/plataforma')),
      item('Sair', async () => { try { await chamar('POST', '/api/v1/auth/sair'); } catch (e) { /* sai mesmo assim */ } location.replace('/'); }));
    const abrir = (sim) => { menu.hidden = !sim; btn.setAttribute('aria-expanded', String(sim)); };
    btn.addEventListener('click', (ev) => { ev.stopPropagation(); abrir(menu.hidden); });
    document.addEventListener('click', (ev) => { if (!caixa.contains(ev.target)) abrir(false); });
    document.addEventListener('keydown', (ev) => { if (ev.key === 'Escape') abrir(false); });
    caixa.append(btn, menu);
    // o painel redesenha o topo a cada mudança: o botão volta para o lugar sempre que some
    const encaixar = () => { const alvo = document.querySelector('.topo-acoes'); if (alvo && !alvo.contains(caixa)) alvo.appendChild(caixa); };
    encaixar();
    new MutationObserver(encaixar).observe(document.getElementById('app') || document.body, { childList: true, subtree: true });
  }
  addEventListener('DOMContentLoaded', () => { chamar('GET', '/api/v1/auth/eu').then(botaoConta).catch(() => {}); });
})();

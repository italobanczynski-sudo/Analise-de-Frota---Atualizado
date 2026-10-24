// Fundo das telas de entrada: "torre de controle" com rotas, veículos e alertas sendo tratados.
'use strict';
(function () {
  const c = document.getElementById('fx');
  if (!c) return;
  const x = c.getContext('2d');
  const parado = matchMedia('(prefers-reduced-motion: reduce)').matches;
  let W, H, dpr, hubs, routes, trucks, tower;
  const Y = '254,219,0', O = '255,154,60';

  function size() {
    dpr = Math.min(devicePixelRatio || 1, 2); W = c.width = innerWidth * dpr; H = c.height = innerHeight * dpr;
    tower = { x: W * 0.5, y: H * 0.52 };
    hubs = [...Array(16)].map((_, i) => {
      const a = i / 16 * 6.283 + Math.random() * 0.4, r = (0.18 + Math.random() * 0.3) * Math.min(W, H) * 1.15;
      return { x: tower.x + Math.cos(a) * r * 1.35, y: tower.y + Math.sin(a) * r * 0.85, alert: 0, ok: 0 };
    });
    routes = hubs.map((h) => ({ h, mx: (h.x + tower.x) / 2 + (Math.random() - 0.5) * W * 0.12, my: (h.y + tower.y) / 2 + (Math.random() - 0.5) * H * 0.2 }));
    trucks = [];
    routes.forEach((r) => { for (let k = 0; k < 2; k++) trucks.push({ r, p: Math.random(), v: 0.05 + Math.random() * 0.07, dir: k ? 1 : -1 }); });
  }
  addEventListener('resize', size);
  size();
  const bez = (r, p) => { const a = 1 - p; return { x: a * a * r.h.x + 2 * a * p * r.mx + p * p * tower.x, y: a * a * r.h.y + 2 * a * p * r.my + p * p * tower.y }; };

  // alertas aparecendo nos polos e sendo resolvidos
  const rings = [];
  if (!parado) {
    setInterval(() => {
      const h = hubs[Math.floor(Math.random() * hubs.length)];
      if (h.alert || h.ok) return;
      h.alert = 1;
      setTimeout(() => { h.alert = 0; h.ok = 1; rings.push({ x: h.x, y: h.y, p: 0 }); setTimeout(() => { h.ok = 0; }, 2200); }, 1800 + Math.random() * 1500);
    }, 900);
  }

  // lista de ocorrências ao lado da marca (só onde existe)
  const feed = document.getElementById('feed');
  const MSGS = [['Desvio de rota', 'Gobrax tratando'], ['Manutenção preventiva', 'Agendada'], ['Motorista em alerta', 'Contato feito'], ['Veículo parado', 'Em acompanhamento'], ['Consumo fora do padrão', 'Ajustado']];
  function addEv() {
    const m = MSGS[Math.floor(Math.random() * MSGS.length)];
    const d = document.createElement('div');
    d.className = 'ev';
    const em = document.createElement('em');
    em.textContent = m[1];
    const sp = document.createElement('span');
    sp.textContent = m[0];
    d.append(document.createElement('i'), sp, em);
    feed.appendChild(d);
    setTimeout(() => { d.classList.add('ok'); em.textContent = 'Resolvido'; }, 1800);
    while (feed.children.length > 3) feed.removeChild(feed.firstChild);
  }
  if (feed && !parado) setTimeout(() => { addEv(); setInterval(addEv, 2600); }, 1700);

  let t = 0;
  function frame() {
    t += 0.016;
    x.fillStyle = '#121212'; x.fillRect(0, 0, W, H);
    let g = x.createRadialGradient(tower.x, tower.y, 0, tower.x, tower.y, W * 0.5);
    g.addColorStop(0, `rgba(${Y},.14)`); g.addColorStop(1, 'transparent');
    x.fillStyle = g; x.fillRect(0, 0, W, H);
    // grade de mapa
    x.strokeStyle = `rgba(${Y},.05)`; x.lineWidth = 1 * dpr; const S = 60 * dpr;
    for (let px = (t * 8 * dpr) % S; px < W; px += S) { x.beginPath(); x.moveTo(px, 0); x.lineTo(px, H); x.stroke(); }
    for (let py = 0; py < H; py += S) { x.beginPath(); x.moveTo(0, py); x.lineTo(W, py); x.stroke(); }
    // anéis da torre
    for (let i = 1; i <= 5; i++) { x.strokeStyle = `rgba(${Y},${0.16 - i * 0.02})`; x.beginPath(); x.arc(tower.x, tower.y, i * Math.min(W, H) * 0.14, 0, 7); x.stroke(); }
    // radar
    const R = Math.min(W, H) * 0.7;
    if (x.createConicGradient) {
      g = x.createConicGradient(t * 0.9, tower.x, tower.y);
      g.addColorStop(0, `rgba(${Y},0)`); g.addColorStop(0.85, `rgba(${Y},0)`); g.addColorStop(1, `rgba(${Y},.22)`);
      x.fillStyle = g; x.beginPath(); x.arc(tower.x, tower.y, R, 0, 7); x.fill();
    }
    // rotas
    routes.forEach((r) => {
      x.strokeStyle = `rgba(${Y},.22)`; x.lineWidth = 1.2 * dpr; x.setLineDash([6 * dpr, 8 * dpr]); x.lineDashOffset = -t * 30 * dpr;
      x.beginPath(); x.moveTo(r.h.x, r.h.y); x.quadraticCurveTo(r.mx, r.my, tower.x, tower.y); x.stroke();
    });
    x.setLineDash([]);
    // veículos com rastro
    trucks.forEach((k) => {
      k.p += k.dir * k.v * 0.016; if (k.p > 1) k.p = 0; if (k.p < 0) k.p = 1;
      for (let j = 0; j < 8; j++) {
        const q = bez(k.r, Math.min(1, Math.max(0, k.p - k.dir * j * 0.008)));
        x.fillStyle = `rgba(${Y},${0.6 * (1 - j / 8)})`; x.beginPath(); x.arc(q.x, q.y, (3 - j * 0.25) * dpr, 0, 7); x.fill();
      }
      const q = bez(k.r, k.p);
      x.shadowColor = `rgb(${Y})`; x.shadowBlur = 12 * dpr; x.fillStyle = '#FFFFFD'; x.beginPath(); x.arc(q.x, q.y, 2.4 * dpr, 0, 7); x.fill(); x.shadowBlur = 0;
    });
    // polos
    hubs.forEach((h) => {
      const col = h.alert ? O : Y;
      if (h.alert) { const p = (t * 1.6) % 1; x.strokeStyle = `rgba(${O},${0.7 * (1 - p)})`; x.lineWidth = 2 * dpr; x.beginPath(); x.arc(h.x, h.y, (6 + 26 * p) * dpr, 0, 7); x.stroke(); }
      x.shadowColor = `rgb(${col})`; x.shadowBlur = (h.alert || h.ok ? 18 : 6) * dpr; x.fillStyle = `rgb(${col})`;
      x.beginPath(); x.arc(h.x, h.y, (h.alert || h.ok ? 5.5 : 3.5) * dpr, 0, 7); x.fill(); x.shadowBlur = 0;
      if (h.ok) { x.strokeStyle = '#121212'; x.lineWidth = 1.8 * dpr; x.beginPath(); x.moveTo(h.x - 2.4 * dpr, h.y); x.lineTo(h.x - 0.6 * dpr, h.y + 2 * dpr); x.lineTo(h.x + 2.8 * dpr, h.y - 2 * dpr); x.stroke(); }
    });
    for (let i = rings.length - 1; i >= 0; i--) {
      const q = rings[i]; q.p += 0.014;
      if (q.p > 1) { rings.splice(i, 1); continue; }
      x.strokeStyle = `rgba(${Y},${0.8 * (1 - q.p)})`; x.lineWidth = 2 * dpr; x.beginPath(); x.arc(q.x, q.y, (6 + 60 * q.p) * dpr, 0, 7); x.stroke();
    }
    // torre central
    const pr = (t * 0.6) % 1;
    x.strokeStyle = `rgba(${Y},${0.5 * (1 - pr)})`; x.lineWidth = 2 * dpr; x.beginPath(); x.arc(tower.x, tower.y, (14 + 50 * pr) * dpr, 0, 7); x.stroke();
    x.shadowColor = `rgb(${Y})`; x.shadowBlur = 26 * dpr; x.fillStyle = `rgb(${Y})`; x.beginPath(); x.arc(tower.x, tower.y, 9 * dpr, 0, 7); x.fill(); x.shadowBlur = 0;
    x.fillStyle = '#121212'; x.beginPath(); x.arc(tower.x, tower.y, 3.5 * dpr, 0, 7); x.fill();
    if (!parado) requestAnimationFrame(frame);
  }
  frame();
  if (parado) addEventListener('resize', frame);
})();

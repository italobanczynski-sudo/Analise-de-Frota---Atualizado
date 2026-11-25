/* =====================================================================
   TELAS · Visão geral · Saúde do cliente · Estudo
   ===================================================================== */
const pc=(c,p)=>(p&&c!=null&&p!=null)?(c-p)/p*100:null;
const hs=s=>{let h=2166136261;for(let i=0;i<s.length;i++){h^=s.charCodeAt(i);h=Math.imul(h,16777619)}return h>>>0};
const pageH=(t,sub,right='')=>`<div class="page-h rise"><div><h1>${t}</h1><p>${sub}</p></div><div class="row">${right}</div></div>`;
const cardH=(t,sub,right='')=>`<div class="card-h"><div><h2>${t}</h2>${sub?`<div class="sub">${sub}</div>`:''}</div>${right?`<div class="row">${right}</div>`:''}</div>`;

function suspCount(C){return suspRecs(C.cid,ST).filter(r=>r.t>=C.P.from&&r.t<=C.P.to).length}
function cmpShort(C){return C.prev.dataDays<C.P.clen*ST.rules.minCmpPct/100}

/* ---------- mapa simplificado (sem Leaflet) ---------- */
function mapSvg(C,h=300){
  const vs=C.fleet.vehicles;let pts='';
  vs.forEach(v=>{const a=hs(v.id),b=hs(v.id+'y');const x=50+(a%500),y=40+(b%(h-80));const st=vehStat(C,v);const col=ST_COL[st];
    pts+=`<g data-tt="<b>${esc(v.id)}</b><br>${esc(v.driver)}<br>${ST_LBL[st]}">${st==='mov'?`<circle class="pulse" cx="${x}" cy="${y}" r="6" fill="${col}"/>`:''}<circle cx="${x}" cy="${y}" r="6" fill="${col}" stroke="var(--card)" stroke-width="2"/></g>`});
  return`<svg viewBox="0 0 600 ${h}" style="width:100%;height:auto;display:block" role="img" aria-label="Mapa simplificado da frota"><rect width="600" height="${h}" fill="var(--soft)"/><path d="M0 ${h*.62} C120 ${h*.3} 220 ${h*.9} 330 ${h*.55} S520 ${h*.25} 600 ${h*.5}" fill="none" stroke="var(--n300)" stroke-width="12" stroke-linecap="round"/>${pts}</svg>`;
}
function statusCounts(C){const o={mov:0,idle:0,stop:0,off:0,unk:0};C.fleet.vehicles.forEach(v=>o[vehStat(C,v)]++);return o}
function statusBadges(sc){return[['ok',sc.mov,'em movimento'],['warn',sc.idle,'motor ligado parado'],['n',sc.stop,'desligados'],['bad',sc.off,'sem sinal'],['n',sc.unk,'sem posição']].filter((x,i)=>x[1]||i<4).map(([c,n,l])=>`<span class="badge ${c}">${n} ${l}</span>`).join('')}

/* ---------- VISÃO GERAL ---------- */
function metricDef(k){return{
  kmpl:{label:'Consumo',get:a=>a.kmpl,fmt:v=>fmt.n(v,2),unit:' km/l',up:true},
  kmday:{label:'Km por dia',get:a=>a.kmDay,fmt:v=>fmt.n(v,0),unit:' km',up:true},
  br100:{label:'Freadas',get:a=>a.br100,fmt:v=>fmt.n(v,1),unit:' /100 km',up:false},
  score:{label:'Nota Gobrax',get:a=>a.score,fmt:v=>fmt.n(v,0),unit:' pts',up:true}}[k]}
function vOverview(C){
  const {cur,prev,P,hl,price}=C;const u=S.u;const md=metricDef(u.ovMetric);
  const extraL=prev.kmpl&&cur.km?cur.l-cur.km/prev.kmpl:null;const extraR=extraL!=null?extraL*price:null;
  const ds=dailySeries(C.cid,ST,P.from,P.to);const wk=weeklySeries(C.cid,ST,P.from,P.to);
  const sc=statusCounts(C);const gl=gainsLosses(cur,prev,price);
  const k=(label,val,delta,note,icon,cls='')=>`<div class="card kpi kp2 lift rise ${cls}"><span class="kico k-${icon}" aria-hidden="true">${ico(icon,26)}</span><div class="tag">${label}</div><div class="v">${val}</div><div>${delta} <span class="dl-n">${note||''}</span></div></div>`;
  const kpis=[
    k('Média de consumo',count(cur.kmpl,2)+'<small>km/l</small>',dl(pc(cur.kmpl,prev.kmpl),true),'vs '+(S.compare==='mes'?'mês anterior':'período anterior'),'fuel'),
    k(extraR!=null&&extraR<0?'Diesel economizado':'Diesel a mais',extraR!=null?count(Math.abs(extraR),0,'R$ '):'—',extraL!=null?`<span class="dl ${extraL>0?'bad':'good'}">${fmt.n(Math.abs(extraL),0)} L</span>`:dl(null),'nos mesmos km','dollar'),
    k('Km por dia trabalhado',count(cur.kmDay,0)+'<small>km</small>',dl(pc(cur.kmDay,prev.kmDay),true),'vs anterior','route'),
    k('Frota ativa',count(cur.active,0)+`<small>de ${cur.total}</small>`,`<span class="dl ${cur.noSignal.length?'bad':'good'}">${cur.noSignal.length} sem sinal</span>`,'há mais de 3 dias','truck')
  ].join('');
  const dsv=ds.map(d=>d.a?md.get(d.a):null);const pv=md.get(prev);
  const chart=lineChart({labels:ds.map(d=>fmt.dm(d.date)),series:[{vals:dsv,color:'var(--laranja)',area:true,dots:ds.length<=31}],base:pv!=null?{v:pv,label:'média do período de comparação '+md.fmt(pv)}:null,yfmt:v=>md.fmt(v),
    tip:i=>ds[i].a?`<b>${fmt.dm(ds[i].date)}</b><br>${md.label}: ${md.fmt(dsv[i])}${md.unit}<br>${fmt.n(ds[i].a.km,0)} km · ${ds[i].a.n} veículos`:`<b>${fmt.dm(ds[i].date)}</b><br>Sem rodagem`,aria:'Evolução diária de '+md.label})+`<div class="tag" style="margin:18px 0 2px">Km rodados por dia</div>`+dailyBars(ds.map(d=>d.a?d.a.km:null),ds.map(d=>fmt.dm(d.date)),{avg:cur.km/Math.max(1,cur.dataDays),u:' km',h:200,aria:'Km rodados por dia'});
  const maxAbs=Math.max(1,...gl.map(g=>Math.abs(g.money)));
  const top=[...gl.slice(0,4),...gl.slice(-4).filter(g=>g.money<0)].filter((g,i,a)=>a.findIndex(x=>x.id===g.id)===i);
  const sumG=gl.filter(g=>g.money>0).reduce((a,g)=>a+g.money,0),sumL=gl.filter(g=>g.money<0).reduce((a,g)=>a+g.money,0);
  const bars=top.length?top.map(g=>{const w=Math.min(50,Math.abs(g.money)/maxAbs*50);return`<div class="brow" data-act="drv-open" data-id="${esc(g.id)}"><div class="bl">${esc(g.id)}</div><div class="btrack"><div class="bfill ${g.money<0?'neg':''}" style="left:${g.money>0?50:50-w}%;width:${w}%;background:${g.money>0?'var(--ok)':'var(--bad)'}"></div><div class="bmid"></div></div><div class="bv" style="color:${g.money>0?'var(--ok-t)':'var(--bad-t)'}">${g.money>0?'+':'−'}${fmt.brl(Math.abs(g.money))}</div></div>`}).join(''):`<div class="empty"><b>Sem base para comparar</b>O período de comparação não tem dados por veículo.</div>`;
  const pts=wk.map(w=>`<div class="wkc"><div class="xs muted">Sem ${w.n}${w.partial?'*':''} · ${fmt.dm(dIso(w.from))}–${fmt.dm(dIso(w.to))}</div><b class="num" style="font-size:18px">${fmt.kmpl(w.a.kmpl)}</b></div>`).join('');
  return`${pageH('Visão geral',`${esc(C.def.name)} · ${P.label} · comparado ao ${P.cmpLabel}`,`<button class="btn ghost" data-act="print">${ico('download',16)} Exportar PDF</button>`)}
  <div class="grid g-kpi">
    <div class="card brand kpi lift rise" style="overflow:hidden"><div class="swoosh"></div><div class="tag" style="color:var(--n600)">Saúde do cliente</div><div class="row" style="flex-wrap:nowrap;gap:14px;align-items:flex-end">${gauge(hl.idx,hl.band)}</div><div>${bandBadge(hl.band)}</div><div class="sm" style="color:var(--n600)">${hl.top?`Pesa mais: ${hl.top.label}.`:''}</div><a class="b sm" href="#/retencao">Abrir saúde do cliente</a></div>
    ${kpis}
  </div>
  <div class="sec">
    <div class="card dark hero rise"><span class="wm" aria-hidden="true">gobrax</span>${cardH('Evolução dia a dia',`${md.label} contra a média do período de comparação`,segHtml([['kmpl','Consumo'],['kmday','Km por dia'],['br100','Freadas'],['score','Nota']],u.ovMetric,'ov-metric'))}${chart}<div class="row" style="margin-top:14px;display:grid;grid-template-columns:repeat(auto-fit,minmax(150px,1fr))">${pts}</div></div>
  </div>
  <div class="grid g-main sec">
    <div class="card rise">${cardH('Quem puxa e quem derruba o resultado',`Diesel economizado ou gasto a mais nos mesmos km, a ${fmt.brl(price,2)}/L`,`<a class="b sm" href="#/motoristas">Abrir motoristas</a>`)}<div class="bars">${bars}</div><div class="row sm" style="margin-top:16px;padding-top:14px;border-top:1px solid var(--soft);gap:22px"><span><b style="color:var(--ok-t)">+${fmt.brl(sumG)}</b> ganhos</span><span><b style="color:var(--bad-t)">−${fmt.brl(Math.abs(sumL))}</b> perdas</span><span><b>${fmt.brl(sumG+sumL)}</b> saldo</span></div></div>
    <div class="card rise">${cardH('Frota agora',`${cur.active} de ${cur.total} com rodagem no período`,`<a class="b sm" href="#/frota?aba=mapa">Ver mapa</a>`)}${liveNote(C)}${miniMapBlock(C,240)}<div class="row sm" style="margin-top:12px;gap:10px">${statusBadges(sc)}</div></div>
  </div>
  ${ovExtra(C)}`;
}

/* ---------- SAÚDE DO CLIENTE ---------- */
async function vRetention(C){
  const u=S.u;const tabs=tabsHtml([['cliente','Cliente'],['carteira','Carteira de clientes']],u.retTab,'ret-tab');
  if(u.retTab==='carteira')return pageH('Saúde dos clientes','Risco de cancelamento de cada cliente, calculado com os resultados Gobrax da frota')+tabs+await vPortfolio(C);
  const cid=C.cid,start=dIdx(C.def.start);const from=Math.max(start,T0-59);
  const days=[];for(let t=from;t<=T0;t++)days.push(t);
  const vals=days.map(t=>metrics(cid,ST,Math.max(start,t-6),t).kmpl);
  const chart=lineChart({labels:days.map(t=>fmt.dm(dIso(t))),series:[{vals,color:'var(--laranja)',area:true}],h:340,yfmt:v=>fmt.n(v,2),tip:i=>vals[i]!=null?`<b>${fmt.dm(dIso(days[i]))}</b><br>Média de 7 dias: ${fmt.n(vals[i],2)} km/l`:'',aria:'Consumo, média móvel de 7 dias'});
  const first=metrics(cid,ST,start,Math.min(T0,start+13)),last=metrics(cid,ST,T0-13,T0);const chg=first&&first.kmpl?pc(last.kmpl,first.kmpl):null;
  const hl=C.hl;const sig=hl.signals.map(s=>`<div class="stack" style="gap:4px"><div class="row" style="justify-content:space-between;flex-wrap:nowrap"><b>${s.label}</b><span class="muted xs">${s.risk!=null?`peso ${fmt.n(s.share,0)}% · risco ${fmt.n(s.risk,0)}`:'sem dado'}</span></div><div class="btrack" style="height:8px"><div class="bfill" style="left:0;width:${s.risk!=null?s.risk:0}%;background:${s.risk>=60?'var(--bad)':s.risk>=35?'var(--warn)':'var(--ok)'}"></div></div><div class="xs muted">${esc(s.txt)}</div></div>`).join('');
  const st=study(cid,ST,C.price);const ratio=st.inv?clamp(st.econ/st.inv,0,3):0;
  const al=buildAlerts(C).filter(x=>!x.act);
  const alh=al.length?al.map(x=>`<a class="li" href="#/${x.r}" style="text-decoration:none;color:inherit"><div class="ava">${ico(x.ic,16)}</div><div style="flex:1"><b class="sm">${x.t}</b><div class="xs muted">${esc(x.s)}</div></div>${ico('chev',16)}</a>`).join(''):'<div class="empty"><b>Sem sinais de atenção</b></div>';
  const mini=(t,v,bg)=>`<div class="card" style="padding:14px;box-shadow:none;background:${bg||'var(--soft)'}"><div class="xs muted">${t}</div><b class="num" style="font-size:22px">${v}</b></div>`;
  return`${pageH('Saúde do cliente · '+esc(C.def.name),`${C.P.label} contra ${C.P.cmpLabel}`,`<button class="btn ghost" data-act="print">${ico('download',16)} Exportar PDF</button>`)}
  ${tabs}
  <div class="grid g-main top">
    <div class="card rise">${cardH('Consumo desde o início do acompanhamento','Média móvel de 7 dias, em km/l')}${chart}
      <div class="grid g3" style="margin-top:14px;gap:12px">${mini('Primeiras 2 semanas',first&&first.kmpl?fmt.n(first.kmpl,2)+' km/l':'—')}${mini('Últimas 2 semanas',fmt.n(last.kmpl,2)+' km/l')}${mini('Mudança no período',dl(chg,true),chg==null?'var(--soft)':chg>=0?'var(--ok-s)':'var(--bad-s)')}</div></div>
    <div class="card rise">${cardH('Índice de saúde',`${C.P.label} contra ${C.P.cmpLabel}`,bandBadge(hl.band))}<div class="row" style="justify-content:center">${gauge(hl.idx,hl.band)}</div><div class="stack" style="margin-top:14px">${sig}</div><details style="margin-top:14px"><summary class="b sm" style="cursor:pointer">Como o índice é calculado</summary><ol class="sm muted" style="padding-left:18px"><li>Cada sinal vira um risco de 0 a 100 (0 é tudo bem).</li><li>O índice é a média ponderada dos riscos; sinal sem dado sai e os pesos se redistribuem.</li><li>Faixa: baixo abaixo de ${ST.rules.bandMedio}, médio de ${ST.rules.bandMedio} a ${ST.rules.bandAlto-1}, alto a partir de ${ST.rules.bandAlto}.</li></ol><div class="xs muted">É um sinal para priorizar a conversa com o cliente, não uma previsão.</div></details></div>
  </div>
  <div class="grid g2 sec top">
    <div class="card brand rise"><div class="swoosh"></div>${cardH('O projeto se paga?',`Desde ${fmt.dm(dIso(st.ps))}`)}<div class="stack"><div><div class="muted sm">Economia de diesel acumulada</div><b class="num" style="font-size:30px;color:${st.econ>=0?'var(--ok-t)':'var(--bad-t)'}">${fmt.brl(st.econ)}</b></div><div><div class="muted sm">Investido na Gobrax</div><b class="num" style="font-size:30px">${fmt.brl(st.inv)}</b></div><div class="btrack" style="height:12px;background:var(--a200)"><div class="bfill" style="left:0;width:${clamp(ratio/3*100,0,100)}%;background:${st.pays?'var(--ok)':'var(--bad)'}"></div><div style="position:absolute;top:-4px;bottom:-4px;left:33.3%;width:3px;background:var(--text)"></div></div><div class="sm" style="color:var(--n600)">${st.pays?`A economia cobre a mensalidade ${fmt.n(st.roi,1)} vez(es).`:`Faltam ${fmt.brl(st.miss)} para a economia cobrir o investimento.`}</div><a class="b sm" href="#/estudo">Abrir estudo do projeto</a></div></div>
    <div class="card rise">${cardH('Sinais de atenção','O que merece olhar agora neste cliente')}<div class="stack" style="gap:2px">${alh}</div></div>
  </div>`;
}
async function vPortfolio(C){
  await Promise.all(CLIENT_DEFS.map(c=>DB.load(c.id)));const P=C.P;const rows=CLIENT_DEFS.map(def=>{
    const cur=metrics(def.id,ST,P.from,P.to),prev=metrics(def.id,ST,P.cf,P.ct);const h=health(cur,prev,ST.rules);const short=prev.dataDays<prev.len*ST.rules.minCmpPct/100;
    return`<tr class="click" data-act="client-open" data-id="${esc(def.id)}"><td><b>${esc(def.name)}</b><div class="xs muted">${def.vehicles} veículos</div></td><td>${bandBadge(h.band)} <b class="num">${h.idx}</b></td><td class="r">${dl(pc(cur.kmpl,prev.kmpl),true)}</td><td class="r">${short?'<span class="xs muted" data-tt="Base de comparação curta">base curta</span>':dl(pc(cur.km/cur.len,prev.km/prev.len),true)}</td><td class="r num">${fmt.n(cur.score,0)}</td><td class="r">${ico('chev',16)}</td></tr>`}).join('');
  return`<div class="card rise">${cardH('Carteira de clientes',`${P.label} contra ${P.cmpLabel}`)}<div class="scroll-x"><table class="tbl"><thead><tr><th>Cliente</th><th>Risco</th><th class="r">Consumo</th><th class="r">Km por dia</th><th class="r">Nota</th><th></th></tr></thead><tbody>${rows}</tbody></table></div></div>`;
}

/* ---------- ESTUDO ---------- */
function vStudy(C){
  const cid=C.cid,st=study(cid,ST,C.price);const ct=st.ct;
  const lo=Math.min(0,...st.weeks.map(w=>w.cum)),hi=Math.max(...st.weeks.map(w=>Math.max(w.cum,w.cumI)),1);
  const chart=lineChart({labels:st.weeks.map(w=>w.label),series:[{vals:st.weeks.map(w=>w.cum),color:st.econ>=0?'var(--ok)':'var(--bad)',area:true,dots:true},{vals:st.weeks.map(w=>w.cumI),color:'var(--text)',dash:'6 5',w:2.5}],ymin:lo*1.1,ymax:hi*1.15,yfmt:v=>fmt.n(v/1000,0)+' mil',tip:i=>`<b>Até ${st.weeks[i].label}</b><br>Economia: ${fmt.brl(st.weeks[i].cum)}<br>Investimento: ${fmt.brl(st.weeks[i].cumI)}`,aria:'Economia acumulada contra investimento acumulado, por semana'});
  const rows=st.months.map(m=>`<tr><td><b>${m.label}</b></td><td class="r num">${fmt.brl(m.inv)}</td><td class="r num b" style="color:${m.econ>=0?'var(--ok-t)':'var(--bad-t)'}">${fmt.brl(m.econ)}</td><td class="r num">${fmt.kmpl(m.kmpl)} km/l</td><td class="r num">${fmt.n(m.score,0)}</td><td class="r num">${fmt.n(m.km,0)}</td></tr>`).join('');
  const chg=pc(st.second.kmpl,st.base);
  const verdict=st.pays?`Sobram <span class="num" style="color:var(--amarelo)">${fmt.brl(st.econ-st.inv)}</span> depois de pagar o investimento`:`Faltam <span class="num" style="color:var(--amarelo)">${fmt.brl(st.miss)}</span> para a economia cobrir o investimento`;
  return`${pageH('Estudo do projeto',`${esc(C.def.name)} · ${fmt.dm(dIso(st.ps))} a ${fmt.dm(TODAY)} · ${st.len} dias · diesel a ${fmt.brl(C.price,2)}/L`,`<button class="btn ghost" data-act="print">${ico('download',16)} Exportar PDF</button>`)}
  <div class="grid g-main">
    <div class="card dark rise stack" style="padding:28px;position:relative;overflow:hidden"><div class="swoosh" style="border-color:rgba(254,219,0,.12)"></div><div class="row"><span class="badge ${st.pays?'ok':'bad'}">${st.pays?'O PROJETO SE PAGA':'AINDA NÃO SE PAGA'}</span></div><div><div style="color:#e1e1dd;font-size:16px">O projeto se paga?</div><div style="font-size:34px;font-weight:900;letter-spacing:-.01em;line-height:1.15">${verdict}</div></div>
      <div class="row" style="gap:32px"><div><div class="muted">Investido no período</div><b class="num" style="font-size:26px">${count(st.inv,0,'R$ ')}</b></div><div><div class="muted">${st.econ>=0?'Diesel economizado':'Diesel gasto a mais'}</div><b class="num" style="font-size:26px;color:${st.econ>=0?'#7be0ac':'#ff9d92'}">${fmt.brl(st.econ)}</b></div><div><div class="muted">Ritmo atual por mês</div><b class="num" style="font-size:26px;color:${st.pace>=0?'#7be0ac':'#ff9d92'}">${fmt.brl(st.pace)}</b></div><div><div class="muted">Retorno sobre o investimento</div><b class="num" style="font-size:26px">${st.roi!=null?fmt.n(st.roi,1)+'×':'—'}</b></div></div>
      <div class="muted sm">Economia calculada contra o consumo do início do acompanhamento (${fmt.n(st.base,2)} km/l, primeira metade do período), nos mesmos km rodados.${REAL.on?' O servidor guarda o último ano de dados: o início é o primeiro dia com registro dentro dessa janela.':''}</div></div>
    <div class="card rise stack">${cardH('Contrato do cliente','')}<div class="kv"><span>Valor por unidade</span><b class="num">${fmt.brl(ct.unit,2)}/mês</b></div><div class="kv"><span>Unidades consideradas</span><b class="num">${ct.units}</b></div><div style="background:var(--soft);border-radius:12px;padding:14px 16px" class="row"><span class="muted">Mensalidade</span><b class="num" style="font-size:24px;margin-left:auto">${fmt.brl(st.mens,2)}</b></div>${REAL.on?'<div class="xs muted">Unidades = veículos com dado na API. Valor por unidade é a premissa de Regras e integração.</div>':''}${ct.units!==C.fleet.vehicles.length?`<div class="banner" style="margin:0">Os dados trazem ${C.fleet.vehicles.length} veículos e o contrato diz ${ct.units}.</div>`:''}</div>
  </div>
  <div class="sec">
    <div class="card rise">${cardH('Economia acumulada contra investimento acumulado','Em reais, semana a semana desde o início do acompanhamento')}${chart}<div class="leg" style="margin-top:10px"><span><i style="background:${st.econ>=0?'var(--ok)':'var(--bad)'}"></i>Economia acumulada</span><span><i style="background:var(--text)"></i>Investimento acumulado</span></div></div>
  </div>
  <div class="card sec rise">${cardH('Mês a mês','Investimento, economia, consumo, nota e km')}<div class="scroll-x"><table class="tbl"><thead><tr><th>Mês</th><th class="r">Investimento</th><th class="r">Economia</th><th class="r">Consumo</th><th class="r">Nota</th><th class="r">Km</th></tr></thead><tbody>${rows}<tr><td><b>Total</b></td><td class="r num b">${fmt.brl(st.inv)}</td><td class="r num b">${fmt.brl(st.econ)}</td><td class="r num">${fmt.kmpl(st.all.kmpl)} km/l</td><td class="r num">${fmt.n(st.all.score,0)}</td><td class="r num">${fmt.n(st.all.km,0)}</td></tr></tbody></table></div>
    <div class="row sm muted" style="margin-top:14px;justify-content:space-between"><span><b>Como calculamos:</b> litros que o consumo do início gastaria nos mesmos km, menos os litros realmente gastos, vezes o preço do diesel. O investimento é a mensalidade proporcional aos dias de cada mês.</span><span>Consumo do início → agora: <b>${fmt.n(st.base,2)} → ${fmt.n(st.second.kmpl,2)} km/l</b> ${dl(chg,true)} · detalhes na <a href="#/geral">Visão geral</a></span></div></div>`;
}

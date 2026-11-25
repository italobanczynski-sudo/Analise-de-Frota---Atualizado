/* =====================================================================
   TELAS · Visão geral · Saúde do cliente · Estudo
   ===================================================================== */
const pc=(c,p)=>(p&&c!=null&&p!=null)?(c-p)/p*100:null;
const hs=s=>{let h=2166136261;for(let i=0;i<s.length;i++){h^=s.charCodeAt(i);h=Math.imul(h,16777619)}return h>>>0};
const pageH=(t,sub,right='')=>`<div class="page-h rise"><div><h1>${t}</h1><p>${sub}</p></div><div class="row">${right}</div></div>`;
/* "!" no canto de um cartão: abre a explicação de como o número é calculado (sem JavaScript, com <details>) */
const infoDot=(titulo,html,o={})=>`<details class="info-dot${o.dark?' dk':''}${o.left?' lf':''}"><summary title="${esc(titulo)}" aria-label="${esc(titulo)}">!</summary><div class="info-pop"><b>${esc(titulo)}</b>${html}</div></details>`;
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
function statusBadges(sc){return[['ok',sc.mov,'em movimento'],['warn',sc.idle,'motor ligado parado'],['n',sc.stop,'desligados'],['n',sc.unk,'sem posição']].filter((x,i)=>x[1]||i<3).map(([c,n,l])=>`<span class="badge ${c}">${n} ${l}</span>`).join('')}

/* ---------- VISÃO GERAL ---------- */
function metricDef(k){return{
  kmpl:{label:'Consumo',get:a=>a.kmpl,fmt:v=>fmt.n(v,2),unit:' km/l',up:true},
  kmday:{label:'Km por dia',get:a=>a.kmDay,fmt:v=>fmt.n(v,0),unit:' km',up:true},
  br100:{label:'Freadas',get:a=>a.br100,fmt:v=>fmt.n(v,1),unit:' /100 km',up:false},
  score:{label:'Nota Gobrax',get:a=>a.score,fmt:v=>fmt.n(v,0),unit:' pts',up:true}}[k]}
function vOverview(C){
  const {cur,prev,P,hl,price}=C;const u=S.u;
  /* km por dia já está nas barras logo abaixo: o gráfico de cima mostra consumo, freadas ou nota */
  if(!['kmpl','br100','score'].includes(u.ovMetric))u.ovMetric='kmpl';const md=metricDef(u.ovMetric);
  const ds=dailySeries(C.cid,ST,P.from,P.to);const wk=weeklySeries(C.cid,ST,P.from,P.to);
  const sc=statusCounts(C);const gl=C.gl;const ec=C.ec;
  const k=(label,val,delta,note,icon,cls='')=>`<div class="card kpi kp2 lift rise ${cls}"><span class="kico k-${icon}" aria-hidden="true">${ico(icon,26)}</span><div class="tag">${label}</div><div class="v">${val}</div><div>${delta} <span class="dl-n">${note||''}</span></div></div>`;
  const kpis=[
    k('Média de consumo',count(cur.kmpl,2)+'<small>km/l</small>',dl(pc(cur.kmpl,prev.kmpl),true),'vs '+(S.compare==='mes'?'mês anterior':'período anterior'),'fuel'),
    k(ec.money!=null&&ec.money<0?'Diesel gasto a mais':'Diesel economizado',ec.money!=null?count(Math.abs(ec.money),0,'R$ '):'sem dado',ec.lit!=null?`<span class="dl ${ec.lit>=0?'good':'bad'}">${fmt.n(Math.abs(ec.lit),0)} L</span>`:'',`vs. linha de base do projeto ${confBadge(C)}`,'dollar'),
    k('Km por dia trabalhado',count(cur.kmDay,0)+'<small>km</small>',dl(pc(cur.kmDay,prev.kmDay),true),'vs anterior','route'),
    k('Frota ativa',count(cur.active,0)+`<small>de ${cur.total}</small>`,`<span class="dl ${C.eq.semTransmitir.length?'bad':'good'}">${C.eq.semTransmitir.length} sem transmitir</span>`,'nenhum dado nos últimos 3 dias','truck')
  ].join('');
  const dsv=ds.map(d=>d.a?md.get(d.a):null);const pv=md.get(prev);
  const chart=lineChart({labels:ds.map(d=>fmt.dm(d.date)),series:[{vals:dsv,color:'var(--laranja)',area:true,dots:ds.length<=31}],base:pv!=null?{v:pv,label:'média do período de comparação '+md.fmt(pv)}:null,yfmt:v=>md.fmt(v),
    tip:i=>ds[i].a?`<b>${fmt.dm(ds[i].date)}</b><br>${md.label}: ${md.fmt(dsv[i])}${md.unit}<br>${fmt.n(ds[i].a.km,0)} km · ${ds[i].a.n} veículos`:`<b>${fmt.dm(ds[i].date)}</b><br>Sem rodagem`,aria:'Evolução diária de '+md.label})+`<div class="tag" style="margin:18px 0 2px">Km rodados por dia</div>`+dailyBars(ds.map(d=>d.a?d.a.km:null),ds.map(d=>fmt.dm(d.date)),{avg:cur.km/Math.max(1,cur.dataDays),u:' km',h:200,aria:'Km rodados por dia'});
  const maxAbs=Math.max(1,...gl.map(g=>Math.abs(g.money)));
  const top=[...gl.slice(0,4),...gl.slice(-4).filter(g=>g.money<0)].filter((g,i,a)=>a.findIndex(x=>x.id===g.id)===i);
  const sumG=gl.filter(g=>g.money>0).reduce((a,g)=>a+g.money,0),sumL=gl.filter(g=>g.money<0).reduce((a,g)=>a+g.money,0);
  const bars=top.length?top.map(g=>{const w=Math.min(50,Math.abs(g.money)/maxAbs*50);return`<div class="brow" data-act="drv-open" data-id="${esc(g.id)}"><div class="bl">${esc(g.id)}</div><div class="btrack"><div class="bfill ${g.money<0?'neg':''}" style="left:${g.money>0?50:50-w}%;width:${w}%;background:${g.money>0?'var(--ok)':'var(--bad)'}"></div><div class="bmid"></div></div><div class="bv" style="color:${g.money>0?'var(--ok-t)':'var(--bad-t)'}">${g.money>0?'+':'−'}${fmt.brl(Math.abs(g.money))}</div></div>`}).join(''):`<div class="empty"><b>Sem dado</b>Sem linha de base do projeto: a economia não é estimada.</div>`;
  const pts=wk.map(w=>`<div class="wkc"${w.partial?' style="opacity:.6"':''}><div class="xs muted">Sem ${w.n} · ${fmt.dm(dIso(w.from))}–${fmt.dm(dIso(w.to))}</div><b class="num" style="font-size:18px${w.partial?';color:var(--n400)':''}">${fmt.kmpl(w.a.kmpl)}</b>${w.partial?`<div class="xs" style="color:var(--n400)">parcial (${w.dias} dia${w.dias>1?'s':''})</div>`:''}</div>`).join('');
  return`${pageH('Visão geral',`${esc(C.def.name)} · ${P.label} · comparado ao ${P.cmpLabel}`,`<button class="btn ghost" data-act="print">${ico('download',16)} Exportar PDF</button>`)}
  <div class="grid g-kpi">
    <div class="card brand kpi lift rise" style="overflow:hidden;text-align:center"><div class="swoosh"></div><div class="tag" style="color:var(--n600)">Nota Gobrax · Premiação</div>${gaugeNota(cur.score,'PERFORMANCE MÉDIA')}<div class="xs" style="color:var(--n600)">${(()=>{const pf=cur.prem;if(!REAL.on)return'Média do período';if(!pf)return'Sem premiação da API no período';return(pf.stale?'Último mês: ':'')+pf.meses.map(m=>MESES[+m.slice(5)-1]+(m===TODAY.slice(0,7)?' (parcial)':'')).join(', ')+' · '+pf.motoristas+' motoristas'})()}</div><a class="b sm" href="#/retencao">Saúde do cliente: ${count(hl.idx,0)} · risco ${String(hl.band).toLowerCase()}</a></div>
    ${kpis}
  </div>
  <div class="sec">
    <div class="card dark hero rise"><span class="wm" aria-hidden="true">gobrax</span>${cardH('Evolução dia a dia',`${md.label} contra a média do período de comparação`,segHtml([['kmpl','Consumo'],['br100','Freadas'],['score','Nota']],u.ovMetric,'ov-metric'))}${chart}<div class="row" style="margin-top:14px;display:grid;grid-template-columns:repeat(auto-fit,minmax(150px,1fr))">${pts}</div></div>
  </div>
  <div class="grid g-main sec">
    <div class="card rise">${cardH('Quem puxa e quem derruba o resultado',`Contribuição de cada veículo para o diesel economizado, ${esc(ec.b.rotulo)}, a ${fmt.brl(price,2)}/L`,`<a class="b sm" href="#/motoristas">Abrir motoristas</a>`)}<div class="bars">${bars}</div><div class="row sm" style="margin-top:16px;padding-top:14px;border-top:1px solid var(--soft);gap:22px"><span><b style="color:var(--ok-t)">+${fmt.brl(sumG)}</b> ganhos</span><span><b style="color:var(--bad-t)">−${fmt.brl(Math.abs(sumL))}</b> perdas</span><span><b>${fmt.brl(sumG+sumL)}</b> saldo (= diesel economizado)</span></div>${concBlock(C)}</div>
    <div class="card rise">${cardH('Frota agora',`${cur.active} de ${cur.total} com rodagem no período`,`<a class="b sm" href="#/frota?aba=mapa">Ver mapa</a>`)}${liveNote(C)}${miniMapBlock(C,240)}<div class="row sm" style="margin-top:12px;gap:10px">${statusBadges(sc)}</div><div class="row sm" style="margin-top:8px;gap:8px;align-items:center">${eqLine(C)}${eqInfo()}</div></div>
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
  const st=C.st;const ratio=!st.noBase&&st.inv?clamp(st.econ/st.inv,0,3):0;
  const al=buildAlerts(C).filter(x=>!x.act);
  const alh=al.length?al.map(x=>`<a class="li" href="#/${x.r}" style="text-decoration:none;color:inherit"><div class="ava">${ico(x.ic,16)}</div><div style="flex:1"><b class="sm">${x.t}</b><div class="xs muted">${esc(x.s)}</div></div>${ico('chev',16)}</a>`).join(''):'<div class="empty"><b>Sem sinais de atenção</b></div>';
  const mini=(t,v,bg)=>`<div class="card" style="padding:14px;box-shadow:none;background:${bg||'var(--soft)'}"><div class="xs muted">${t}</div><b class="num" style="font-size:22px">${v}</b></div>`;
  return`${pageH('Saúde do cliente · '+esc(C.def.name),`${C.P.label} contra ${C.P.cmpLabel}`,`<button class="btn ghost" data-act="print">${ico('download',16)} Exportar PDF</button>`)}
  ${tabs}
  <div class="grid g-main top">
    <div class="card rise">${cardH('Consumo desde o início do acompanhamento','Média móvel de 7 dias, em km/l')}${chart}
      <div class="grid g3" style="margin-top:14px;gap:12px">${mini('Primeiras 2 semanas',first&&first.kmpl?fmt.n(first.kmpl,2)+' km/l':'—')}${mini('Últimas 2 semanas',fmt.n(last.kmpl,2)+' km/l')}${mini('Mudança no período',dl(chg,true),chg==null?'var(--soft)':chg>=0?'var(--ok-s)':'var(--bad-s)')}</div></div>
    <div class="card rise">${cardH('Índice de saúde',`${C.P.label} contra ${C.P.cmpLabel}`,bandBadge(hl.band))}<div class="row" style="justify-content:center">${gauge(hl.idx,hl.band)}</div>${hl.piso?`<div class="xs" style="text-align:center;color:var(--warn-t);margin-top:6px"><b>Elevado para risco médio</b>: o projeto ainda não se paga (índice calculado: ${hl.calc}).</div>`:''}<div class="stack" style="margin-top:14px">${sig}</div><details style="margin-top:14px"><summary class="b sm" style="cursor:pointer">Como o índice é calculado</summary><ol class="sm muted" style="padding-left:18px"><li>Cada sinal vira um risco de 0 a 100 (0 é tudo bem).</li><li>O índice é a média ponderada dos riscos; sinal sem dado sai e os pesos se redistribuem.</li><li>Faixa: baixo abaixo de ${ST.rules.bandMedio}, médio de ${ST.rules.bandMedio} a ${ST.rules.bandAlto-1}, alto a partir de ${ST.rules.bandAlto}.</li><li>Cobertura do investimento (economia acumulada ÷ investimento, vs. linha de base do projeto): 1,2× ou mais = risco 0; de 1,0× a 1,2× = de 30 a 0; abaixo de 1,0× = 50 + 150 × (1 − cobertura).</li><li>Projeto que ainda não se paga fica no mínimo em risco médio.</li></ol><div class="xs muted">É um sinal para priorizar a conversa com o cliente, não uma previsão.</div></details></div>
  </div>
  <div class="grid g2 sec top">
    <div class="card brand rise"><div class="swoosh"></div>${cardH('O projeto se paga?',`Desde ${fmt.dm(dIso(st.ps))} · ${esc(C.ec.b.rotulo)}`)}${st.noBase?'<div class="empty"><b>Sem dado</b>Sem linha de base: o início do acompanhamento não tem consumo válido.</div>':`<div class="stack"><div><div class="muted sm">Economia de diesel acumulada</div><b class="num" style="font-size:30px;color:${st.econ>=0?'var(--ok-t)':'var(--bad-t)'}">${fmt.brl(st.econ)}</b></div><div><div class="muted sm">Investido na Gobrax</div><b class="num" style="font-size:30px">${fmt.brl(st.inv)}</b></div><div class="btrack" style="height:12px;background:var(--a200)"><div class="bfill" style="left:0;width:${clamp(ratio/3*100,0,100)}%;background:${st.pays?'var(--ok)':'var(--bad)'}"></div><div style="position:absolute;top:-4px;bottom:-4px;left:33.3%;width:3px;background:var(--text)"></div></div><div class="sm" style="color:var(--n600)">${st.pays?`A economia cobre a mensalidade ${fmt.n(st.roi,1)} vez(es).`:`Faltam ${fmt.brl(st.miss)} para a economia cobrir o investimento.`}</div><a class="b sm" href="#/estudo">Abrir estudo do projeto</a></div>`}</div>
    <div class="card rise">${cardH('Sinais de atenção','O que merece olhar agora neste cliente')}<div class="stack" style="gap:2px">${alh}</div></div>
  </div>`;
}
async function vPortfolio(C){
  await Promise.all(CLIENT_DEFS.map(c=>DB.load(c.id)));const P=C.P;const rows=CLIENT_DEFS.map(def=>{
    const cur=metrics(def.id,ST,P.from,P.to),prev=metrics(def.id,ST,P.cf,P.ct);const sx=study(def.id,ST,dieselDe(def));const h=health(cur,prev,ST.rules,sx.noBase?null:{econ:sx.econ,inv:sx.inv});const short=prev.dataDays<prev.len*ST.rules.minCmpPct/100;
    return`<tr class="click" data-act="client-open" data-id="${esc(def.id)}"><td><b>${esc(def.name)}</b><div class="xs muted">${def.vehicles} veículos</div></td><td>${bandBadge(h.band)} <b class="num">${h.idx}</b></td><td class="r">${dl(pc(cur.kmpl,prev.kmpl),true)}</td><td class="r">${short?'<span class="xs muted" data-tt="Base de comparação curta">base curta</span>':dl(pc(cur.km/cur.len,prev.km/prev.len),true)}</td><td class="r num">${fmt.n(cur.score,0)}</td><td class="r">${ico('chev',16)}</td></tr>`}).join('');
  return`<div class="card rise">${cardH('Carteira de clientes',`${P.label} contra ${P.cmpLabel}`)}<div class="scroll-x"><table class="tbl"><thead><tr><th>Cliente</th><th>Risco</th><th class="r">Consumo</th><th class="r">Km por dia</th><th class="r">Nota</th><th></th></tr></thead><tbody>${rows}</tbody></table></div></div>`;
}

/* ---------- ESTUDO ---------- */
const podeEditarContrato=()=>!REAL.on||(REAL.eu&&REAL.eu.papel!=='leitor');
function contratoForm(C,ct,mens){
  const ed=podeEditarContrato(),dis=ed?'':' disabled';const pm=premissasDe(C.def);
  const campo=(id,rot,val,ph,suf)=>`<div class="field" style="flex:1 1 150px"><label for="${id}">${rot}</label><div class="row" style="flex-wrap:nowrap;gap:6px"><input class="inp num" id="${id}" inputmode="decimal" value="${val>0?esc(fmt.n(val,id==='ct-valor'?2:0)):''}" placeholder="${esc(ph)}"${dis}><span class="xs muted" style="white-space:nowrap">${suf}</span></div></div>`;
  const nota=[ct.unitAuto?'valor padrão de '+fmt.brl(ct.unit,2)+' por unidade':'',ct.unitsAuto?'unidades = '+ct.units+' veículos com dado na API':''].filter(Boolean).join(' · ');
  return`<div class="card rise stack">${cardH('Contrato do cliente',ed?'Edite e salve: vale aqui, no painel clássico e em todas as contas.':'Só CS e administrador mudam o contrato.')}
    <div class="row" style="align-items:flex-end;gap:12px">${campo('ct-valor','Valor por unidade',pm.valorEquip,fmt.n(ct.unit,2),'R$/mês')}${campo('ct-qtd','Unidades contratadas',pm.qtd,String(ct.units),'un.')}</div>
    <div style="background:var(--soft);border-radius:12px;padding:14px 16px" class="row"><span class="muted">Mensalidade</span><b class="num" style="font-size:24px;margin-left:auto">${fmt.brl(mens,2)}</b></div>
    ${nota?`<div class="xs muted">${esc(nota)}: informe o contrato para a conta ficar certa.</div>`:''}
    ${ed?`<button class="btn primary" data-act="contrato-salvar">${ico('check',16)} Salvar contrato</button>`:''}</div>`;
}
function vStudy(C){
  const cid=C.cid,st=C.st;const ct=st.ct;
  if(st.noBase)return pageH('Estudo do projeto',esc(C.def.name))+`<div class="card"><div class="empty"><b>Sem dado</b>Sem linha de base: as primeiras semanas do acompanhamento não têm consumo válido. A economia não é estimada.</div></div>`+contratoForm(C,ct,st.mens);
  /* granularidade escolhida pela pessoa: mês, semana ou dia (gráfico e tabela juntos) */
  const gran=['mes','semana','dia'].includes(S.u.estGran)?S.u.estGran:'mes';const DS=['dom','seg','ter','qua','qui','sex','sáb'];
  const ser=gran==='mes'?st.months.map(m=>({label:m.label.replace(/ \d{4}$/,'')+'/'+m.ym.slice(2,4),full:m.label,sub:m.cal<28?m.cal+' dias':'',inv:m.inv,econ:m.econ,lit:m.lit,kmpl:m.kmpl,score:m.score,km:m.km}))
    :gran==='semana'?st.weeks.map(w=>({label:fmt.dm(dIso(w.from)),full:fmt.dm(dIso(w.from))+' a '+fmt.dm(dIso(w.to)),sub:w.cal<7?w.cal+' dias':'',inv:w.inv,econ:w.econ,lit:w.lit,kmpl:w.kmpl,score:w.score,km:w.km}))
    :st.days.map(d=>({label:fmt.dm(d.date),full:fmt.dmy(d.date)+' · '+DS[wdOf(d.t)],sub:d.n?'':'sem rodagem',inv:d.inv,econ:d.econ,lit:d.lit,kmpl:d.kmpl,score:d.score,km:d.km}));
  const nomeG={mes:'mês',semana:'semana',dia:'dia'}[gran];
  const bars=monthBars(ser,{sub:gran==='mes',h:gran==='dia'?300:330,base:st.base,aria:'Investimento, economia e consumo por '+nomeG});
  const ajuda=infoDot('Como calculamos',`<p>A <b>linha de base</b> é o consumo do início do projeto: ${esc(st.rotIni)} (${fmt.dm(dIso(st.pIni[0]))} a ${fmt.dm(dIso(st.pIni[1]))}), <b>${fmt.n(st.base,2)} km/l</b>.</p>
    <p><b>Litros economizados</b> em cada ${nomeG} = km rodados ÷ ${fmt.n(st.base,2)} (o que a frota gastaria no consumo da linha de base) − litros realmente gastos. Consumo acima da linha de base economiza litros; abaixo, gasta a mais.</p>
    <p><b>Economia</b> = litros economizados × diesel a ${fmt.brl(C.price,2)}/L.</p>
    <p><b>Investimento</b> = mensalidade (${fmt.brl(ct.unit,2)} × ${ct.units} unidades = ${fmt.brl(st.mens,2)}) dividida pelos dias do mês e somada nos dias de cada ${nomeG}.</p>
    <p>A linha laranja é o consumo (km/l) de cada ${nomeG}, no eixo da direita; a tracejada é a linha de base. Quanto mais a linha laranja fica acima da tracejada, maior a barra verde.</p>`,{left:true});
  const rows=ser.map(m=>{const sal=m.econ-m.inv;return`<tr><td><b>${m.full}</b>${m.sub?` <span class="xs muted">${m.sub}</span>`:''}</td><td class="r num">${fmt.brl(m.inv)}</td><td class="r num" style="color:${m.lit>=0?'var(--ok-t)':'var(--bad-t)'}">${m.km?fmt.n(m.lit,0)+' L':'—'}</td><td class="r num b" style="color:${m.econ>=0?'var(--ok-t)':'var(--bad-t)'}">${fmt.brl(m.econ)}</td><td class="r num b" style="color:${sal>=0?'var(--ok-t)':'var(--bad-t)'}">${fmt.brl(sal)}</td><td class="r num">${m.inv?(m.econ<0?'negativa':fmt.n(m.econ/m.inv*100,0)+'%'):'—'}</td><td class="r num">${m.kmpl!=null?fmt.kmpl(m.kmpl)+' km/l':'—'}</td><td class="r num">${fmt.n(m.score,0)}</td><td class="r num">${fmt.n(m.km,0)}</td></tr>`}).join('');
  const segG=segHtml([['mes','Mês'],['semana','Semana'],['dia','Dia']],gran,'est-gran');
  const chg=pc(st.second.kmpl,st.base);const per=p=>fmt.dm(dIso(p[0]))+' a '+fmt.dm(dIso(p[1]));
  const verdict=st.pays?`Sobram <span class="num" style="color:var(--amarelo)">${fmt.brl(st.econ-st.inv)}</span> depois de pagar o investimento`:`Faltam <span class="num" style="color:var(--amarelo)">${fmt.brl(st.miss)}</span> para a economia cobrir o investimento`;
  return`${pageH('Estudo do projeto',`${esc(C.def.name)} · ${fmt.dm(dIso(st.ps))} a ${fmt.dm(TODAY)} · ${st.len} dias · diesel a ${fmt.brl(C.price,2)}/L`,`<button class="btn ghost" data-act="print">${ico('download',16)} Exportar PDF</button>`)}
  <div class="grid g-main">
    <div class="card dark rise stack" style="padding:28px;position:relative;overflow:hidden"><div class="swoosh" style="border-color:rgba(254,219,0,.12)"></div><div class="row"><span class="badge ${st.pays?'ok':'bad'}">${st.pays?'O PROJETO SE PAGA':'AINDA NÃO SE PAGA'}</span></div><div><div style="color:#e1e1dd;font-size:16px">O projeto se paga?</div><div style="font-size:34px;font-weight:900;letter-spacing:-.01em;line-height:1.15">${verdict}</div></div>
      <div class="row" style="gap:32px"><div><div class="muted">Investido no período</div><b class="num" style="font-size:26px">${count(st.inv,0,'R$ ')}</b></div><div><div class="muted">${st.econ>=0?'Diesel economizado':'Diesel gasto a mais'}</div><b class="num" style="font-size:26px;color:${st.econ>=0?'#7be0ac':'#ff9d92'}">${fmt.brl(st.econ)}</b></div><div><div class="muted">Ritmo atual por mês</div><b class="num" style="font-size:26px;color:${st.pace>=0?'#7be0ac':'#ff9d92'}">${fmt.brl(st.pace)}</b></div><div><div class="muted">Retorno sobre o investimento</div><b class="num" style="font-size:26px">${st.roi!=null?fmt.n(st.roi,1)+'×':'—'}</b></div></div>
      <div class="muted sm">Linha de base: ${st.rotIni} (${per(st.pIni)}), ${fmt.n(st.base,2)} km/l. Agora: ${st.rotAgora} (${per(st.pAgora)}), ${fmt.n(st.second.kmpl,2)} km/l ${dl(chg,true)}. Mesma regra do painel clássico.</div></div>
    ${contratoForm(C,ct,st.mens)}
  </div>
  <div class="sec">${confBlock(C)}</div>
  <div class="sec">
    <div class="card rise">${cardH('Investimento, economia e consumo por '+nomeG,'Mensalidade da Gobrax, diesel economizado (litros e reais) e o consumo de cada '+nomeG+' contra a linha de base. Passe o mouse para ver o detalhe.',segG+ajuda)}${bars}<div class="leg" style="margin-top:10px"><span><i style="background:var(--n600)"></i>Investimento</span><span><i style="background:var(--ok)"></i>Economia</span><span><i style="background:var(--bad)"></i>Economia negativa</span><span><i style="background:var(--laranja)"></i>Consumo (km/l)</span><span><i style="background:none;border-top:2px dashed var(--n500);height:0;margin-top:5px"></i>Linha de base ${fmt.n(st.base,2)} km/l</span></div></div>
  </div>
  <div class="card sec rise">${cardH({mes:'Mês a mês',semana:'Semana a semana',dia:'Dia a dia'}[gran],'Investimento, litros e reais economizados, saldo, cobertura, consumo, nota e km',segG)}<div class="scroll-x"${gran==='mes'?'':' style="max-height:520px;overflow:auto"'}><table class="tbl"><thead><tr><th>${{mes:'Mês',semana:'Semana',dia:'Dia'}[gran]}</th><th class="r">Investimento</th><th class="r">Litros economizados</th><th class="r">Economia</th><th class="r">Saldo</th><th class="r">Cobertura</th><th class="r">Consumo</th><th class="r">Nota</th><th class="r">Km</th></tr></thead><tbody>${rows}</tbody></table></div>
    <div class="sm muted" style="margin-top:14px"><b>Como calculamos:</b> litros economizados = litros que o consumo da linha de base gastaria nos mesmos km, menos os litros realmente gastos; a economia é esses litros vezes o preço do diesel. O investimento é a mensalidade (valor por unidade × unidades) proporcional aos dias de cada mês.</div></div>`;
}

/* ---------- Confiança da linha de base (item 6): só sinaliza, não ajusta o número ---------- */
const CONF_CLS={alta:'ok','média':'warn',baixa:'bad','sem dado':'n'};
function confTable(C){
  const c=C.conf,f=c.f,cur=C.cur;if(!c.b.base)return'';
  const r=(l,a,b,d)=>`<tr><td>${l}</td><td class="r num">${a}</td><td class="r num">${b}</td><td class="r">${d||''}</td></tr>`;
  return`<table class="tbl"><thead><tr><th></th><th class="r">Linha de base</th><th class="r">Período atual</th><th class="r">Diferença</th></tr></thead><tbody>
  ${r('Período',fmt.dm(dIso(c.b.pIni[0]))+' a '+fmt.dm(dIso(c.b.pIni[1])),C.P.label)}
  ${r('Dias com dado',fmt.n(f.dataDays,0),fmt.n(cur.dataDays,0))}
  ${r('Km',fmt.n(f.km,0),fmt.n(cur.km,0))}
  ${r('Consumo',fmt.kmpl(f.kmpl)+' km/l',fmt.kmpl(cur.kmpl)+' km/l',c.dk!=null?`<span class="dl ${Math.abs(c.dk)<ST.rules.difPequena?'n':c.dk>0?'good':'bad'}">${fmt.sp(c.dk)}</span>`:'')}
  ${r('Km por dia por veículo',fmt.n(f.kmDay,0),fmt.n(cur.kmDay,0),c.kd!=null?`<span class="dl ${Math.abs(c.kd)>ST.rules.kmDiaMuda?'bad':'n'}">${fmt.sp(c.kd)}</span>`:'')}
  ${r('Veículos com dado',fmt.n(f.active,0),fmt.n(cur.active,0))}
  ${r('Km de veículos que também rodaram na base','',c.mix!=null?fmt.n(c.mix,0)+'%':'sem dado')}</tbody></table>`;
}
function confBadge(C){
  const c=C.conf;const av=c.avisos.length?`<ul>${c.avisos.map(a=>`<li>${esc(a)}</li>`).join('')}</ul>`:'<p>Nenhum aviso: base e período atual são comparáveis.</p>';
  return`<span class="badge ${CONF_CLS[c.nivel]}" style="margin-left:4px">confiança ${c.nivel}</span> ${infoDot('Confiança do número',`<p>${esc(C.ec.b.rotulo)}.</p>${confTable(C)}${av}<p class="xs">Alta sem avisos, média com 1 ou 2, baixa com 3 ou mais. Nenhum ajuste é feito no número.</p>`)}`;
}
function confBlock(C){
  const c=C.conf;
  return`<div class="card rise">${cardH('Confiança da linha de base',`Diesel economizado ${esc(C.ec.b.rotulo)}`,`<span class="badge ${CONF_CLS[c.nivel]}">CONFIANÇA ${c.nivel.toUpperCase()}</span>`)}${confTable(C)}
  <div class="stack" style="gap:6px;margin-top:12px">${c.avisos.map(a=>`<div class="sm" style="background:var(--warn-s,#fff4d6);border-left:4px solid var(--warn);border-radius:8px;padding:8px 12px">⚠ ${esc(a)}</div>`).join('')||'<div class="sm muted">Nenhum aviso: base e período atual são comparáveis.</div>'}</div>
  <div class="xs muted" style="margin-top:8px">Os avisos só sinalizam; nenhum ajuste é feito no número. Alta sem avisos, média com 1 ou 2, baixa com 3 ou mais.</div></div>`;
}

/* ---------- Caso individual × tendência da frota (item 3) ---------- */
function concBlock(C){
  const k=C.conc;if(!k.top)return'';
  const linha=`<div class="row sm" style="margin-top:10px;gap:22px;flex-wrap:wrap"><span>Sem o maior caso (<b>${esc(k.top.id)}</b>): <b>${fmt.brl(k.semTop)}</b></span><span><b>${fmt.n(k.share,0)}%</b> da perda vem de ${esc(k.top.id)}</span></div>`;
  if(!k.alerta)return linha;
  return linha+`<div class="sm" style="margin-top:10px;background:var(--warn-s,#fff4d6);border-left:4px solid var(--warn);border-radius:8px;padding:8px 12px">⚠ Um único veículo responde por ${fmt.n(k.share,0)}% da perda (limite: ${ST.rules.concPerda}%). Sinais em volta dele, sem apontar causa:</div>${sinaisVeiculo(C,k.top.id)}`;
}
/* consumo de outros condutores no mesmo veículo (registros por motorista do servidor); sem esse dado, "sem dado" */
function outrosCondutores(C,id,from,to){
  if(!REAL.on)return null;const raw=(REAL.raw||[]).find(c=>c.id===C.def.srvId);if(!raw)return null;const main=nkey((C.fleet.byId[id]||{}).driver||'');
  let km=0,l=0;for(const r of raw.motoristas||[]){if(r.placa!==id||!r.motorista||nkey(r.motorista)===main)continue;const t=dIdx(String(r.dia).slice(0,10));if(t<from||t>to)continue;const k=+r.km||0,li=+r.litros||0;if(k>0&&li>0){km+=k;l+=li}}
  return l?km/l:null;
}
function sinaisVeiculo(C,id){
  const c=C.cur.byVeh[id],p=C.prev.byVeh[id];const sd='<span class="muted">sem dado</span>';
  const o1=outrosCondutores(C,id,C.P.cf,C.P.ct),o2=outrosCondutores(C,id,C.P.from,C.P.to);
  const row=(l,a,b,good,d,suf,pp)=>{const ok=v=>v!=null&&isFinite(v);const ch=ok(a)&&ok(b)?(pp?a-b:b?(a-b)/b*100:null):null;
    return`<tr><td>${l}</td><td class="r num">${ok(b)?fmt.n(b,d)+suf:sd}</td><td class="r num">${ok(a)?fmt.n(a,d)+suf:sd}</td><td class="r">${ch==null?'':pp?`<span class="dl ${(good?ch>=0:ch<=0)?'good':'bad'}">${ch>0?'+':''}${fmt.n(ch,1)} pp</span>`:dl(ch,good)}</td></tr>`};
  return`<div class="scroll-x"><table class="tbl" style="margin-top:8px"><thead><tr><th>Sinal (${esc(id)})</th><th class="r">Antes</th><th class="r">Agora</th><th class="r">Mudança</th></tr></thead><tbody>
  ${row('Consumo do veículo com o condutor principal (km/l)',c&&c.kmpl,p&&p.kmpl,true,2,'')}
  ${row('Mesmo veículo com outros condutores (km/l)',o2,o1,true,2,'')}
  ${row('Km por dia trabalhado',c&&c.kmDay,p&&p.kmDay,true,0,'')}
  ${row('Motor ocioso',c&&c.idle,p&&p.idle,false,1,'%',true)}
  ${row('Faixa verde',c&&c.green,p&&p.green,true,1,'%',true)}
  ${row('Freadas a cada 100 km',c&&c.br100,p&&p.br100,false,1,'')}
  ${row('Nota Gobrax',c&&c.score,p&&p.score,true,0,'')}</tbody></table></div><div class="xs muted" style="margin-top:6px">Antes = ${esc(C.P.cmpLabel)}. Só mostra os sinais; não aponta causa.</div>`;
}

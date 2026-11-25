/* =====================================================================
   GRÁFICOS AVANÇADOS · radar, dispersão, divergente, medidor, calendário,
   barras diárias, comparativos e catálogo de indicadores
   ===================================================================== */
let _uid=0;const uid=p=>p+(++_uid);

/* ---------- Catálogo de indicadores (espelha a API: totalMileage, totalConsumption, totalBreaking…) ---------- */
const IND=[
  {k:'br100',l:'Freadas a cada 100 km',s:'Freadas totais a cada 100 km rodados',g:'Indicadores de condução',get:a=>a.br100,d:1,u:' /100 km',up:false},
  {k:'brh100',l:'Freadas em alta velocidade',s:'Freadas em alta velocidade a cada 100 km',g:'Indicadores de condução',get:a=>a.brh100,d:2,u:' /100 km',up:false},
  {k:'speed',l:'Velocidade média',s:'Média ponderada pelos km rodados',g:'Indicadores de condução',get:a=>a.speed,d:1,u:' km/h',up:null},
  {k:'green',l:'Faixa verde',s:'Tempo dirigindo na faixa econômica do motor',g:'Indicadores de condução',get:a=>a.green,d:1,u:'%',up:true},
  {k:'idle',l:'Motor ligado parado',s:'Motor ligado com o veículo parado',g:'Indicadores de condução',get:a=>a.idle,d:1,u:'%',up:false},
  {k:'emb',l:'Aproveitamento de embalo',s:'Tempo rodando no embalo, sem acelerar',g:'Indicadores de condução',get:a=>a.emb,d:1,u:'%',up:true},
  {k:'fm',l:'Freio motor',s:'Uso do freio motor nas reduções',g:'Indicadores de condução',get:a=>a.fm,d:1,u:'%',up:true},
  {k:'pa',l:'Piloto automático',s:'Tempo rodando com piloto automático',g:'Indicadores de condução',get:a=>a.pa,d:1,u:'%',up:true},
  {k:'exv',l:'Excesso de velocidade',s:'Tempo acima do limite de velocidade',g:'Indicadores de condução',get:a=>a.exv,d:1,u:'%',up:false},
  {k:'score',l:'Nota Gobrax',s:'Nota ponderada pelos km rodados',g:'Indicadores de condução',get:a=>a.score,d:0,u:' pts',up:true},
  {k:'kmpl',l:'Média de consumo',s:'Quilômetros por litro',g:'Consumo e eficiência',get:a=>a.kmpl,d:2,u:' km/l',up:true},
  {k:'lph100',l:'Litros a cada 100 km',s:'Consumo relativo à distância',g:'Consumo e eficiência',get:a=>a.lph100,d:1,u:' L',up:false},
  {k:'l',l:'Diesel consumido',s:'Total de litros no período',g:'Consumo e eficiência',get:a=>a.l,d:0,u:' L',up:null},
  {k:'km',l:'Km rodados',s:'Distância total no período',g:'Produtividade',get:a=>a.km,d:0,u:' km',up:true},
  {k:'n',l:'Dias trabalhados',s:'Dias com rodagem',g:'Produtividade',get:a=>a.n,d:0,u:' dias',up:true,veh:true},
  {k:'kmDay',l:'Km por dia trabalhado',s:'Média nos dias com km rodado',g:'Produtividade',get:a=>a.kmDay,d:0,u:' km',up:true}
];
const INDK=Object.fromEntries(IND.map(i=>[i.k,i]));
const fInd=(i,v)=>v==null||!isFinite(v)?'—':fmt.n(v,i.d)+i.u;
/* resultado de um indicador: antes x agora */
function indRes(i,cur,prev){
  const a=cur!=null?i.get(cur):null,b=prev?i.get(prev):null;
  if(a==null||b==null||!b)return{ch:null,k:'base',label:'SEM BASE',cls:'n'};
  const ch=(a-b)/Math.abs(b)*100;
  if(Math.abs(ch)<2)return{ch,k:'same',label:'ESTÁVEL',cls:'n'};
  if(i.up===null)return{ch,k:ch>0?'up':'down',label:ch>0?'SUBIU':'DESCEU',cls:'n'};
  const good=(ch>0)===i.up;return{ch,k:good?'better':'worse',label:good?'MELHOROU':'PIOROU',cls:good?'ok':'bad'};
}
const indGood=(i,ch)=>ch==null||i.up===null?null:(ch>0)===i.up;

/* agregado de um veículo (ou da frota) em um intervalo */
function aggFor(C,from,to,id){
  if(!id)return metrics(C.cid,ST,from,to);
  return aggregate(validRecs(C.cid,ST).filter(r=>r.veh===id&&r.t>=from&&r.t<=to));
}
/* séries por granularidade: dia · semana · quinzena · mês */
function buckets(C,gran,id){
  const start=dIdx(C.def.start),out=[];const push=(f,e,label,full)=>{const f0=f,e0=full!=null?full:e;f=Math.max(f,start);if(e<f)return;const a=aggFor(C,f,e,id);out.push({from:f,to:e,label,a:a.n?a:null,dias:e-f+1,parcial:gran!=='dia'&&(f>f0||e<e0)})};
  if(gran==='dia'){for(let t=C.P.from;t<=C.P.to;t++)push(t,t,fmt.dm(dIso(t)))}
  else if(gran==='semana'){const mon=T0-((wdOf(T0)+6)%7);for(let i=7;i>=0;i--){const f=mon-7*i,e=Math.min(T0,f+6);if(e<start)continue;push(f,e,fmt.dm(dIso(Math.max(f,start))),f+6)}}
  else if(gran==='quinzena'){const d=new Date(T0*DAY);let y=d.getUTCFullYear(),m=d.getUTCMonth();const list=[];
    for(let k=0;k<7;k++){const first=dIdx(`${y}-${String(m+1).padStart(2,'0')}-01`),mid=first+14,last=Math.round(Date.UTC(y,m+1,1)/DAY)-1;list.unshift([mid+1,Math.min(T0,last),`2ª quinz. ${MESES[m]}`,last],[first,Math.min(T0,mid),`1ª quinz. ${MESES[m]}`,mid]);m--;if(m<0){m=11;y--}}
    list.filter(x=>x[0]<=T0&&x[1]>=start).slice(-6).forEach(x=>push(x[0],x[1],x[2],x[3]))}
  else{const d=new Date(T0*DAY);let y=d.getUTCFullYear(),m=d.getUTCMonth();const list=[];
    for(let k=0;k<6;k++){const first=dIdx(`${y}-${String(m+1).padStart(2,'0')}-01`),last=Math.round(Date.UTC(y,m+1,1)/DAY)-1;list.unshift([first,Math.min(T0,last),MESES[m]+'/'+String(y).slice(2),last]);m--;if(m<0){m=11;y--}}
    list.filter(x=>x[1]>=start).forEach(x=>push(x[0],x[1],x[2],x[3]))}
  return out;
}

/* ---------- Medidor de Nota Gobrax (ticks) ---------- */
function gaugeNota(v,sub='NOTA GOBRAX',ink=false){
  const N=41;let t='';const has=v!=null&&isFinite(v);
  for(let i=0;i<N;i++){const p=i/(N-1)*100,ang=Math.PI*(1-i/(N-1));const c=Math.cos(ang),s=Math.sin(ang);
    const lit=has&&p<=v+0.01;const col=ink?(lit?'#2b2b2b':'rgba(255,255,253,.55)'):!lit?'var(--n200)':p<40?'var(--bad)':p<70?'var(--warn)':'var(--ok)';
    t+=`<line class="tk" style="--i:${i}" x1="${(100+c*66).toFixed(1)}" y1="${(98-s*66).toFixed(1)}" x2="${(100+c*88).toFixed(1)}" y2="${(98-s*88).toFixed(1)}" stroke="${col}" stroke-width="3" stroke-linecap="round"/>`}
  return`<div class="gnota ${ink?'ink':''}"><svg viewBox="0 0 200 110" role="img" aria-label="${has?'Nota Gobrax '+Math.round(v)+' de 100':'Sem nota'}">${t}</svg><div class="gv num">${has?count(Math.round(v),0):'—'}</div><div class="tag">${sub}</div></div>`;
}

/* ---------- Radar "Perfil de condução" ---------- */
function radar(axes){
  const have=axes.filter(a=>a.cur!=null&&a.fleet);if(have.length<3)return`<div class="empty" style="padding:40px 10px"><b>Poucos indicadores</b>Faltam dados para montar o gráfico de teia.</div>`;
  const Sz=340,c=Sz/2,R=98,n=axes.length;
  const rr=(v,f,up)=>{if(v==null||!f)return .04;const q=up?v/f:(v>0?f/v:2);return clamp(.5+(q-1)*.8,.06,1)};
  const pt=(i,r)=>{const a=-Math.PI/2+i*2*Math.PI/n;return[c+Math.cos(a)*R*r,c+Math.sin(a)*R*r]};
  const poly=(f)=>axes.map((a,i)=>pt(i,f(a)).map(x=>x.toFixed(1)).join(',')).join(' ');
  let g='';[.25,.5,.75,1].forEach(r=>g+=`<polygon points="${poly(()=>r)}" fill="none" stroke="var(--line)" stroke-width="1"/>`);
  axes.forEach((a,i)=>{const [x,y]=pt(i,1);g+=`<line x1="${c}" y1="${c}" x2="${x.toFixed(1)}" y2="${y.toFixed(1)}" stroke="var(--line)"/>`;
    const [lx,ly]=pt(i,1.2);const cs=Math.cos(-Math.PI/2+i*2*Math.PI/n);g+=`<text x="${lx.toFixed(1)}" y="${(ly+4).toFixed(1)}" text-anchor="${Math.abs(cs)<.2?'middle':cs>0?'start':'end'}">${esc(a.l)}</text>`});
  const hasPrev=axes.some(a=>a.prev!=null);
  const body=`<polygon class="rad" points="${poly(a=>.5)}" fill="none" stroke="var(--n400)" stroke-width="1.8" stroke-dasharray="5 4"/>`
    +(hasPrev?`<polygon class="rad" style="--d:.1s" points="${poly(a=>rr(a.prev,a.fleet,a.up))}" fill="none" stroke="var(--info)" stroke-width="2" stroke-dasharray="3 3"/>`:'')
    +`<polygon class="rad" style="--d:.25s" points="${poly(a=>rr(a.cur,a.fleet,a.up))}" fill="var(--amarelo)" fill-opacity=".28" stroke="var(--laranja)" stroke-width="2.6" stroke-linejoin="round"/>`
    +axes.map((a,i)=>{const [x,y]=pt(i,rr(a.cur,a.fleet,a.up));return`<circle class="pt" style="--i:${i}" cx="${x.toFixed(1)}" cy="${y.toFixed(1)}" r="5" fill="var(--laranja)" stroke="var(--card)" stroke-width="2" data-tt="<b>${esc(a.l)}</b><br>Agora: ${esc(a.fc(a.cur))}${a.prev!=null?'<br>Antes: '+esc(a.fc(a.prev)):''}<br>Frota: ${esc(a.fc(a.fleet))}"/>`}).join('');
  return`<svg class="chart radar" viewBox="${-30} 0 ${Sz+60} ${Sz}" role="img" aria-label="Perfil de condução">${g}${body}</svg><div class="leg" style="justify-content:center;margin-top:6px"><span><i style="background:var(--amarelo)"></i>Agora</span>${hasPrev?'<span><i style="background:var(--info)"></i>Antes</span>':''}<span><i style="background:var(--n400)"></i>Média da frota</span></div>`;
}
function radarAxes(cur,prev,fleet){
  const ks=['kmpl','br100','green','idle','emb','fm','pa','exv','score','kmDay'],lab={kmpl:'Consumo',br100:'Freadas',green:'Faixa verde',idle:'Parado ligado',emb:'Embalo',fm:'Freio motor',pa:'Piloto aut.',exv:'Excesso vel.',score:'Nota',kmDay:'Km/dia'};
  return ks.map(k=>{const i=INDK[k];return{l:lab[k],cur:cur?i.get(cur):null,prev:prev?i.get(prev):null,fleet:fleet?i.get(fleet):null,up:i.up,fc:v=>fInd(i,v)}});
}

/* ---------- Dispersão com quadrantes ---------- */
function scatter(pts,o){
  const W=o.w||720,H=o.h||380,pl=52,pr=20,pt=20,pb=44;
  if(pts.length<2)return`<div class="empty"><b>Poucos condutores com base de comparação</b></div>`;
  let xs=pts.map(p=>p.x),ys=pts.map(p=>p.y);let xlo=Math.min(...xs),xhi=Math.max(...xs),yl=Math.max(Math.abs(Math.min(...ys)),Math.abs(Math.max(...ys)),5);
  const px=(xhi-xlo)*.1||.1;xlo-=px;xhi+=px;yl=Math.ceil(yl*1.1/5)*5;
  const X=v=>pl+(v-xlo)/(xhi-xlo)*(W-pl-pr),Y=v=>pt+(yl-v)/(2*yl)*(H-pt-pb);
  const mx=X(o.xMid),my=Y(o.yMid);
  let g=`<rect x="${mx}" y="${pt}" width="${Math.max(0,W-pr-mx)}" height="${my-pt}" fill="var(--ok-s)" opacity=".55"/><rect x="${pl}" y="${my}" width="${Math.max(0,mx-pl)}" height="${H-pb-my}" fill="var(--bad-s)" opacity=".55"/>`;
  g+=`<text class="ql" x="${W-pr-8}" y="${pt+16}" text-anchor="end">${o.q[0]}</text><text class="ql" x="${pl+8}" y="${pt+16}">${o.q[1]}</text><text class="ql" x="${W-pr-8}" y="${H-pb-10}" text-anchor="end">${o.q[2]}</text><text class="ql" x="${pl+8}" y="${H-pb-10}">${o.q[3]}</text>`;
  for(let k=-yl;k<=yl;k+=yl/2){g+=`<line class="grid-l" x1="${pl}" x2="${W-pr}" y1="${Y(k)}" y2="${Y(k)}"/><text x="${pl-8}" y="${Y(k)+4}" text-anchor="end">${k>0?'+':k<0?'−':''}${fmt.n(Math.abs(k),0)}%</text>`}
  for(let k=0;k<=4;k++){const v=xlo+(xhi-xlo)*k/4;g+=`<text x="${X(v)}" y="${H-pb+18}" text-anchor="middle">${fmt.n(v,2)}</text>`}
  g+=`<line x1="${mx}" x2="${mx}" y1="${pt}" y2="${H-pb}" stroke="var(--n400)" stroke-dasharray="5 5"/><line x1="${pl}" x2="${W-pr}" y1="${my}" y2="${my}" stroke="var(--n400)" stroke-dasharray="5 5"/>`;
  g+=`<text x="${(pl+W-pr)/2}" y="${H-6}" text-anchor="middle">${esc(o.xl)}</text><text transform="rotate(-90)" x="${-(pt+H-pb)/2}" y="12" text-anchor="middle">${esc(o.yl)}</text>`;
  const dots=pts.map((p,i)=>{const col=p.x>=o.xMid&&p.y>=o.yMid?'var(--ok)':p.x<o.xMid&&p.y<o.yMid?'var(--bad)':'var(--laranja)';
    return`<g class="sp" data-act="drv-open" data-id="${esc(p.id)}" style="--i:${i}" data-tt="${esc(p.tip)}"><circle cx="${X(p.x).toFixed(1)}" cy="${Y(clamp(p.y,-yl,yl)).toFixed(1)}" r="7" fill="${col}" fill-opacity=".85" stroke="var(--card)" stroke-width="2"/></g>`}).join('');
  return`<svg class="chart scat${o._quiet?' noanim':''}" data-cr="${chartReg((q)=>scatter(pts,q),o)}" viewBox="0 0 ${W} ${H}" role="img" aria-label="${esc(o.aria||'Dispersão')}">${g}${dots}</svg>`;
}

/* ---------- Lista divergente (ganhos e perdas) ---------- */
function diverging(items,o={}){
  if(!items.length)return`<div class="empty"><b>Sem comparação disponível</b></div>`;
  const max=Math.max(...items.map(i=>Math.abs(i.v)),1e-9);
  return`<div class="dv">${items.map((it,k)=>{const w=Math.abs(it.v)/max*50;const pos=it.v>=0;
    return`<div class="dvr" ${it.id?`data-act="drv-open" data-id="${esc(it.id)}"`:''}><div class="dvl"><b>${esc(it.l)}</b><span class="xs muted">${esc(it.s||'')}</span></div><div class="dvt"><i class="dvz"></i><div class="bfill dvb ${pos?'':'neg'}" style="--i:${k};${pos?'left:50%':'right:50%'};width:${w}%;background:${pos?'var(--ok)':'var(--bad)'}"></div></div><div class="dvv num b" style="color:${pos?'var(--ok-t)':'var(--bad-t)'}">${o.f?o.f(it.v):fmt.brl(it.v)}</div></div>`}).join('')}</div>`;
}

/* ---------- Calendário de desempenho (pontos) ---------- */
/* get(t) → {val, extra} ou null; m = {name, d, u, f:[bom, atenção], leg:[3 rótulos]} */
function dotCal(from,to,get,m){
  m=m||{name:'Nota',d:0,u:'',f:[70,40],leg:['Bom (70+)','Atenção (40 a 69)','Crítico (abaixo de 40)']};
  const mon=from-((wdOf(from)+6)%7);const wk=Math.floor((to-mon)/7)+1;const L=['Seg','Ter','Qua','Qui','Sex','Sáb','Dom'];
  let h=L.map((l,i)=>`<span class="dl-l" style="grid-row:${i+2};grid-column:1">${l}</span>`).join('');let mesAnt='';
  for(let t=from;t<=to;t++){const col=Math.floor((t-mon)/7)+2,row=(wdOf(t)+6)%7+2;const r=get(t);const v=r?r.val:null;
    const mes=dIso(t).slice(0,7);if(mes!==mesAnt){mesAnt=mes;const fimMes=Math.min(to,Math.round(Date.UTC(+mes.slice(0,4),+mes.slice(5),1)/DAY)-1);
      /* mês com menos de 2 semanas no início do calendário: sem nome, senão encosta no nome do mês seguinte */
      if(!(t===from&&fimMes-t<13))h+=`<span class="dl-l" style="grid-row:1;grid-column:${col}/span 4;font-weight:600;color:var(--text)">${MESES[+mes.slice(5)-1]}</span>`}
    const cl=v==null?'e':v>=m.f[0]?'g':v>=m.f[1]?'y':'r';const tip=r?`<b>${fmt.dmy(dIso(t))}</b><br>${esc(m.name)}: ${v==null?'sem dado':fmt.n(v,m.d)+m.u}${r.extra?'<br>'+r.extra:''}`:`${fmt.dmy(dIso(t))} · sem rodagem`;
    h+=`<i class="d ${cl}" style="grid-row:${row};grid-column:${col};--i:${Math.min(t-from,60)}" data-tt="${esc(tip)}"></i>`}
  const z=wk>9?15:24;/* muitas semanas: pontos menores para o calendário caber no cartão */
  return`<div class="dotcal${wk>9?' sm':''}" style="grid-template-columns:30px repeat(${wk},${z}px);grid-template-rows:18px repeat(7,${z}px)">${h}</div><div class="leg" style="margin-top:14px"><span><i style="background:var(--ok);border-radius:99px"></i>${m.leg[0]}</span><span><i style="background:var(--warn);border-radius:99px"></i>${m.leg[1]}</span><span><i style="background:var(--bad);border-radius:99px"></i>${m.leg[2]}</span><span><i style="background:var(--soft);border-radius:99px"></i>Sem rodagem ou sem dado</span></div>`;
}
/* calendário dos últimos 3 meses (frota ou um veículo). Cor do dia como no painel clássico: nota do dia quando a API
   manda; senão início da faixa verde; senão faixa verde; sem nenhum dos três, consumo contra a média dos 3 meses. */
function calBlock(C,id){
  const to=C.P.to,from=Math.max(dIdx(C.def.start),to-90);
  const ds=dailySeries(C.cid,ST,from,to,id?x=>x===id:null);const has=k=>ds.some(d=>d.a&&d.a[k]!=null);
  let m;
  if(has('scd'))m={k:'scd',name:'Nota do dia',d:0,u:'',f:[70,40],leg:['Bom (70+)','Atenção (40 a 69)','Crítico (abaixo de 40)']};
  else if(has('ifv'))m={k:'ifv',name:'Início da faixa verde',d:0,u:'%',f:[85,55],leg:['Ideal (85%+)','Aceitável (55% a 84%)','Abaixo de 55%']};
  else if(has('green'))m={k:'green',name:'Faixa verde',d:0,u:'%',f:[85,55],leg:['85%+','55% a 84%','Abaixo de 55%']};
  else{const a=aggregate(validRecs(C.cid,ST).filter(r=>r.t>=from&&r.t<=to&&(!id||r.veh===id))).kmpl||0;
    m={k:'kmpl',name:'Consumo',d:2,u:' km/l',f:[a*1.02,a*.98],leg:['Acima da média ('+fmt.n(a,2)+' km/l)','Na média','Abaixo da média']}}
  const html=dotCal(from,to,t=>{const d=ds[t-from];return d&&d.a?{val:d.a[m.k],extra:fmt.n(d.a.km,0)+' km · '+fmt.kmpl(d.a.kmpl)+' km/l'}:null},m);
  return{html,sub:`Últimos 3 meses (${fmt.dm(dIso(from))} a ${fmt.dm(dIso(to))}). Cada dia colorido pel${m.k==='kmpl'||m.k==='ifv'?'o':'a'} ${m.name.toLowerCase()}. Passe o mouse para ver o valor.`};
}

/* ---------- Barras diárias (hachuradas, destaque no hover) ---------- */
function dailyBars(vals,labels,o={}){
  const W=o.w||560,H=o.h||240,pl=44,pr=10,pt=18,pb=28;const v=vals.filter(x=>x!=null);if(!v.length)return`<div class="empty"><b>Sem dados no período</b></div>`;
  const max0=Math.max(...v)||1;const nice=r=>{const p=Math.pow(10,Math.floor(Math.log10(r)));const f=r/p;return(f<1.5?1:f<3?2:f<7?5:10)*p};
  const stp=nice(max0/4),hi=Math.ceil(max0/stp)*stp;const n=vals.length,bw=(W-pl-pr)/n;const y=x=>pt+(hi-x)/hi*(H-pt-pb);const id=uid('ht');
  let g=`<defs><pattern id="${id}" width="6" height="6" patternUnits="userSpaceOnUse" patternTransform="rotate(45)"><rect width="6" height="6" fill="var(--soft)"/><line x1="0" y1="0" x2="0" y2="6" stroke="var(--n300)" stroke-width="2.4"/></pattern></defs>`;
  for(let k=0;k<=hi+1e-9;k+=stp)g+=`<line class="grid-l" x1="${pl}" x2="${W-pr}" y1="${y(k)}" y2="${y(k)}"/><text x="${pl-8}" y="${y(k)+4}" text-anchor="end">${fmt.n(k,0)}</text>`;
  const st=Math.max(1,Math.ceil(n/8));labels.forEach((l,i)=>{if(i%st===0)g+=`<text x="${pl+bw*i+bw/2}" y="${H-8}" text-anchor="middle">${esc(l)}</text>`});
  const bars=vals.map((x,i)=>x==null?'':`<rect class="dbar" style="--i:${i}" x="${(pl+bw*i+bw*.15).toFixed(1)}" y="${y(x).toFixed(1)}" width="${(bw*.7).toFixed(1)}" height="${(H-pb-y(x)).toFixed(1)}" rx="3" fill="url(#${id})" data-tt="<b>${esc(labels[i])}</b><br>${fmt.n(x,o.d||0)}${o.u||''}"/>`).join('');
  const avg=o.avg!=null&&isFinite(o.avg)?`<line x1="${pl}" x2="${W-pr}" y1="${y(o.avg)}" y2="${y(o.avg)}" stroke="var(--n500)" stroke-dasharray="5 4" stroke-width="1.5"/>`:'';
  return`<svg class="chart dbars${o._quiet?' noanim':''}" data-cr="${chartReg((q)=>dailyBars(vals,labels,q),o)}" viewBox="0 0 ${W} ${H}" role="img" aria-label="${esc(o.aria||'Barras por dia')}">${g}${bars}${avg}</svg>`;
}

/* ---------- Barras por mês: investimento ao lado da economia ---------- */
function monthBars(rows,o={}){ // rows:[{label,inv,econ,sub}]
  if(!rows.length)return`<div class="empty"><b>Sem meses no período</b></div>`;
  const kv=rows.map(r=>r.kmpl).filter(v=>v!=null&&isFinite(v)),comK=kv.length>0&&o.base>0;
  const W=o.w||720,H=o.h||300,pl=62,pr=comK?54:12,pt=26,pb=o.sub?46:30;
  const vals=rows.flatMap(r=>[r.inv,r.econ]).filter(v=>v!=null&&isFinite(v));let lo=Math.min(0,...vals),hi=Math.max(1,...vals);
  const nice=r=>{const p=Math.pow(10,Math.floor(Math.log10(r)));const f=r/p;return(f<1.5?1:f<3?2:f<7?5:10)*p};
  const stp=nice((hi-lo)/4||1);lo=Math.floor(lo/stp)*stp;hi=Math.ceil(hi*1.12/stp)*stp;
  const y=v=>pt+(hi-v)/(hi-lo)*(H-pt-pb);const n=rows.length,gw=(W-pl-pr)/n,bw=Math.min(54,gw*(n>16?.42:.3)),muitas=n>16,passo=Math.max(1,Math.ceil(n/12));
  const k=v=>Math.abs(v)>=1000?fmt.n(v/1000,1)+' mil':fmt.n(v,0);
  /* eixo da direita: consumo (km/l), com folga em volta da linha de base */
  let kLo=0,kHi=1;if(comK){kLo=Math.min(o.base,...kv);kHi=Math.max(o.base,...kv);const f=Math.max(.05,(kHi-kLo)*.25);kLo-=f;kHi+=f}
  const yk=v=>pt+(kHi-v)/(kHi-kLo)*(H-pt-pb);
  let g='';for(let v=lo;v<=hi+1e-9;v+=stp)g+=`<line class="grid-l" x1="${pl}" x2="${W-pr}" y1="${y(v)}" y2="${y(v)}"/><text x="${pl-8}" y="${y(v)+4}" text-anchor="end">${v?'R$ '+k(v):'0'}</text>`;
  if(lo<0)g+=`<line x1="${pl}" x2="${W-pr}" y1="${y(0)}" y2="${y(0)}" stroke="var(--n500)" stroke-width="1.5"/>`;
  const bar=(x,v,cls,fill,i,tip)=>{if(v==null||!isFinite(v))return'';const y0=y(0),yv=y(v),top=Math.min(y0,yv),h=Math.max(1,Math.abs(y0-yv));
    return`<rect class="dbar" style="--i:${i};transform-origin:${v>=0?'bottom':'top'}" x="${x.toFixed(1)}" y="${top.toFixed(1)}" width="${bw.toFixed(1)}" height="${h.toFixed(1)}" rx="${muitas?1.5:4}" fill="${fill}" data-tt="${esc(tip)}"/>${muitas?'':`<text x="${(x+bw/2).toFixed(1)}" y="${(v>=0?top-6:top+h+13).toFixed(1)}" text-anchor="middle" style="font-weight:700;font-size:11px;fill:var(--text)">${k(v)}</text>`}`};
  rows.forEach((r,i)=>{const cx=pl+gw*i+gw/2;const tip=`<b>${esc(r.full||r.label)}</b><br>Investimento: ${fmt.brl(r.inv)}<br>${r.lit!=null&&r.km?`Litros economizados: ${fmt.n(r.lit,0)} L<br>`:''}Economia: ${fmt.brl(r.econ)}<br>Saldo: ${fmt.brl(r.econ-r.inv)}${r.kmpl!=null?`<br>Consumo: ${fmt.n(r.kmpl,2)} km/l${o.base?` (base ${fmt.n(o.base,2)})`:''}`:''}`;
    const gap=muitas?1:3;g+=bar(cx-bw-gap,r.inv,'inv','var(--n600)',Math.min(i*2,60),tip)+bar(cx+gap,r.econ,'eco',r.econ>=0?'var(--ok)':'var(--bad)',Math.min(i*2+1,60),tip);
    if(i%passo===0)g+=`<text x="${cx}" y="${H-(o.sub?24:8)}" text-anchor="middle" style="font-weight:600;fill:var(--text)">${esc(r.label)}</text>`;
    const sub=[r.lit!=null&&r.km&&!muitas?fmt.n(r.lit,0)+' L':'',r.sub].filter(Boolean).join(' · ');
    if(o.sub&&sub&&!muitas)g+=`<text x="${cx}" y="${H-8}" text-anchor="middle" style="font-size:10.5px">${esc(sub)}</text>`});
  if(comK){
    for(let i=0;i<=3;i++){const v=kLo+(kHi-kLo)*i/3;g+=`<text x="${W-pr+6}" y="${yk(v)+4}" style="font-size:10.5px">${fmt.n(v,2)}</text>`}
    g+=`<text x="${W-4}" y="${pt-10}" text-anchor="end" style="font-size:10.5px">km/l</text>`;
    g+=`<line x1="${pl}" x2="${W-pr}" y1="${yk(o.base).toFixed(1)}" y2="${yk(o.base).toFixed(1)}" stroke="var(--n500)" stroke-width="1.5" stroke-dasharray="5 4"/>`;
    let d='',pts='';rows.forEach((r,i)=>{if(r.kmpl==null||!isFinite(r.kmpl))return;const x=(pl+gw*i+gw/2).toFixed(1),yy=yk(r.kmpl).toFixed(1);d+=(d?'L':'M')+x+' '+yy;
      if(!muitas)pts+=`<circle cx="${x}" cy="${yy}" r="4.5" fill="var(--card)" stroke="var(--laranja)" stroke-width="2.5" data-tt="${esc(`<b>${esc(r.full||r.label)}</b><br>Consumo: ${fmt.n(r.kmpl,2)} km/l<br>Linha de base: ${fmt.n(o.base,2)} km/l`)}"/>`});
    g+=`<path d="${d}" fill="none" stroke="var(--laranja)" stroke-width="2.5" stroke-linejoin="round"/>${pts}`;
  }
  return`<svg class="chart${o._quiet?' noanim':''}" data-cr="${chartReg(q=>monthBars(rows,q),o)}" viewBox="0 0 ${W} ${H}" role="img" aria-label="${esc(o.aria||'Investimento e economia por mês')}">${g}</svg>`;
}

/* ---------- Comparativo: barra com marca "antes" e linha da média ---------- */
function compRows(rows,o){ // rows:[{id,name,sub,val,prev,ch}]
  if(!rows.length)return`<div class="empty"><b>Sem condutores com esse indicador no período</b></div>`;
  const max=Math.max(...rows.map(r=>r.val),o.avg||0,1e-9);const avg=o.avg!=null?o.avg:0;
  return`<div class="cmpl">${rows.map((r,i)=>{const good=o.up?r.val>=avg:r.val<=avg;const w=r.val/max*100;
    return`<div class="cmr" data-act="drv-open" data-id="${esc(r.id)}"><span class="rk">${i+1}º</span><span class="ava xs">${esc((r.name||'?')[0])}</span><span class="cn">${esc(r.name)}</span>
    <div class="cmt hatch"><div class="bfill" style="--i:${Math.min(i,20)};left:0;width:${w}%;background:${good?'var(--ok)':'var(--bad)'}"></div>${r.prev!=null?`<i class="before" style="left:${Math.min(100,r.prev/max*100)}%" data-tt="Antes: ${esc(fmt.n(r.prev,o.d)+o.u)}"></i>`:''}<i class="avgl" style="left:${avg/max*100}%"></i></div>
    <span class="cv num b">${fmt.n(r.val,o.d)}${o.u}</span><span class="cd">${r.ch!=null?dl(r.ch,o.up===true,1):'<span class="muted xs">sem base</span>'}</span></div>`}).join('')}</div>`;
}
/* melhoraram · estáveis · pioraram */
function improvedStack(rows){
  return rows.map(r=>{const seg=(n,c)=>n?`<div style="flex:${n};background:${c}">${n}</div>`:'';
    return`<div class="stk"><div class="sl"><b>${esc(r.l)}</b><span class="xs muted">${esc(r.sub||'')}</span></div><div class="sb">${seg(r.b,'var(--ok)')}${seg(r.s,'var(--n300)')}${seg(r.w,'var(--bad)')}</div><div class="sd">${r.dl}</div></div>`}).join('');
}

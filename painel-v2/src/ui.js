/* =====================================================================
   UI · FORMATAÇÃO · GRÁFICOS · OVERLAYS · ESTADO
   ===================================================================== */
const _nf={};const nfc=d=>_nf[d]||(_nf[d]=new Intl.NumberFormat('pt-BR',{minimumFractionDigits:d,maximumFractionDigits:d}));
const fmt={
  n:(v,d=0)=>v==null||!isFinite(v)?'—':(v<0?'−':'')+nfc(d).format(Math.abs(v)),
  brl:(v,d=0)=>v==null||!isFinite(v)?'—':(v<0?'−':'')+'R$ '+nfc(d).format(Math.abs(v)),
  sp:(v,d=1)=>v==null||!isFinite(v)?'—':(v>0.0001?'+':v<-0.0001?'−':'')+nfc(d).format(Math.abs(v))+'%',
  dm:iso=>iso.slice(8,10)+'/'+iso.slice(5,7),
  dmy:iso=>iso.slice(8,10)+'/'+iso.slice(5,7)+'/'+iso.slice(0,4),
  kmpl:v=>v==null||!isFinite(v)?'—':nfc(2).format(v)
};
const esc=s=>String(s??'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
const $=(s,r=document)=>r.querySelector(s);const $$=(s,r=document)=>[...r.querySelectorAll(s)];
function dl(v,goodUp=true,d=1,suffix='%'){
  if(v==null||!isFinite(v))return'<span class="dl flat">—</span>';
  const good=Math.abs(v)<0.05?null:(v>0)===goodUp;const cls=good==null?'flat':good?'good':'bad';
  const s=(v>0.0001?'+':v<-0.0001?'−':'')+nfc(d).format(Math.abs(v))+suffix;return`<span class="dl ${cls}">${s}</span>`;
}
function count(v,d=0,pre='',suf=''){if(v==null||!isFinite(v))return'—';return`<span class="num" data-count="${v}" data-dec="${d}" data-pre="${esc(pre)}" data-suf="${esc(suf)}">${pre}${fmt.n(v,d)}${suf}</span>`}

const ICONS={
  home:'<path d="M3 10.5 12 3l9 7.5V21a1 1 0 0 1-1 1h-5v-7H9v7H4a1 1 0 0 1-1-1z"/>',
  shield:'<path d="M12 3 4 6v6c0 5 3.4 8.4 8 9 4.6-.6 8-4 8-9V6z"/><path d="m9 12 2 2 4-4"/>',
  coin:'<circle cx="12" cy="12" r="9"/><path d="M12 6.5v11M9.2 9.3c0-1 1.2-1.8 2.8-1.8s2.8.8 2.8 1.8-1.2 1.6-2.8 1.9-2.8.9-2.8 1.9 1.2 1.8 2.8 1.8 2.8-.8 2.8-1.8"/>',
  fuel:'<path d="M4 21V5.5A2.5 2.5 0 0 1 6.5 3h5A2.5 2.5 0 0 1 14 5.5V21M3 21h12M4 10h10"/><path d="M14 8.5h2.2A1.8 1.8 0 0 1 18 10.3V16a1.6 1.6 0 0 0 3.2 0V8.2L18.5 5.5"/>',
  dollar:'<path d="M12 2.5v19M16.7 7C16.2 5.6 14.4 4.6 12 4.6c-2.7 0-4.7 1.2-4.7 3.3 0 5 9.4 2.4 9.4 7.3 0 2-2 3.3-4.7 3.3-2.5 0-4.4-1-4.9-2.7"/>',
  trend:'<path d="M3 17 9 11l4 4 8-8M15 7h6v6"/>',
  route:'<path d="M6 21 9 3M18 21 15 3M12 4.5v3M12 10.5v3M12 16.5v3"/>',
  truck:'<path d="M2 5h11v11H2zM13 9h4l4 4v3h-8"/><circle cx="7" cy="18" r="2"/><circle cx="17" cy="18" r="2"/>',
  users:'<circle cx="9" cy="8" r="3.5"/><path d="M2.5 20c0-3.6 2.9-6 6.5-6s6.5 2.4 6.5 6M16 4.7a3.5 3.5 0 0 1 0 6.6M18 14.3c2.2.6 3.5 2.5 3.5 5.7"/>',
  sliders:'<path d="M4 7h9M17 7h3M4 17h3M11 17h9"/><circle cx="15" cy="7" r="2"/><circle cx="9" cy="17" r="2"/>',
  search:'<circle cx="11" cy="11" r="7"/><path d="m20 20-3.5-3.5"/>',
  bell:'<path d="M6 16v-5a6 6 0 0 1 12 0v5l1.5 2h-15zM10 21h4"/>',
  moon:'<path d="M20 14.5A8 8 0 0 1 9.5 4 8 8 0 1 0 20 14.5z"/>',
  sun:'<circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4"/>',
  menu:'<path d="M4 7h16M4 12h16M4 17h16"/>',plus:'<path d="M12 5v14M5 12h14"/>',check:'<path d="m5 12 5 5 9-10"/>',x:'<path d="M6 6l12 12M18 6 6 18"/>',
  download:'<path d="M12 4v11M7 11l5 5 5-5M4 20h16"/>',arrow:'<path d="M7 17 17 7M8 7h9v9"/>',msg:'<path d="M4 5h16v11H9l-5 4z"/>',
  alert:'<path d="M12 3 2 20h20zM12 10v5M12 17.5v.5"/>',copy:'<rect x="8" y="8" width="12" height="12" rx="2"/><path d="M16 8V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2h2"/>',
  edit:'<path d="M4 20h4L19 9l-4-4L4 16z"/>',refresh:'<path d="M20 11a8 8 0 0 0-14-4l-2 2M4 13a8 8 0 0 0 14 4l2-2M4 4v5h5M20 20v-5h-5"/>',
  pin:'<path d="M12 21s7-6.2 7-11.5A7 7 0 0 0 5 9.5C5 14.8 12 21 12 21z"/><circle cx="12" cy="9.5" r="2.5"/>',
  chev:'<path d="m9 6 6 6-6 6"/>',file:'<path d="M6 3h8l4 4v14H6zM14 3v4h4M9 13h6M9 17h6"/>',key:'<circle cx="8" cy="15" r="4"/><path d="m11 12 9-9M16 7l3 3"/>',
  flag:'<path d="M5 21V4M5 4h11l-2 4 2 4H5"/>',zap:'<path d="M13 2 4 14h7l-1 8 9-12h-7z"/>',user:'<circle cx="12" cy="8" r="4"/><path d="M4 21c0-4 3.6-6 8-6s8 2 8 6"/>',
  out:'<path d="M15 4h4a1 1 0 0 1 1 1v14a1 1 0 0 1-1 1h-4M10 17l-5-5 5-5M5 12h11"/>'
};
const ico=(n,s=20)=>`<svg width="${s}" height="${s}" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">${ICONS[n]||''}</svg>`;

/* ---------- Gráficos (SVG) ---------- */
const CHREG={n:0,m:new Map()};
function chartReg(fn,o){if(o._id)return o._id;const id=++CHREG.n;CHREG.m.set(id,{fn,o});return id}
function fitCharts(root=document,quiet=false){
  $$('svg.chart[data-cr]',root).forEach(svg=>{const ent=CHREG.m.get(+svg.dataset.cr);if(!ent)return;
    const cw=Math.round(svg.getBoundingClientRect().width);if(cw<120)return;if(Math.abs(cw-(ent.cw||0))<4)return;
    const html=ent.fn({...ent.o,w:cw,_id:+svg.dataset.cr,_quiet:quiet}).trim();if(!html.startsWith('<svg'))return;
    ent.cw=cw;const t=document.createElement('div');t.innerHTML=html;svg.replaceWith(t.firstChild)})
}
function lineChart(o){
  const W=o.w||720,H=o.h||250,pl=o.pl||46,pr=o.pr||14,pt=o.marks?34:(o.vlabels?26:16),pb=o.sub?42:28;const n=o.labels.length;
  const vals=o.series.flatMap(s=>s.vals).filter(v=>v!=null&&isFinite(v));
  if(!vals.length)return`<div class="empty"><b>Sem dados no período</b>Escolha outro período.</div>`;
  let lo=o.ymin!=null?o.ymin:Math.min(...vals,o.base?o.base.v:Infinity),hi=o.ymax!=null?o.ymax:Math.max(...vals,o.base?o.base.v:-Infinity);
  if(o.ymin==null){const pd=(hi-lo)*0.12||1;lo-=pd;hi+=pd}if(hi===lo)hi=lo+1;
  const nice=r=>{const p=Math.pow(10,Math.floor(Math.log10(r)));const f=r/p;return(f<1.5?1:f<3?2:f<7?5:10)*p};
  const stp=nice((hi-lo)/4);lo=Math.floor(lo/stp+1e-9)*stp;hi=Math.ceil(hi/stp-1e-9)*stp;
  const x=i=>pl+(n<=1?(W-pl-pr)/2:i*(W-pl-pr)/(n-1));const y=v=>pt+(hi-v)/(hi-lo)*(H-pt-pb);
  const yf=o.yfmt||(v=>fmt.n(v,2));let g='';
  const nt=Math.round((hi-lo)/stp);for(let k=0;k<=nt;k++){const v=lo+stp*k;const yy=y(v);g+=`<line class="grid-l" x1="${pl}" x2="${W-pr}" y1="${yy}" y2="${yy}"/><text x="${pl-8}" y="${yy+4}" text-anchor="end">${yf(v)}</text>`}
  const step=Math.max(1,Math.ceil(n/7));for(let i=0;i<n;i+=step)g+=`<text x="${x(i)}" y="${H-(o.sub?20:6)}" text-anchor="middle">${esc(o.labels[i])}</text>`;
  if(o.sub)o.sub.forEach((t,i)=>{if(t&&i%step===0)g+=`<text x="${x(i)}" y="${H-6}" text-anchor="middle" style="font-size:10.5px;font-weight:700">${esc(t)}</text>`});
  let body='';
  if(o.base){const yy=y(o.base.v);body+=`<line x1="${pl}" x2="${W-pr}" y1="${yy}" y2="${yy}" stroke="var(--n400)" stroke-width="1.5" stroke-dasharray="5 5"/><text x="${W-pr}" y="${yy-6}" text-anchor="end">${esc(o.base.label)}</text>`}
  o.series.forEach((s,si)=>{
    let d='',ar='',seg=[];
    s.vals.forEach((v,i)=>{if(v==null||!isFinite(v)){if(seg.length)flush();return}seg.push([x(i),y(v)])});flush();
    function flush(){if(!seg.length)return;d+='M'+seg.map(p=>p[0].toFixed(1)+' '+p[1].toFixed(1)).join(' L');if(s.area&&seg.length>1)ar+='M'+seg[0][0].toFixed(1)+' '+(H-pb)+' L'+seg.map(p=>p[0].toFixed(1)+' '+p[1].toFixed(1)).join(' L')+' L'+seg[seg.length-1][0].toFixed(1)+' '+(H-pb)+' Z';seg=[]}
    if(ar)body+=`<path class="areaf" d="${ar}" fill="${s.color}" fill-opacity=".14"/>`;
    if(d)body+=`<path class="draw" pathLength="100" d="${d}" fill="none" stroke="${s.color}" stroke-width="${s.w||3}" stroke-linejoin="round" stroke-linecap="round"${s.dash?` stroke-dasharray="${s.dash}" style="animation:none;stroke-dashoffset:0"`:''}/>`;
    if(s.dots)s.vals.forEach((v,i)=>{if(v!=null&&isFinite(v))body+=`<circle class="pt" style="--i:${i}" cx="${x(i)}" cy="${y(v)}" r="4.5" fill="${s.dotColors?'var(--card)':'var(--amarelo)'}" stroke="${s.dotColors?s.dotColors[i]:'var(--preto)'}" stroke-width="2.5"/>`});
  });
  if(o.vline!=null&&o.vline>=0){const mx=x(o.vline);body+=`<line x1="${mx}" x2="${mx}" y1="${pt-6}" y2="${H-pb}" stroke="var(--text)" stroke-width="1.4" stroke-dasharray="4 4" opacity=".55"/>`}
  if(o.vlabels)o.series[0].vals.forEach((v,i)=>{if(v!=null&&isFinite(v))body+=`<text class="pt" style="--i:${i};font-weight:800;font-size:12px;fill:var(--text)" x="${x(i)}" y="${y(v)-10}" text-anchor="middle">${o.vlabels(v,i)}</text>`});
  if(o.marks)o.marks.forEach(m=>{const mx=x(m.i);body+=`<line x1="${mx}" x2="${mx}" y1="${pt-4}" y2="${H-pb}" stroke="var(--text)" stroke-width="1.5" stroke-dasharray="4 4" opacity=".55"/><g class="pt" style="--i:${m.n}"><circle cx="${mx}" cy="${pt-16}" r="11" fill="var(--preto)"/><text x="${mx}" y="${pt-12}" text-anchor="middle" style="fill:#fff;font-weight:700;font-size:12px">${m.n}</text></g>`});
  let hits='';const bw=(W-pl-pr)/Math.max(1,n-1);
  for(let i=0;i<n;i++){const tip=o.tip?o.tip(i):esc(o.labels[i]);if(!tip)continue;const first=o.series[0].vals[i];
    hits+=`<g class="hit" data-tt="${esc(tip)}"><rect x="${x(i)-bw/2}" y="${pt}" width="${bw}" height="${H-pt-pb}"/><line class="xh" x1="${x(i)}" x2="${x(i)}" y1="${pt}" y2="${H-pb}"/>${first!=null&&isFinite(first)?`<circle class="xd" cx="${x(i)}" cy="${y(first)}" r="6" fill="${o.series[0].color}" stroke="var(--card)" stroke-width="2"/>`:''}</g>`}
  return`<svg class="chart${o._quiet?' noanim':''}" data-cr="${chartReg(lineChart,o)}" viewBox="0 0 ${W} ${H}" role="img" aria-label="${esc(o.aria||'Gráfico')}">${g}${body}${hits}</svg>`;
}
function spark(vals,color='var(--laranja)'){
  const v=vals.filter(x=>x!=null);if(v.length<2)return'';const lo=Math.min(...v),hi=Math.max(...v)||1,W=84,H=34;
  const pts=vals.map((x,i)=>x==null?null:[i*(W/(vals.length-1)),H-3-((x-lo)/((hi-lo)||1))*(H-6)]).filter(Boolean);
  return`<svg class="spark" viewBox="0 0 ${W} ${H}" aria-hidden="true"><path class="draw" pathLength="100" d="M${pts.map(p=>p[0].toFixed(1)+' '+p[1].toFixed(1)).join(' L')}" fill="none" stroke="${color}" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"/></svg>`;
}
function gauge(v,band){
  const col=band==='ALTO'?'var(--bad)':band==='MÉDIO'?'var(--warn)':'var(--ok)';const len=251.3,off=len*(1-clamp(v,0,100)/100);
  return`<div class="gauge"><svg viewBox="0 0 200 112" role="img" aria-label="Índice de risco ${v} de 100"><path d="M20 100 A80 80 0 0 1 180 100" fill="none" stroke="var(--soft)" stroke-width="16" stroke-linecap="round"/><path class="garc" style="--to:${off}" d="M20 100 A80 80 0 0 1 180 100" fill="none" stroke="${col}" stroke-width="16" stroke-linecap="round"/></svg><div class="gv">${count(v,0)}<small>de 100</small></div></div>`;
}
function bandBadge(b){return`<span class="badge ${b==='ALTO'?'bad':b==='MÉDIO'?'warn':'ok'}">RISCO ${b}</span>`}

/* ---------- Overlays ---------- */
const ovRoot=()=>$('#ov');
function closeOv(){const r=ovRoot();if(!r)return;const d=$('.drawer',r),m=$('.modal',r);[d,m].forEach(e=>e&&(e.style.transition='opacity .2s,transform .3s',e.style.opacity='0'));if(d)d.style.transform='translateX(100%)';setTimeout(()=>r.innerHTML='',200)}
function openDrawer(html){ovRoot().innerHTML=`<div class="scrim" data-act="ov-close"></div><aside class="drawer" role="dialog" aria-modal="true">${html}</aside>`;const d=$('.drawer');d.focus&&d.setAttribute('tabindex','-1');enhance(d,true)}
function openModal(html){ovRoot().innerHTML=`<div class="scrim" data-act="ov-close"></div><div class="modal" role="dialog" aria-modal="true">${html}</div>`;enhance($('.modal'),true);const i=$('.modal input,.modal textarea,.modal select');i&&i.focus()}
function toast(msg){const r=$('#toasts');const t=document.createElement('div');t.className='toast';t.textContent=msg;r.appendChild(t);setTimeout(()=>{t.classList.add('out');setTimeout(()=>t.remove(),300)},3200)}

/* ---------- Pós-render: contadores, abas, tooltips ---------- */
function enhance(root,anim=true){
  $$('.rise',root).forEach((e,i)=>e.style.setProperty('--i',Math.min(i,14)));
  $$('.bfill',root).forEach((e,i)=>e.style.setProperty('--i',i));
  fitCharts(root);placeTabs(root);
  try{initMaps(root)}catch(e){console.warn('mapa',e)}
  const els=$$('[data-count]',root);
  if(!anim||matchMedia('(prefers-reduced-motion: reduce)').matches)return;
  els.forEach(e=>{
    const target=parseFloat(e.dataset.count),d=+e.dataset.dec,pre=e.dataset.pre,suf=e.dataset.suf;const t0=performance.now(),dur=1000;
    const step=now=>{const p=Math.min(1,(now-t0)/dur),k=1-Math.pow(1-p,3);e.textContent=pre+fmt.n(target*k,d)+suf;if(p<1)requestAnimationFrame(step)};
    e.textContent=pre+fmt.n(0,d)+suf;requestAnimationFrame(step);
  });
}
function placeTabs(root=document){
  $$('.tabs',root).forEach(t=>{let ind=$('.tab-ind',t);if(!ind){ind=document.createElement('i');ind.className='tab-ind';t.appendChild(ind)}
    const on=$('button.on',t);if(on){ind.style.width=on.offsetWidth+'px';ind.style.transform=`translateX(${on.offsetLeft}px)`}});
}
function placeNav(){
  const nav=$('.nav');if(!nav)return;const on=$('a.on',nav);const ind=$('.nav-ind',nav);if(!ind)return;
  if(!on){ind.style.opacity=0;return}
  const first=!ind.dataset.placed;if(first)ind.style.transition='none';
  ind.style.opacity=1;ind.style.transform=`translateY(${on.offsetTop}px)`;ind.style.height=on.offsetHeight+'px';
  if(first){ind.dataset.placed='1';void ind.offsetHeight;requestAnimationFrame(()=>ind.style.transition='')}
}
function tabsHtml(items,cur,act,extra=''){
  return`<div class="tabs" role="tablist">${items.map(([k,l,c])=>`<button role="tab" class="${k===cur?'on':''}" data-act="${act}" data-v="${k}" aria-selected="${k===cur}">${l}${c!=null?`<span class="cnt">${c}</span>`:''}</button>`).join('')}${extra}</div>`;
}
function segHtml(items,cur,act){return`<div class="seg">${items.map(([k,l])=>`<button class="${k===cur?'on':''}" data-act="${act}" data-v="${k}">${l}</button>`).join('')}</div>`}

/* ---------- Estado persistente (preferências locais; nada de dado de cliente) ---------- */
const LS_KEY='gbx-frota-v3';
function defaultStore(){return{rules:{...DEFAULT_RULES},flags:{},contracts:{},phones:{},notifRead:false}}
function loadStore(){const d=defaultStore();try{const s=JSON.parse(localStorage.getItem(LS_KEY)||'null');if(s){return{...d,...s,rules:{...d.rules,...s.rules}}}}catch(e){}return d}
const ST=loadStore();
function persist(){try{localStorage.setItem(LS_KEY,JSON.stringify({rules:ST.rules,flags:ST.flags,contracts:ST.contracts}))}catch(e){}}

/* ---------- Estado de interface ---------- */
const S={client:'via',period:{preset:'month',from:'',to:''},compare:'mes',theme:'light',collapsed:false,
  u:{ovMetric:'kmpl',retTab:'cliente',fleetTab:'visao',fleetFilial:'',fleetRank:'veh',drvTab:'resumo',drvSel:null,drvFilter:'all',drvSort:'money',drvQ:'',drvDet:'resumo',msgDay:'',cmpMetric:'kmpl',weekSel:null,premMonth:'',dataTab:'regras',mapFilter:'all',evoGran:'semana',evoMetric:'kmpl',dayMetric:'br100',dayView:'grafico'}};
try{const p=JSON.parse(localStorage.getItem('gbx-prefs')||'null');if(p){S.theme=p.theme||'light';S.collapsed=!!p.collapsed;if(p.client)S.client=p.client}}catch(e){}
function savePrefs(){try{localStorage.setItem('gbx-prefs',JSON.stringify({theme:S.theme,collapsed:S.collapsed,client:S.client}))}catch(e){}}

(()=>{
"use strict";
const key="inkveil.personal.dex.v1",uuid=s=>typeof s==="string"&&/^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$/i.test(s);
const sw=s=>typeof s==="string"&&/^sw-[1-9][0-9]{0,9}$/.test(s)&&Number(s.slice(3))<=2147483647,valid=s=>uuid(s)||sw(s),same=(id,c)=>sw(id)?sw(c):uuid(c);
let state={bookmarks:{},history:{},read:{}};
function load(){if(account)return;try{const s=JSON.parse(localStorage.getItem(key)||"{}");for(const k of Object.keys(state))if(s[k]&&typeof s[k]==="object"&&!Array.isArray(s[k]))state[k]=s[k];}catch{}}
let account=false,revision=0,chain=Promise.resolve(),blocked=false;
function warning(message){let n=document.getElementById("personal-sync-status");if(!n){n=document.createElement("p");n.id="personal-sync-status";n.setAttribute("role","status");const target=document.querySelector(".iv-library-meta,.iv-settings-card");if(target)target.prepend(n);else if(message.includes("gagal")||message.includes("tidak tersedia"))document.querySelector("main")?.prepend(n);else return;}n.textContent=message;}
async function initialize(){load();try{const r=await fetch("/api/dex-library",{cache:"no-store"});if(!r.ok)throw Error();const d=await r.json();if(d.authenticated){account=true;revision=d.revision;state=d.state;warning("Pustaka gabungan: akun MySQL.");}else warning("Pustaka gabungan: mode browser lokal.");}catch{blocked=true;warning("Library tidak tersedia; perubahan dinonaktifkan hingga reload.");}}
function save(){if(blocked)return;if(account){const snapshot=JSON.parse(JSON.stringify(state));chain=chain.then(async()=>{if(blocked)return;try{const r=await fetch("/api/dex-library",{method:"POST",headers:{"Content-Type":"application/json","X-Inkveil-Request":"dex-library"},body:JSON.stringify({revision,state:snapshot}),cache:"no-store"});if(!r.ok)throw Error();const d=await r.json();if(!d.authenticated)throw Error();revision=d.revision;}catch{blocked=true;warning("Sinkronisasi gagal atau data berubah. Reload sebelum mengubah library lagi.");}});return;}
try{localStorage.setItem(key,JSON.stringify(state));}catch{const n=document.getElementById("dex-status");if(n)n.textContent="Penyimpanan pustaka gagal.";}}
const lang=s=>s==="en"?"en":"id";
const detail=(id,l)=>sw(id)?"/suwayomi/"+id.slice(3):"/live/"+id+"?lang="+lang(l);
const chapter=(id,c,l,page)=>{if(sw(id)&&sw(c))return "/suwayomi/"+id.slice(3)+"/"+c.slice(3)+(Number.isInteger(page)&&page>0?"#page-"+page:"");let q="";try{q=localStorage.getItem("inkveil-quality")==="original"?"&quality=original":"";}catch{}return "/live/"+id+"/"+c+"?lang="+lang(l)+q;};
const sorted=o=>Object.values(o).filter(x=>x&&valid(x.id)).sort((a,b)=>(b.time||0)-(a.time||0));
const trim=(o,max)=>Object.fromEntries(sorted(o).slice(0,max).map(x=>[x.id,x]));
function entry(item,reading){const box=document.createElement("article");box.className="personal-entry";const title=document.createElement("strong");title.textContent=item.title||item.id;const a=document.createElement("a");a.textContent=reading?"Lanjut chapter "+(item.number||""):sw(item.id)?"Detail Suwayomi":"Detail MangaDex";a.href=reading&&same(item.id,item.chapter)?chapter(item.id,item.chapter,item.lang,item.page):detail(item.id,item.lang);box.append(title,a);if(!reading){const b=document.createElement("button");b.type="button";b.className="secondary";b.textContent="Hapus bookmark";b.onclick=()=>{if(blocked)return;load();delete state.bookmarks[item.id];save();render();};box.append(b);}return box;}
function fill(id,items,reading){const el=document.getElementById(id);if(el)el.replaceChildren(...items.map(x=>entry(x,reading)));}
function render(){const books=sorted(state.bookmarks),history=sorted(state.history);const stats=document.getElementById("iv-reading-stats");if(stats)stats.textContent=books.length+" komik tersimpan · "+history.length+" komik dalam riwayat · "+Object.keys(state.read).length+" chapter ditandai dibaca";fill("personal-bookmarks",books,false);fill("personal-history",history,true);fill("personal-continue-list",history.slice(0,6),true);const section=document.getElementById("personal-continue");if(section)section.hidden=!history.length;}
async function start(){await initialize();
const root=document.querySelector("[data-personal-id]");
if(root&&valid(root.dataset.personalId)){
 const id=root.dataset.personalId,title=root.dataset.personalTitle||id,l=lang(root.dataset.personalLang),latest=root.dataset.personalLatest;
 const buttons=[...document.querySelectorAll("[data-personal-bookmark]")];
 function mark(){for(const b of buttons){b.textContent=state.bookmarks[id]?"Hapus bookmark":"Simpan komik";b.setAttribute("aria-pressed",String(!!state.bookmarks[id]));}}
 buttons.forEach(b=>b.onclick=()=>{if(blocked)return;load();if(state.bookmarks[id])delete state.bookmarks[id];else state.bookmarks[id]={id,title,lang:l,latest:same(id,latest)?latest:"",time:Date.now()};state.bookmarks=trim(state.bookmarks,500);save();mark();});
 mark();
 const saved=state.history[id],resume=root.querySelector("[data-personal-resume]");
 if(resume&&saved&&same(id,saved.chapter)){resume.href=chapter(id,saved.chapter,saved.lang,saved.page);resume.hidden=false;}
 const notice=root.querySelector("[data-personal-update]"),bm=state.bookmarks[id];
 if(notice&&bm&&uuid(latest)&&bm.lang===l&&bm.latest&&bm.latest!==latest){notice.textContent="Ada perubahan pada chapter terbaru sejak terakhir dicek.";notice.hidden=false;bm.latest=latest;save();}
 if(bm&&uuid(latest)&&!bm.latest){bm.latest=latest;bm.lang=l;save();}
 const c0=root.dataset.personalReading;
 if(sw(id)&&sw(c0)&&!blocked){
 const pages=[...root.querySelectorAll(".iv-suwayomi-pages img")];let last=-1,timer;
 pages.forEach((img,i)=>img.id="page-"+i);
 const hash=/^#page-([0-9]+)$/.exec(location.hash);if(hash&&pages[Number(hash[1])])pages[Number(hash[1])].scrollIntoView();
 const observer=new IntersectionObserver(entries=>{for(const e of entries){if(!e.isIntersecting||!e.target.complete||!e.target.naturalWidth)continue;const i=pages.indexOf(e.target);if(i<0||i===last)continue;last=i;clearTimeout(timer);timer=setTimeout(()=>{if(blocked)return;load();if(!state.history[id]||state.history[id].chapter!==c0)return;state.history[id].page=i;state.history[id].time=Date.now();if(i===pages.length-1){state.read[c0]=Date.now();state.read=Object.fromEntries(Object.entries(state.read).sort((a,b)=>b[1]-a[1]).slice(0,2000));}save();},800);}}, {threshold:0,rootMargin:"-20% 0px -60% 0px"});
 pages.forEach(img=>{observer.observe(img);img.addEventListener("load",()=>{observer.unobserve(img);observer.observe(img)},{once:true});});
 }
 root.querySelectorAll("[data-personal-chapter]").forEach(a=>{if(state.read[a.dataset.personalChapter])a.classList.add("personal-read");});
 const c=root.dataset.personalReading;
 if(same(id,c)&&!blocked){state.history[id]={id,title,lang:l,chapter:c,number:(root.dataset.personalNumber||"").slice(0,80),page:saved&&saved.chapter===c?saved.page||0:0,time:Date.now()};state.history=trim(state.history,100);save();
 window.addEventListener("inkveil-page-loaded",e=>{if(!blocked&&e.detail.index===e.detail.total-1){load();state.read[c]=Date.now();state.read=Object.fromEntries(Object.entries(state.read).sort((a,b)=>b[1]-a[1]).slice(0,2000));save();}});}
}
document.getElementById("personal-clear")?.addEventListener("click",()=>{if(!blocked&&confirm("Hapus riwayat MangaDex dan Suwayomi?")){load();state.history={};state.read={};save();render();}});
const rail=document.getElementById("personal-featured-list"),cards=[...document.querySelectorAll(".dex-grid .dex-card")];
if(rail&&cards.length&&!new URLSearchParams(location.search).get("q")&&Number(new URLSearchParams(location.search).get("offset")||0)===0){rail.replaceChildren(...cards.slice(0,6).map(x=>x.cloneNode(true)));document.getElementById("personal-featured").hidden=false;document.querySelectorAll("[data-rail]").forEach(b=>b.onclick=()=>rail.scrollBy({left:Number(b.dataset.rail)*rail.clientWidth*.8,behavior:"smooth"}));}
document.addEventListener("click",e=>{const a=e.target.closest?.("a[href]");if(!a)return;const u=new URL(a.href,location.href);if(u.origin===location.origin&&u.pathname==="/"&&!u.search){try{const source=localStorage.getItem("inkveil-default-source"),l=localStorage.getItem("inkveil-default-lang");if(["local","mangadex"].includes(source))u.searchParams.set("source",source);if(["id","en"].includes(l))u.searchParams.set("lang",l);a.href=u.href;}catch{}}if(u.origin===location.origin&&/^\/live\/[a-f0-9-]+\/[a-f0-9-]+$/i.test(u.pathname)&&!u.searchParams.has("quality")){try{if(localStorage.getItem("inkveil-quality")==="original"){u.searchParams.set("quality","original");a.href=u.href;}}catch{}}});
render();
const updates=document.getElementById("iv-check-updates"),updateStatus=document.getElementById("iv-update-status"),updateResults=document.getElementById("iv-update-results");
let updateOffset=0;
if(updates&&updateStatus&&updateResults)updates.addEventListener("click",async()=>{
 const books=sorted(state.bookmarks);if(blocked||!books.length){updateStatus.textContent="Belum ada bookmark yang dapat diperiksa.";return;}
 if(updateOffset>=books.length)updateOffset=0;
 const items=books.slice(updateOffset,updateOffset+5);updates.disabled=true;updateStatus.textContent="Memeriksa "+items.length+" judul...";
 try{const r=await fetch("/api/reading-updates",{method:"POST",headers:{"Content-Type":"application/json","X-Inkveil-Request":"reading-updates"},body:JSON.stringify({items}),cache:"no-store"});if(!r.ok)throw Error();const result=await r.json();
 updateResults.replaceChildren(...result.map(v=>{const item=items.find(i=>i.id===v.id),p=document.createElement("p");p.textContent=(item?.title||v.id)+" — "+(!v.available?"Belum dapat diperiksa":v.changed?"Chapter terbaru berbeda":!item?.latest?"Belum ada pembanding":"Tidak berubah");return p;}));updateOffset+=items.length;updateStatus.textContent="Diperiksa "+updateOffset+" dari "+books.length+" bookmark. Suwayomi menggunakan metadata tersimpan.";updates.textContent=updateOffset<books.length?"Cek 5 judul berikutnya":"Periksa ulang";
 }catch{updateStatus.textContent="Pemeriksaan gagal. Coba lagi."}finally{updates.disabled=false;}
});
const search=document.querySelector('#catalog-search input[name=q]');
if(search){const form=search.form;const list=document.createElement("datalist");list.id="iv-recent-searches";search.setAttribute("list",list.id);form.append(list);let recent=[];try{const a=JSON.parse(localStorage.getItem("inkveil-search-history")||"[]");if(Array.isArray(a))recent=a.filter(x=>typeof x==="string"&&x.length<=200).slice(0,12);}catch{}
 const suggestions=[...new Set([...recent,...sorted(state.bookmarks).map(x=>x.title)].filter(x=>typeof x==="string"&&x.length<=200))].slice(0,30);for(const s of suggestions){const o=document.createElement("option");o.value=s;list.append(o);}
 form.addEventListener("submit",()=>{const q=search.value.trim();if(q&&q.length<=200)try{localStorage.setItem("inkveil-search-history",JSON.stringify([q,...recent.filter(x=>x!==q)].slice(0,12)));}catch{}});
}
const settings=document.getElementById("personal-settings-form");if(settings){for(const [name,key,values] of [["quality","inkveil-quality",["saver","original"]],["width","inkveil-dex-width",["normal","wide","compact"]],["source","inkveil-default-source",["","local","mangadex"]],["lang","inkveil-default-lang",["id","en"]]]){const el=settings.elements[name];try{const v=localStorage.getItem(key);if(values.includes(v))el.value=v;}catch{}}settings.addEventListener("submit",e=>{e.preventDefault();try{for(const [name,key] of [["quality","inkveil-quality"],["width","inkveil-dex-width"],["source","inkveil-default-source"],["lang","inkveil-default-lang"]])localStorage.setItem(key,settings.elements[name].value);warning("Preferensi disimpan di browser.");}catch{warning("Gagal menyimpan preferensi.");}});}
}
start();
})();
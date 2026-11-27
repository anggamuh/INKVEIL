(()=>{
"use strict";
const key="inkveil.personal.dex.v1",uuid=s=>typeof s==="string"&&/^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$/i.test(s);
let state={bookmarks:{},history:{},read:{}};
function load(){try{const s=JSON.parse(localStorage.getItem(key)||"{}");for(const k of Object.keys(state))if(s[k]&&typeof s[k]==="object"&&!Array.isArray(s[k]))state[k]=s[k];}catch{}}
load();
function save(){try{localStorage.setItem(key,JSON.stringify(state));}catch{const n=document.getElementById("dex-status");if(n)n.textContent="Penyimpanan pustaka gagal.";}}
const lang=s=>s==="en"?"en":"id";
const detail=(id,l)=>"/live/"+id+"?lang="+lang(l);
const chapter=(id,c,l)=>{let q="";try{q=localStorage.getItem("inkveil-quality")==="original"?"&quality=original":"";}catch{}return "/live/"+id+"/"+c+"?lang="+lang(l)+q;};
const sorted=o=>Object.values(o).filter(x=>x&&uuid(x.id)).sort((a,b)=>(b.time||0)-(a.time||0));
const trim=(o,max)=>Object.fromEntries(sorted(o).slice(0,max).map(x=>[x.id,x]));
function entry(item,reading){const box=document.createElement("article");box.className="personal-entry";const title=document.createElement("strong");title.textContent=item.title||item.id;const a=document.createElement("a");a.textContent=reading?"Lanjut chapter "+(item.number||""):"Detail MangaDex";a.href=reading&&uuid(item.chapter)?chapter(item.id,item.chapter,item.lang):detail(item.id,item.lang);box.append(title,a);if(!reading){const b=document.createElement("button");b.type="button";b.className="secondary";b.textContent="Hapus bookmark";b.onclick=()=>{load();delete state.bookmarks[item.id];save();render();};box.append(b);}return box;}
function fill(id,items,reading){const el=document.getElementById(id);if(el)el.replaceChildren(...items.map(x=>entry(x,reading)));}
function render(){const books=sorted(state.bookmarks),history=sorted(state.history);fill("personal-bookmarks",books,false);fill("personal-history",history,true);fill("personal-continue-list",history.slice(0,6),true);const section=document.getElementById("personal-continue");if(section)section.hidden=!history.length;}
const root=document.querySelector("[data-personal-id]");
if(root&&uuid(root.dataset.personalId)){
 const id=root.dataset.personalId,title=root.dataset.personalTitle||id,l=lang(root.dataset.personalLang),latest=root.dataset.personalLatest;
 const buttons=[...document.querySelectorAll("[data-personal-bookmark]")];
 function mark(){for(const b of buttons){b.textContent=state.bookmarks[id]?"Hapus bookmark":"Simpan komik";b.setAttribute("aria-pressed",String(!!state.bookmarks[id]));}}
 buttons.forEach(b=>b.onclick=()=>{load();if(state.bookmarks[id])delete state.bookmarks[id];else state.bookmarks[id]={id,title,lang:l,latest:uuid(latest)?latest:"",time:Date.now()};state.bookmarks=trim(state.bookmarks,500);save();mark();});
 mark();
 const saved=state.history[id],resume=root.querySelector("[data-personal-resume]");
 if(resume&&saved&&uuid(saved.chapter)){resume.href=chapter(id,saved.chapter,saved.lang);resume.hidden=false;}
 const notice=root.querySelector("[data-personal-update]"),bm=state.bookmarks[id];
 if(notice&&bm&&uuid(latest)&&bm.lang===l&&bm.latest&&bm.latest!==latest){notice.textContent="Ada perubahan pada chapter terbaru sejak terakhir dicek.";notice.hidden=false;bm.latest=latest;save();}
 if(bm&&uuid(latest)&&!bm.latest){bm.latest=latest;bm.lang=l;save();}
 root.querySelectorAll("[data-personal-chapter]").forEach(a=>{if(state.read[a.dataset.personalChapter])a.classList.add("personal-read");});
 const c=root.dataset.personalReading;
 if(uuid(c)){state.history[id]={id,title,lang:l,chapter:c,number:root.dataset.personalNumber,time:Date.now()};state.history=trim(state.history,100);save();
 window.addEventListener("inkveil-page-loaded",e=>{if(e.detail.index===e.detail.total-1){load();state.read[c]=Date.now();state.read=Object.fromEntries(Object.entries(state.read).sort((a,b)=>b[1]-a[1]).slice(0,2000));save();}});}
}
document.getElementById("personal-clear")?.addEventListener("click",()=>{if(confirm("Hapus riwayat MangaDex?")){load();state.history={};state.read={};save();render();}});
const rail=document.getElementById("personal-featured-list"),cards=[...document.querySelectorAll(".dex-grid .dex-card")];
if(rail&&cards.length&&(!location.search||new URLSearchParams(location.search).get("offset")==="0")){rail.replaceChildren(...cards.slice(0,6).map(x=>x.cloneNode(true)));document.getElementById("personal-featured").hidden=false;document.querySelectorAll("[data-rail]").forEach(b=>b.onclick=()=>rail.scrollBy({left:Number(b.dataset.rail)*rail.clientWidth*.8,behavior:"smooth"}));}
document.addEventListener("click",e=>{const a=e.target.closest?.("a[href]");if(!a)return;const u=new URL(a.href,location.href);if(u.origin===location.origin&&/^\/live\/[a-f0-9-]+\/[a-f0-9-]+$/i.test(u.pathname)&&!u.searchParams.has("quality")){try{if(localStorage.getItem("inkveil-quality")==="original"){u.searchParams.set("quality","original");a.href=u.href;}}catch{}}});
render();
})();
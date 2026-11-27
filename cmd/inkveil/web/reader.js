(() => {
"use strict";
const key = "inkveil.library.v1";
const valid = s => typeof s === "string" && /^[a-zA-Z0-9][a-zA-Z0-9_-]{0,79}$/.test(s);
let state = { bookmarks: [], history: [], width: "normal" };
try {
 const saved = JSON.parse(localStorage.getItem(key) || "{}");
 if (Array.isArray(saved.bookmarks)) state.bookmarks = saved.bookmarks.filter(valid).slice(0,500);
 if (Array.isArray(saved.history)) state.history = saved.history.filter(x=>x && valid(x.slug)&&valid(x.chapter)).slice(0,100);
 if (["normal","wide","compact"].includes(saved.width)) state.width=saved.width;
} catch (_) {}
let accountMode=false;
function save(){try{if(!accountMode)localStorage.setItem(key,JSON.stringify(state));else{const guest=JSON.parse(localStorage.getItem(key)||"{}");guest.width=state.width;localStorage.setItem(key,JSON.stringify(guest));}}catch(_){}}
async function remote(action){const res=await fetch("/api/library",action?{method:"POST",headers:{"Content-Type":"application/json","X-Inkveil-Request":"library"},body:JSON.stringify(action),cache:"no-store"}:{cache:"no-store"});if(!res.ok)throw Error("Library unavailable");const data=await res.json();if(action&&!data.authenticated)throw Error("Session expired");return data;}
function applyRemote(data){state.bookmarks=data.bookmarks;state.history=data.history;refresh();}
function reportError(){const msg=document.getElementById("library-mode");if(msg)msg.textContent="Sinkronisasi gagal. Muat ulang atau login kembali.";else alert("Sinkronisasi gagal. Muat ulang atau login kembali.");}
async function initialize(){
try{const data=await remote();accountMode=data.authenticated;if(accountMode)applyRemote(data);const mode=document.getElementById("library-mode");if(mode)mode.textContent=accountMode?"Library tersinkron ke akun MySQL.":"Mode tamu — tersimpan di browser ini.";}catch(_){reportError();return;}

function refresh(){
 document.querySelectorAll("[data-bookmark]").forEach(b=>{
 const on=state.bookmarks.includes(b.dataset.bookmark);b.setAttribute("aria-pressed",String(on));b.textContent=on?"Hapus bookmark":"Simpan komik";
 });
 let count=0;
 document.querySelectorAll("[data-library-bookmark]").forEach(el=>{el.hidden=!state.bookmarks.includes(el.dataset.libraryBookmark);if(!el.hidden)count++;});
 const empty=document.getElementById("bookmark-empty");if(empty)empty.hidden=count>0;
 for(const id of ["history-list","continue-list"]){
  const list=document.getElementById(id);if(!list)continue;list.replaceChildren();
  for(const item of state.history.slice(0,id==="continue-list"?5:100)){
   const article=document.createElement("article"),title=document.createElement("strong"),link=document.createElement("a");
   title.textContent=typeof item.title==="string"?item.title:item.slug;link.textContent="Lanjut chapter "+item.chapter;link.className="chapter";link.href="/read/"+encodeURIComponent(item.slug)+"/"+encodeURIComponent(item.chapter);
   article.append(title,link);list.append(article);
  }
 }
 const historyEmpty=document.getElementById("history-empty");if(historyEmpty)historyEmpty.hidden=state.history.length>0;
 const section=document.getElementById("continue-section");if(section)section.hidden=state.history.length===0;
}
document.querySelectorAll("[data-bookmark]").forEach(b=>b.addEventListener("click",async()=>{
 const slug=b.dataset.bookmark;if(!valid(slug))return;
 if(accountMode){b.disabled=true;try{applyRemote(await remote({action:"bookmark",slug,saved:!state.bookmarks.includes(slug)}));}catch(_){reportError();}finally{b.disabled=false;}return;}
 state.bookmarks=state.bookmarks.includes(slug)?state.bookmarks.filter(x=>x!==slug):[slug,...state.bookmarks].slice(0,500);save();refresh();
}));
const reader=document.querySelector("[data-reader-slug]");
if(reader){
 const slug=reader.dataset.readerSlug,chapter=reader.dataset.readerChapter;
 if(valid(slug)&&valid(chapter)&&accountMode){try{applyRemote(await remote({action:"read",slug,chapter}));}catch(_){reportError();}}
 else if(valid(slug)&&valid(chapter)){state.history=[{slug,chapter,title:reader.querySelector("h2")?.textContent||slug},...state.history.filter(x=>x.slug!==slug)].slice(0,100);save();}
 const width=document.getElementById("reader-width");if(width){width.value=state.width;reader.dataset.width=state.width;width.addEventListener("change",()=>{if(["normal","wide","compact"].includes(width.value)){state.width=width.value;reader.dataset.width=state.width;save();}});}
 document.getElementById("reader-top")?.addEventListener("click",()=>window.scrollTo({top:0,behavior:"smooth"}));
 document.addEventListener("keydown",e=>{if(e.ctrlKey||e.altKey||e.metaKey||e.shiftKey||e.target.closest?.("input,select,textarea,button,a,[contenteditable]"))return;const selector=e.key==="ArrowLeft"?"[data-prev]":e.key==="ArrowRight"?"[data-next]":null;if(selector){const a=document.querySelector(selector);if(a){e.preventDefault();location.assign(a.href);}}});
}
document.getElementById("clear-history")?.addEventListener("click",async()=>{if(confirm("Hapus seluruh riwayat baca?")){if(accountMode){try{applyRemote(await remote({action:"clear"}));}catch(_){reportError();}return;}state.history=[];save();refresh();}});
refresh();
}
initialize();
})();
(()=>{"use strict";
const pages=[...document.querySelectorAll("img[data-sw-page]")];if(!pages.length)return;
const status=document.getElementById("sw-load-status"),pending=[],queued=new Set(),loaded=new Set(),failed=new Set();let active=0;
function update(){if(status)status.textContent=loaded.size+"/"+pages.length+" halaman dimuat"+(failed.size?" · "+failed.size+" gagal":"");}
const pause=ms=>new Promise(resolve=>setTimeout(resolve,ms));
async function load(img){
 const path=img.getAttribute("data-sw-page");if(!path||!/^\/sw-page\/[a-f0-9]{64}\/[0-9]+$/.test(path))throw Error("Invalid page URL");
 const r=await fetch(path,{cache:"no-store",credentials:"same-origin",signal:AbortSignal.timeout(28000)});
 if(!r.ok){const e=new Error("HTTP "+r.status);e.status=r.status;throw e;}
 const blob=await r.blob();if(!blob.type.startsWith("image/"))throw Error("Invalid image");
 const url=URL.createObjectURL(blob);
 try{await new Promise((resolve,reject)=>{img.onload=resolve;img.onerror=reject;img.src=url;});}finally{img.onload=null;img.onerror=null;URL.revokeObjectURL(url);}
}
async function run(img){
 const button=img.parentElement.querySelector(".sw-retry");
 try{for(let n=0;;n++){try{await load(img);break;}catch(e){if(n>=2||![502,503,504].includes(e.status))throw e;await pause(1500*(n+1));}}
 loaded.add(img);failed.delete(img);img.parentElement.classList.add("sw-loaded");button.hidden=true;
 }catch{failed.add(img);button.hidden=false;}finally{active--;queued.delete(img);update();pump();}
}
function pump(){while(active<2&&pending.length){active++;run(pending.shift());}}
function enqueue(img){if(!img||!img.matches("img[data-sw-page]")||loaded.has(img)||queued.has(img))return;queued.add(img);pending.push(img);pump();}
const observer="IntersectionObserver" in window?new IntersectionObserver(entries=>{for(const e of entries)if(e.isIntersecting){observer.unobserve(e.target);enqueue(e.target.querySelector("img[data-sw-page]"));}},{rootMargin:"600px"}):null;
pages.forEach(img=>{img.parentElement.querySelector(".sw-retry").onclick=()=>enqueue(img);if(observer)observer.observe(img.parentElement);else enqueue(img);});
update();
})();
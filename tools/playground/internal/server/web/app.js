const $=id=>document.getElementById(id);
const common=`body{font-family:Arial,'PingFang SC',sans-serif;color:#253044;font-size:14px;line-height:1.65}h1{font-size:30px;line-height:1.25;margin:0 0 12px}h2{font-size:19px;margin:26px 0 10px}h3{font-size:15px;margin:16px 0 7px}p{margin:0 0 12px}table{border-collapse:collapse;width:100%;margin:14px 0}th,td{border:1px solid #dce2e9;padding:9px 11px;text-align:left}th{background:#eff3f8;font-weight:600}ul,ol{padding-left:25px;margin:10px 0}li{margin-bottom:6px}a{color:#3158c8}small{font-size:11px;color:#7a8698}.muted{color:#7a8698}.label{font-size:11px;letter-spacing:2px;color:#3158c8;margin-bottom:14px}.lead{font-size:16px;line-height:1.8;color:#59677b}blockquote{margin:18px 0;padding-left:16px;border-left:3px solid #3158c8;color:#64748b}pre{font:12px/1.7 monospace;background:#f3f5f8;padding:14px;white-space:pre-wrap}code{font-family:monospace;background:#f0f3f6}hr{border:0;border-top:1px solid #dce2e9;margin:20px 0}`;
const wrap=(title,body)=>`<!doctype html>\n<html lang="zh-CN">\n<head>\n<meta charset="UTF-8">\n<title>${title}</title>\n<style>\n${common}\n</style>\n</head>\n<body>\n${body}\n</body>\n</html>`;
const samples={
report:wrap('季度项目进展报告',`<p class="label">PROJECT UPDATE · 2026 Q3</p>
<h1>让每一次交付<br>都有清晰的进展。</h1>
<p class="lead">季度项目进展报告</p>
<p class="muted">产品与研发团队 · 2026 年 10 月 20 日</p>
<hr>
<h2>01 本期概览</h2>
<p>本季度围绕<strong>产品体验、交付效率与质量保障</strong>推进工作。核心流程已完成更新，后续将继续完善复杂场景，并根据实际使用反馈迭代。</p>
<p>这份文档由 HTML 生成。导出后，您可以在 Word 中<strong>插入新段落</strong>、调整文字和表格，后续内容会随之重新排版。</p>
<h2>02 关键进展</h2>
<table>
  <thead><tr><th>工作方向</th><th>本期进展</th><th>状态</th></tr></thead>
  <tbody>
    <tr><td>编辑体验</td><td>完成核心交互与基础样式统一</td><td>已完成</td></tr>
    <tr><td>文档导出</td><td>接入段落、列表、图片与表格</td><td>验证中</td></tr>
    <tr><td>质量保障</td><td>补充长文与复杂表格验证样本</td><td>进行中</td></tr>
  </tbody>
</table>
<h2>03 下一步计划</h2>
<ol>
  <li>收集真实文档，补充更多样式与分页场景。</li>
  <li>在 Word 中增加整段内容，检查自动重排和跨页。</li>
  <li>对比原始 HTML 与实际文档，持续提升版式还原。</li>
</ol>
<p><small>示例数据仅用于测试排版。您可以直接替换左侧 HTML。</small></p>`),
contract:wrap('长文编辑与跨页测试',`<p class="label">DOCUMENT FLOW TEST</p>
<h1>长文编辑与跨页测试</h1>
<p class="lead">用于验证连续编辑、段落重排、标题分页以及原生列表。</p>
${Array.from({length:12},(_,i)=>`<h2>${String(i+1).padStart(2,'0')} ${['项目目标','服务内容','交付安排','质量要求','协作机制','变更流程','信息管理','验收安排','维护计划','资料归档','持续改进','附加说明'][i]}</h2>
<p>各参与方依据实际需求推进工作，保持信息同步，并及时记录重要决定。所有交付内容应保持清晰、完整且便于后续维护。本文段落用于测试 Word 的自然排版与自动分页，您可以在任意位置增加多段文字，观察后续章节如何移动。</p>
<p>在修改文档时，正文应延续当前段落样式。较长的内容应自然流入下一页，表格与图片需要保持合理位置，避免固定高度造成文字截断。</p>
${i%3===0?'<ul><li>明确责任人与完成时间。</li><li>检查交付内容与验收条件。</li><li>记录结果并同步后续安排。</li></ul>':''}`).join('\n')}`),
table:wrap('表格与合并单元格',`<p class="label">TABLE STRUCTURE</p>
<h1>采购清单</h1>
<p class="muted">原生表格 · 横向与纵向合并 · 表头重复</p>
<table>
<thead><tr><th>分类</th><th>项目名称</th><th>数量</th><th>单价</th><th>小计</th></tr></thead>
<tbody>
<tr><td rowspan="2">设备</td><td>工作站</td><td>2</td><td>8,000</td><td>16,000</td></tr>
<tr><td>显示器</td><td>4</td><td>2,000</td><td>8,000</td></tr>
<tr><td>服务</td><td>部署与培训</td><td>1</td><td>3,000</td><td>3,000</td></tr>
<tr><td colspan="4" style="text-align:right;font-weight:bold">合计</td><td><strong>27,000</strong></td></tr>
</tbody></table>
<h2>交付说明</h2>
<ol start="3"><li>设备配置以确认清单为准。</li><li>完成部署后安排操作培训。<ul><li>基础使用方法</li><li>文档导出与编辑</li></ul></li><li>验收后归档。</li></ol>
<h2>文字样式测试</h2>
<p><strong>粗体文字</strong> · <em>斜体文字</em> · <u>下划线</u> · <s>删除线</s> · H<sub>2</sub>O · x<sup>2</sup></p>
<p>这是一个<a href="https://example.com">可点击的超链接</a>，以及带有<span style="background:#fff2bc;color:#8a5c00">背景色的文字</span>。</p>
<pre>func exportDocument() {\n    // 代码块保留缩进与换行\n    return "editable DOCX"\n}</pre>`),
layout:wrap('多列布局与图文测试',`<p class="label">LAYOUT & MEDIA</p>
<h1>一份内容，多种表达。</h1>
<p class="lead">横向 Flex 布局转换为可编辑表格，文字继续保持原生段落。</p>
<div style="display:flex;gap:18px;margin:22px 0">
 <div style="flex:1;background:#eff4fc;padding:18px"><h3 style="margin:0 0 8px;color:#3158c8">产品设计</h3><p style="margin:0">以清晰的信息结构与一致的视觉表达，帮助使用者完成任务。</p></div>
 <div style="flex:1;background:#f1f6f2;padding:18px"><h3 style="margin:0 0 8px;color:#367654">工程实现</h3><p style="margin:0">保留原生段落、列表和表格，让导出的文档可以持续编辑。</p></div>
</div>
<h2>图片与正文</h2>
<p><img alt="抽象几何插图" width="450" height="170" src="data:image/svg+xml;base64,${btoa('<svg xmlns="http://www.w3.org/2000/svg" width="450" height="170"><rect width="450" height="170" fill="#eef2fa"/><circle cx="110" cy="88" r="55" fill="#b8c9ed"/><rect x="196" y="40" width="65" height="90" fill="#3158c8"/><path d="M320 130 L365 40 L410 130 Z" fill="#90b5a4"/></svg>')}"></p>
<p>图片嵌入 DOCX 文件，可以在 Word 中移动或调整尺寸。新增文字时，行内图片会随正文一起移动。</p>
<h2>后续编辑</h2>
<p>尝试在这里添加一段更长的说明，或在前面的表格单元格里增加多段内容。单元格允许向下扩展，文档继续自动分页。</p>`)
};
Object.assign(samples,complexSamples);
let editor='html',preview='html',latest=null,pdfURL=null,docURL=null,controller=null,dirty=false,timer=null,renderID=0,pageURLs=[];
if(typeof batchSampleKeys!=='undefined'){const group=document.createElement('optgroup');group.label='新增复杂案例 · '+batchSampleKeys.length+' 个';for(const key of batchSampleKeys){const option=document.createElement('option');option.value=key;option.textContent=complexSampleInfo[key].title;group.append(option)}$('sample').append(group);const link=document.createElement('a');link.href='/batch/report.html';link.className='text-button';link.textContent='批量测试报告';document.querySelector('.top-actions').prepend(link)}
const entryParams=new URLSearchParams(location.search),preset=samples[entryParams.get('sample')]?entryParams.get('sample'):null,draftKey=preset?'htmlword-draft-v1:'+preset:'htmlword-draft-v1';
const settingsIDs=['page-size','orientation','margin','font-size','font','media','header-text','footer-text','page-numbers','remote'];
let renderedSignature='';
function documentSignature(){return $('html').value+'\u0000'+$('css').value+'\u0000'+JSON.stringify(options())}
function options(){return{page_size:$('page-size').value,landscape:$('orientation').value==='landscape',margin_mm:+$('margin').value,font:$('font').value||'Arial',font_size:+$('font-size').value,media:$('media').value,header:$('header-text').value,footer:$('footer-text').value,page_numbers:$('page-numbers').checked,allow_remote:$('remote').checked}}
function pageDimensions(o){let [w,h]=o.page_size==='A3'?[297,420]:o.page_size==='Letter'?[215.9,279.4]:[210,297];return o.landscape?[h,w]:[w,h]}
function safeSource(){
 const o=options(),doc=new DOMParser().parseFromString($('html').value,'text/html');doc.querySelectorAll('script,iframe,object,embed,base,meta[http-equiv]').forEach(x=>x.remove());for(const el of doc.querySelectorAll('*'))for(const a of [...el.attributes])if(a.name.startsWith('on')||a.name==='srcdoc')el.removeAttribute(a.name);
 const meta=doc.createElement('meta');meta.httpEquiv='Content-Security-Policy';const remote=o.allow_remote?' http: https:':'';meta.content=`default-src 'none'; img-src data: blob:${remote}; style-src 'unsafe-inline'${remote}; font-src data:${remote}; script-src 'none'; base-uri 'none'; form-action 'none'`;doc.head.prepend(meta);
 const style=doc.createElement('style');const [pw]=pageDimensions(o);style.textContent=`html{width:${(pw-o.margin_mm*2)*96/25.4}px}body{margin:0;font-family:${o.font};font-size:${o.font_size*96/72}px}img,svg{max-width:100%}*,*::before,*::after{animation:none!important;transition:none!important}`;meta.after(style);
 const extra=doc.createElement('style');extra.textContent=$('css').value;doc.head.append(extra);return '<!doctype html>'+doc.documentElement.outerHTML;
}
function resizePaper(){
 const o=options(),[pw,ph]=pageDimensions(o),px=pw*96/25.4,margin=o.margin_mm*96/25.4,stage=$('html-preview');if(stage.hidden)return;
 const scale=Math.min(1,(stage.clientWidth-(window.innerWidth<1050?28:48))/px);const paper=$('paper'),frame=$('html-frame');paper.style.width=(px-2*margin)+'px';paper.style.padding=margin+'px';paper.style.minHeight=(ph*96/25.4-2*margin)+'px';paper.style.transform=`scale(${Math.max(.15,scale)})`;
 try{const body=frame.contentDocument?.body,html=frame.contentDocument?.documentElement;if(body)frame.style.height=Math.max(ph*96/25.4-2*margin,body.scrollHeight,body.offsetHeight,html.scrollHeight)+'px'}catch{}
 $('paper-space').style.height=paper.offsetHeight*scale+'px';$('paper-space').style.width=px*scale+'px';$('paper-space').style.margin='0 auto';$('zoom').textContent=Math.round(scale*100)+'%';$('page-caption').textContent=`${o.page_size} · ${o.landscape?'横向':'纵向'} · ${o.margin_mm} mm`;
}
function updateSource(){
 const source=$(editor).value;$('source-size').textContent=source.length.toLocaleString()+' 字符';$('line-numbers').textContent=Array.from({length:source.split('\n').length},(_,i)=>i+1).join('\n');$('line-numbers').scrollTop=$(editor).scrollTop;$('css-dot').textContent=$('css').value.trim()?'•':'';
 clearTimeout(timer);timer=setTimeout(()=>{$('html-frame').srcdoc=safeSource();saveDraft()},300);dirty=true;if(latest&&documentSignature()!==renderedSignature){$('status-text').textContent='内容已修改，请重新转换以更新 Word。';$('word-ready').hidden=true;$('word-pages').hidden=true;$('pdf-frame').hidden=true;$('pdf-note').hidden=true;$('word-empty').hidden=false;$('word-empty').querySelector('h2').textContent='生成这个网页的 Word 预览';$('word-empty').querySelector('p').textContent='左侧内容已更新。点击生成预览后，右侧会显示这份 HTML 对应的实际 DOCX。';$('preview-convert').hidden=false}
}
function saveDraft(){try{localStorage.setItem(draftKey,JSON.stringify({html:$('html').value,css:$('css').value,sample:$('sample').value,options:options()}))}catch{}}
function setEditor(which){editor=which;for(const btn of document.querySelectorAll('[data-editor]')){const on=btn.dataset.editor===which;btn.classList.toggle('active',on);btn.setAttribute('aria-selected',on)}$('html').hidden=which!=='html';$('css').hidden=which!=='css';updateSource()}
function setPreview(which){preview=which;for(const btn of document.querySelectorAll('[data-preview]')){const on=btn.dataset.preview===which;btn.classList.toggle('active',on);btn.setAttribute('aria-selected',on)}const compare=which==='compare';document.querySelector('.workspace').classList.toggle('compare-mode',compare);$('preview-content').classList.toggle('comparing',compare);$('html-preview').hidden=which==='word';$('word-preview').hidden=which==='html';$('zoom').hidden=which==='word';$('preview-caption').textContent=compare?'左侧 HTML 原稿 / 右侧生成的 DOCX · 可分别滚动检查':which==='html'?'按纸张可用宽度显示':'DOCX 实际渲染预览';if(which!=='word')resizePaper()}
function updateSampleInfo(){const info=complexSampleInfo[$('sample').value];$('scenario-info').hidden=!info;if(info){$('scenario-title').textContent=info.title;$('scenario-features').textContent=info.features;$('scenario-note').textContent=info.note}}
function status(message,kind=''){document.querySelector('.result-bar').className='result-bar '+kind;$('status-text').textContent=message}
function toast(message){$('toast').textContent=message;$('toast').hidden=false;setTimeout(()=>$('toast').hidden=true,3500)}
function fromBase64(base64,type){const bytes=Uint8Array.from(atob(base64),c=>c.charCodeAt(0));return new Blob([bytes],{type})}
function download(){if(!latest)return;const a=document.createElement('a');a.href=docURL;a.download=latest.filename;document.body.append(a);a.click();a.remove()}
async function convert(downloadFile=true){
 if(controller)return;if(!$('html').value.trim()){toast('请先输入 HTML');return}
 const margin=+$('margin').value,size=+$('font-size').value;if(margin<0||margin>60||size<5||size>72){toast('请检查页边距和字号范围');return}
 controller=new AbortController();$('convert').disabled=true;$('render-preview').disabled=true;$('preview-convert').disabled=true;$('sample').disabled=true;$('html').readOnly=true;$('css').readOnly=true;for(const id of settingsIDs)$(id).disabled=true;$('cancel').hidden=false;$('result-actions').hidden=true;status('正在计算样式、生成原生 DOCX，并渲染文档预览…','busy');const revision=++renderID,requestSignature=documentSignature();
 try{
  const title=new DOMParser().parseFromString($('html').value,'text/html').title||'document';const response=await fetch('/api/convert',{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({html:$('html').value,css:$('css').value,filename:title,options:options(),preview:true}),signal:controller.signal});const data=await response.json();if(!response.ok)throw Error(data.error||'转换失败');if(revision!==renderID)return;
  latest=data;dirty=false;renderedSignature=requestSignature;if(docURL)URL.revokeObjectURL(docURL);docURL=URL.createObjectURL(fromBase64(data.docx_base64,'application/vnd.openxmlformats-officedocument.wordprocessingml.document'));
  if(pdfURL)URL.revokeObjectURL(pdfURL);for(const u of pageURLs)URL.revokeObjectURL(u);pageURLs=[];$('pdf-frame').hidden=true;$('pdf-note').hidden=true;$('word-pages').hidden=true;$('word-pages').replaceChildren();
  if(data.preview_pages?.length){data.preview_pages.forEach((b64,i)=>{const img=document.createElement('img');const u=URL.createObjectURL(fromBase64(b64,'image/png'));pageURLs.push(u);img.src=u;img.alt=`Word 文档第 ${i+1} 页`;img.loading='lazy';$('word-pages').append(img)});$('word-pages').hidden=false;$('pdf-note').hidden=false;$('word-empty').hidden=true;$('word-ready').hidden=false}else if(data.pdf_base64){pdfURL=URL.createObjectURL(fromBase64(data.pdf_base64,'application/pdf'));$('pdf-frame').src=pdfURL+'#view=FitH';$('pdf-frame').hidden=false;$('pdf-note').hidden=false;$('word-empty').hidden=true;$('word-ready').hidden=false}else{$('word-empty').hidden=false;$('word-empty').querySelector('h2').textContent='DOCX 已生成';$('word-empty').querySelector('p').textContent=data.preview_error||'当前服务未配置 LibreOffice。请下载 DOCX，在 Word 或 WPS 中查看并编辑。';$('preview-convert').hidden=true}
  $('result-details').hidden=false;$('stats').replaceChildren();const labels={paragraphs:'段落',tables:'表格',images:'图片',charts:'原生图表',checkboxes:'可选框',lists:'列表',characters:'字符'};for(const [key,label]of Object.entries(labels)){const item=document.createElement('span'),value=document.createElement('strong');value.textContent=(data.stats[key]||0).toLocaleString();item.append(value,label);$('stats').append(item)}
  const warnings=[...(data.warnings||[])];if(data.preview_error)warnings.push(data.preview_error);$('warnings').hidden=!warnings.length;$('warning-count').textContent=`${warnings.length} 项转换说明`;$('warning-list').replaceChildren();for(const warning of warnings){const li=document.createElement('li');li.textContent=warning;$('warning-list').append(li)}
  $('duration').textContent=`${(data.duration_ms/1000).toFixed(2)} s · ${(data.size/1024).toFixed(1)} KB`;$('result-actions').hidden=false;status('转换完成，正文可在 Word 中继续编辑和自动跨页。','success');if(downloadFile)download();if(!downloadFile)setPreview(preview==='compare'?'compare':'word');saveDraft();
 }catch(e){if(e.name==='AbortError')status('已取消转换。');else status(e.message,'error')}finally{controller=null;$('convert').disabled=false;$('render-preview').disabled=false;$('preview-convert').disabled=false;$('sample').disabled=false;$('html').readOnly=false;$('css').readOnly=false;for(const id of settingsIDs)$(id).disabled=false;$('cancel').hidden=true}
}
function loadSample(){if(dirty&&latest&&!Object.values(samples).includes($('html').value)&&!confirm('替换当前内容为示例？'))return;$('html').value=samples[$('sample').value];$('css').value='';updateSampleInfo();updateSource();setPreview(preview==='compare'?'compare':'html')}
async function importFiles(files){
 if(!files.length)return;try{
  const htmlFile=[...files].find(f=>/\.(html?|txt)$/i.test(f.name)),cssFiles=[...files].filter(f=>/\.css$/i.test(f.name)),images=[...files].filter(f=>f.type.startsWith('image/'));
  let source=htmlFile?await htmlFile.text():$('html').value;let css=cssFiles.length?(await Promise.all(cssFiles.map(f=>f.text()))).join('\n'):$('css').value;
  const map=new Map(await Promise.all(images.map(async f=>[f.name,await new Promise((resolve,reject)=>{const r=new FileReader();r.onload=()=>resolve(r.result);r.onerror=reject;r.readAsDataURL(f)})])));
  if(map.size||cssFiles.length){const doc=new DOMParser().parseFromString(source,'text/html');for(const img of doc.images){const name=(img.getAttribute('src')||'').split('/').pop();if(map.has(name))img.src=map.get(name)}for(const link of doc.querySelectorAll('link[rel="stylesheet"]'))if(cssFiles.some(f=>link.getAttribute('href')?.endsWith(f.name)))link.remove();source='<!doctype html>\n'+doc.documentElement.outerHTML;for(const [name,data]of map){const re=new RegExp(name.replace(/[.*+?^${}()|[\]\\]/g,'\\$&'),'g');css=css.replace(re,data)}}
  $('html').value=source;$('css').value=css;updateSource();setPreview('html');toast(`已导入 ${files.length} 个文件`);
 }catch(e){status('导入失败：'+e.message,'error')}
}
for(const btn of document.querySelectorAll('[data-editor]'))btn.onclick=()=>setEditor(btn.dataset.editor);for(const btn of document.querySelectorAll('[data-preview]'))btn.onclick=()=>setPreview(btn.dataset.preview);
for(const id of ['html','css']){$(id).addEventListener('input',updateSource);$(id).addEventListener('scroll',()=>{$('line-numbers').scrollTop=$(id).scrollTop});$(id).addEventListener('keydown',e=>{if(e.key==='Tab'){e.preventDefault();const a=e.target,start=a.selectionStart;a.setRangeText('  ',start,a.selectionEnd,'end');updateSource()}})}
$('convert').onclick=()=>convert(true);$('render-preview').onclick=()=>convert(false);$('preview-convert').onclick=()=>convert(false);$('download-again').onclick=download;$('cancel').onclick=()=>controller?.abort();$('sample').onchange=loadSample;$('reset').onclick=loadSample;
$('settings-toggle').onclick=()=>{const open=$('settings').hidden;$('settings').hidden=!open;$('settings-toggle').setAttribute('aria-expanded',open);resizePaper()};for(const id of settingsIDs)$(id).addEventListener('change',updateSource);
$('import').onclick=()=>$('files').click();$('files').onchange=e=>{importFiles(e.target.files);e.target.value=''};
document.addEventListener('dragover',e=>e.preventDefault());document.addEventListener('drop',e=>{e.preventDefault();if(e.dataTransfer.files.length)importFiles(e.dataTransfer.files)});
document.addEventListener('keydown',e=>{if((e.metaKey||e.ctrlKey)&&e.key==='Enter'){e.preventDefault();convert(true)}});window.addEventListener('resize',resizePaper);
$('html-frame').onload=()=>{resizePaper();for(const i of $('html-frame').contentDocument?.images||[])i.addEventListener('load',resizePaper);setTimeout(resizePaper,350)};
try{const draft=JSON.parse(localStorage.getItem(draftKey));if(draft?.html&&!preset){$('html').value=draft.html;$('css').value=draft.css||'';$('sample').value=draft.sample||'report';const o=draft.options||{};const mapping={'page-size':'page_size','margin':'margin_mm','font-size':'font_size','font':'font','media':'media','header-text':'header','footer-text':'footer'};for(const [id,key]of Object.entries(mapping))if(o[key]!==undefined)$(id).value=o[key];$('orientation').value=o.landscape?'landscape':'portrait';$('remote').checked=!!o.allow_remote;$('page-numbers').checked=!!o.page_numbers}else{$('sample').value=preset||'dashboard';$('html').value=samples[preset||'dashboard']}}catch{$('sample').value='dashboard';$('html').value=samples.dashboard}
updateSampleInfo();updateSource();dirty=false;if(entryParams.get('compare')==='1')setPreview('compare');if(preset&&entryParams.get('auto')==='1')convert(false);
fetch('/api/health').then(r=>r.json()).then(data=>{$('health').classList.toggle('bad',!data.ok);$('health').lastChild.textContent=data.ok?'Go 引擎已就绪':'未找到 Chrome';if(!data.pdf_preview)$('word-empty').querySelector('p').textContent='当前环境未配置 LibreOffice。转换后可以下载 DOCX，在 Word 或 WPS 中测试。'}).catch(()=>{$('health').classList.add('bad');$('health').lastChild.textContent='服务连接失败'});

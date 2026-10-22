// Executed in an isolated, script-free Chrome document. The result is a semantic
// document tree; browser line boxes are intentionally not converted to hard breaks.
(async () => {
  const warnings = new Set(),diagnostics=[];
  const policy=globalThis.__htmlwordPolicy||{};
  const maxNodes=policy.maxNodes||16000,maxDepth=policy.maxDepth||80,maxImages=policy.maxImages||100;
  if(document.querySelectorAll('*').length>maxNodes)throw Error('HTMLWORD_LIMIT: DOM node limit exceeded');
  if(document.querySelectorAll('img,svg').length>maxImages)throw Error('HTMLWORD_LIMIT: image count limit exceeded');
  function location(el){if(!el)return '';const parts=[];for(let n=el;n&&n!==document.body&&parts.length<12;n=n.parentElement){let i=1;for(let v=n.previousElementSibling;v;v=v.previousElementSibling)if(v.tagName===n.tagName)i++;parts.unshift(n.tagName.toLowerCase()+':nth-of-type('+i+')')}return 'body'+(parts.length?' > '+parts.join(' > '):'')}
  function warn(code,message,el){warnings.add(message);const path=location(el);if(diagnostics.length<200&&!diagnostics.some(d=>d.code===code&&d.location===path))diagnostics.push({code,message,location:path})}
  const stats = {paragraphs:0,tables:0,images:0,characters:0,lists:0,charts:0,checkboxes:0};
  let listID=0, visited=0;
  const footerBlocks=[];
  const skip = new Set(['SCRIPT','STYLE','LINK','META','HEAD','TITLE','NOSCRIPT','TEMPLATE','IFRAME','OBJECT','EMBED']);
  const num = x => Number.parseFloat(x)||0;
  const color = c => {
    if(!c || c==='transparent') return '';
    const v=c.match(/[\d.]+/g); if(!v || v.length<3) return '';
    const a=v.length>3?+v[3]:1;if(!a)return '';
    return v.slice(0,3).map(x=>Math.round(+x*a+255*(1-a)).toString(16).padStart(2,'0')).join('').toUpperCase();
  };
  const sides=['Top','Right','Bottom','Left'];
  function style(el) {
    const c=getComputedStyle(el),r=el.getBoundingClientRect();
    const families=c.fontFamily.split(',').map(f=>f.replace(/["']/g,'').trim());
    const eastAsia=families.find(f=>/PingFang|YaHei|SimSun|SimHei|Songti|Heiti|Hiragino|CJK|Noto Sans SC|Noto Serif SC|Source Han|微软雅黑|宋体|黑体|苹方/i.test(f))||families[0];
    return {font:families[0],east_asia:eastAsia,size:num(c.fontSize),color:color(c.color),background:color(c.backgroundColor),
      bold:num(c.fontWeight)>=600,italic:c.fontStyle!=='normal',underline:c.textDecorationLine.includes('underline'),strike:c.textDecorationLine.includes('line-through'),
      vertical:c.verticalAlign,align:c.textAlign,direction:c.direction,white_space:c.whiteSpace,line_height:num(c.lineHeight),letter_spacing:num(c.letterSpacing),text_indent:num(c.textIndent),
      margin:sides.map(s=>num(c['margin'+s])),padding:sides.map(s=>num(c['padding'+s])),borders:sides.map(s=>({width:num(c['border'+s+'Width']),style:c['border'+s+'Style'],color:color(c['border'+s+'Color'])})),
      width:r.width,height:r.height,no_wrap:c.whiteSpace==='nowrap',inline_box:c.display.startsWith('inline'),break_before:['always','page','left','right'].includes(c.breakBefore),break_after:['always','page','left','right'].includes(c.breakAfter),keep:c.breakInside==='avoid'||c.breakInside==='avoid-page',keep_next:['avoid','avoid-page'].includes(c.breakAfter)};
  }
  function hidden(el){const c=getComputedStyle(el);return skip.has(el.tagName)||el.hidden||c.display==='none'||c.visibility==='hidden'||c.visibility==='collapse'}
  function note(el){
    const c=getComputedStyle(el);
    if(c.position==='absolute'||c.position==='fixed')warn('POSITION_TO_FLOW','绝对定位与固定定位内容按阅读顺序转换为流式内容，坐标不锁定。',el);
    if(c.transform!=='none')warn('TRANSFORM_UNSUPPORTED','CSS transform 的旋转与变换未映射，保留原生可编辑内容。',el);
    if(c.boxShadow!=='none'||c.filter!=='none'||c.clipPath!=='none'||num(c.borderRadius)>0)warn('DECORATION_APPROXIMATED','阴影、滤镜、裁剪和圆角未完全映射。',el);
    if(c.backgroundImage!=='none')warn('BACKGROUND_UNSUPPORTED','CSS 背景图片或渐变未映射；请使用 img 元素表达必要图片。',el);
    if(c.opacity!=='1')warn('OPACITY_UNSUPPORTED','整体透明度未完整映射。',el);
    if(c.writingMode!=='horizontal-tb')warn('WRITING_MODE_UNSUPPORTED','竖排文字按横排内容输出。',el);
  }
  async function raster(el){
    try{
      let src=el.currentSrc||el.src;
      if(el.tagName.toLowerCase()==='svg')src='data:image/svg+xml;base64,'+btoa(unescape(encodeURIComponent(new XMLSerializer().serializeToString(el))));
      const im=new Image();im.crossOrigin='anonymous';im.src=src;
      await Promise.race([im.decode(),new Promise((_,rej)=>setTimeout(()=>rej(Error('timeout')),4000))]);
      const r=el.getBoundingClientRect(),w=r.width||im.naturalWidth||300,h=r.height||im.naturalHeight||150;
      const scale=Math.min(2,4096/Math.max(w,h));const canvas=document.createElement('canvas');canvas.width=Math.max(1,Math.round(w*scale));canvas.height=Math.max(1,Math.round(h*scale));
      canvas.getContext('2d').drawImage(im,0,0,canvas.width,canvas.height);
      return canvas.toDataURL('image/png');
    }catch{return el.currentSrc||el.src||''}
  }
  const imageData=new WeakMap();
  await Promise.all([...document.querySelectorAll('img,svg')].slice(0,maxImages).map(async el=>imageData.set(el,await raster(el))));
  function picture(el){
    stats.images++;const r=el.getBoundingClientRect();
    if(el.tagName.toLowerCase()==='svg')warn('SVG_RASTERIZED','SVG 作为静态图片导出，SVG 内部文字不属于原生可编辑正文。',el);
    const s=style(el);return {source:imageData.get(el)||el.currentSrc||el.src||'',alt:el.getAttribute('alt')||el.getAttribute('aria-label')||'图片',width:r.width||num(el.getAttribute('width'))||300,height:r.height||num(el.getAttribute('height'))||150,outline:s.borders[0],radius:num(getComputedStyle(el).borderRadius)};
  }
  const chartCache=new WeakMap();
  // Native Office charts require real data. Never invent values from a bitmap.
  // JSON works on visual elements; a typed table annotation reads its own cells.
  function chartData(el){
    if(chartCache.has(el))return chartCache.get(el);
    const raw=el.getAttribute('data-office-chart');if(!raw)return null;
    let chart=null;
    try{
      if(raw.trim().startsWith('{'))chart=JSON.parse(raw);
      else if(el.tagName==='TABLE'){
        const rows=[...el.rows].filter(r=>!hidden(r));
        if(rows.length<2||[...el.querySelectorAll('td,th')].some(c=>c.colSpan>1||c.rowSpan>1))throw Error('表格需要单层表头及至少一行数据');
        const head=[...rows[0].cells].map(c=>c.textContent.trim()),data=rows.slice(1).map(r=>[...r.cells].map(c=>c.textContent.trim()));
        if(head.length<2||data.some(r=>r.length!==head.length))throw Error('数据列不完整');
        const value=s=>{s=s.replace(/,/g,'').replace(/%$/,'').trim();if(!/^[+-]?(?:\d+\.?\d*|\.\d+)(?:e[+-]?\d+)?$/i.test(s))throw Error('非数值单元格');return Number(s)};
        chart={type:raw,categories:data.map(r=>r[0]),series:head.slice(1).map((name,i)=>({name,values:data.map(r=>value(r[i+1]))})),show_axes:true,legend:head.length>2};
      }else throw Error('需要图表 JSON 数据');
      if(!['line','bar','column','pie','doughnut'].includes(chart.type)||!Array.isArray(chart.categories)||!chart.categories.length||chart.categories.length>512||!Array.isArray(chart.series)||!chart.series.length||chart.series.length>16)throw Error('图表类型或数据规模无效');
      chart.categories=chart.categories.map(String);
      if(chart.series.some(s=>!Array.isArray(s.values)||s.values.length!==chart.categories.length||s.values.some(v=>typeof v!=='number'||!Number.isFinite(v)||Math.abs(v)>1e100)))throw Error('系列数值与类别不一致');
      if(['pie','doughnut'].includes(chart.type)&&(chart.series.length!==1||chart.series[0].values.some(v=>v<0)||chart.series[0].values.reduce((a,b)=>a+b,0)<=0))throw Error('饼图数据无效');
      if(chart.max!==undefined&&(!Number.isFinite(chart.max)||chart.max<0))throw Error('坐标上限无效');
      const rect=el.getBoundingClientRect();chart={...chart,width:rect.width||300,height:el.tagName==='TABLE'?num(el.getAttribute('data-chart-height'))||240:rect.height||150};
      chart.series=chart.series.map((s,i)=>({...s,name:String(s.name||'Series '+(i+1))}));
    }catch(e){warn('CHART_DATA_INVALID','原生图表数据无效，保留原始内容：'+e.message,el);chart=null}
    chartCache.set(el,chart);return chart;
  }
  function chartRun(el,chart){stats.charts++;return {chart,style:style(el)}}
  function pseudo(el,which,runs){const c=getComputedStyle(el,which),s=c.content;if(s&&s!=='none'&&s!=='normal'&&/^(["']).*\1$/.test(s)){runs.push({text:s.slice(1,-1),style:{...style(el),color:color(c.color),size:num(c.fontSize)}})}}
  function inlines(node,runs,link=''){
    if(node.nodeType===Node.TEXT_NODE){
      const s=style(node.parentElement);let text=node.nodeValue;
      if(!['pre','pre-wrap','break-spaces'].includes(s.white_space))text=text.replace(/[\t\n\r ]+/g,' ');
      const transform=getComputedStyle(node.parentElement).textTransform;
      if(transform==='uppercase')text=text.toUpperCase();if(transform==='lowercase')text=text.toLowerCase();if(transform==='capitalize')text=text.replace(/\b\p{L}/gu,c=>c.toUpperCase());
      if(s.inline_box&&node.parentElement.childNodes.length===1&&s.borders.every(b=>b.width>0))text='\u00a0'.repeat(Math.round(s.padding[3]/Math.max(1,s.size*.28)))+text.replace(/ /g,'\u00a0')+'\u00a0'.repeat(Math.round(s.padding[1]/Math.max(1,s.size*.28)));
      if(text)runs.push({text,style:s,link});return;
    }
    if(node.nodeType!==Node.ELEMENT_NODE||hidden(node))return;
    note(node);const tag=node.tagName.toLowerCase();
    const chart=chartData(node);if(chart&&tag!=='table'){runs.push(chartRun(node,chart));return}
    if(tag==='br'){runs.push({break:true,style:style(node)});return}
    if(tag==='img'||tag==='svg'){runs.push({image:picture(node),style:style(node)});return}
    if(tag==='canvas'){warn('CANVAS_UNSUPPORTED','Canvas 位图内容无法恢复为可编辑文字，未作为正文转换。',el);return}
    if(tag==='input'||tag==='textarea'||tag==='select'){
      if(tag==='input'&&node.type==='checkbox'){stats.checkboxes++;runs.push({checkbox:{checked:node.checked,disabled:node.disabled,name:node.getAttribute('aria-label')||node.name||node.labels?.[0]?.textContent.trim()||node.parentElement.textContent.trim()},style:style(node)});return}
      if(tag==='input'&&node.type==='radio')warn('RADIO_AS_TEXT','单选按钮保留当前状态文字，暂未映射互斥的 Office 控件组。',el);
      const text=tag==='input'&&node.type==='radio'?(node.checked?'◉':'○'):(node.value||node.getAttribute('placeholder')||'');runs.push({text,style:style(node)});return;
    }
    let href=link;if(tag==='a')href=node.getAttribute('href')||link;
    const c=getComputedStyle(node);pseudo(node,'::before',runs);
    for(const child of node.childNodes)inlines(child,runs,href);
    pseudo(node,'::after',runs);
    if(c.display==='inline-block'&&num(c.marginRight)>0)runs.push({text:' ',style:{...style(node),size:4/3,letter_spacing:Math.max(0,num(c.marginRight)-.37),background:'',borders:[]}});
  }
  function cleanRuns(runs){
    if(!runs.some(r=>['pre','pre-wrap','break-spaces'].includes(r.style.white_space))){
      let prevSpace=true;
      for(const r of runs){if(r.text!==undefined){if(prevSpace)r.text=r.text.replace(/^ /,'');prevSpace=r.text.endsWith(' ')}else prevSpace=!r.checkbox}
      const last=[...runs].reverse().find(r=>r.text);if(last)last.text=last.text.replace(/ $/,'');
    }
    const first=runs.findIndex(r=>r.text||r.checkbox||r.image||r.chart);
    if((!policy.explicitSemantics||policy.symbolCheckboxes)&&first>=0&&runs[first].text&&/^[□☐☑☒](?=\s)/u.test(runs[first].text)){
      const r=runs[first],checked=['☑','☒'].includes(r.text[0]);r.text=r.text.slice(1);
      runs.splice(first,0,{checkbox:{checked,name:r.text.trim().slice(0,100)},style:r.style});stats.checkboxes++;
    }
    return runs.filter(r=>r.text||r.break||r.image||r.chart||r.checkbox);
  }
  function paragraph(el,nodes,ownStyle=style(el)){
    const runs=[];for(const n of nodes)inlines(n,runs);const result=cleanRuns(runs);
    stats.paragraphs++;stats.characters+=result.reduce((n,r)=>n+(r.text?.length||0),0);
    const inlineBoxes=[...el.querySelectorAll('*')].filter(e=>getComputedStyle(e).display==='inline-block');if(inlineBoxes.length)ownStyle={...ownStyle,line_height:Math.max(ownStyle.line_height,...inlineBoxes.map(e=>{const s=style(e);return s.height+s.margin[0]+s.margin[2]}))};
    const numeric=/^\d{1,3}$/.test(el.textContent.trim());if(numeric&&ownStyle.width<64){ownStyle={...ownStyle,no_wrap:true};for(const r of result)r.style={...r.style,no_wrap:true,width:ownStyle.width}}
    const tag=el.tagName.toLowerCase();return {kind:'paragraph',tag,id:el.id||'',style:ownStyle,level:/^h[1-6]$/.test(tag)?+tag[1]:0,runs:result};
  }
  function isBlock(el){const c=getComputedStyle(el);return !['inline','inline-block','inline-flex','inline-grid','contents'].includes(c.display)||['table','ul','ol','pre','hr'].includes(el.tagName.toLowerCase())}
  function children(el,depth){
    const blocks=[];let pending=[],previous=null;
    function flush(){if(pending.some(n=>n.nodeType===1||n.textContent.trim()))blocks.push(paragraph(el,pending,contentStyle(style(el))));pending=[]}
    for(const node of el.childNodes){
      if(node.nodeType===1&&hidden(node))continue;
      if(node.nodeType===1&&node.tagName==='FOOTER'&&(!policy.explicitSemantics||policy.footerMode!=='preserve')&&el===document.body&&node.textContent.length<300&&style(node).size<=14&&style(node).height<100&&!node.querySelector('img,svg,table,input,textarea,select,h1,h2,h3')){
        flush();const s=style(node),candidate=layout(node,depth+1,contentStyle(s));
        const footer=candidate?.kind==='paragraph'?candidate:paragraph(node,[...node.childNodes],contentStyle(s));
        footer.style={...footer.style,margin:[0,0,0,0],padding:[4,0,4,0],borders:s.borders};
        footerBlocks.push(footer);previous=null;warn('FOOTER_PROMOTED','短文档页尾已转换为原生 Word 页脚，避免正文页尾标识单独占页。',node);continue;
      }
      if(node.nodeType===1&&isBlock(node)){
        flush();const next=walk(node,depth+1),display=getComputedStyle(el).display;
        if(previous&&next.length&&!/flex|grid/.test(display)&&getComputedStyle(node).position==='static'){
          const tail=blocks[blocks.length-1],gap=node.getBoundingClientRect().top-previous.getBoundingClientRect().bottom;
          if(tail&&gap>=-.5&&gap<tail.style.margin[2]+next[0].style.margin[0]){
            tail.style.margin[2]=Math.min(tail.style.margin[2],Math.max(0,gap));
            next[0].style.margin[0]=Math.max(0,gap-tail.style.margin[2]);
          }
        }
        blocks.push(...next);previous=node;
      }else {pending.push(node);if(node.textContent.trim())previous=null}
    }
    flush();return blocks;
  }
  // The cell owns the item's box decoration. Its content must still go through
  // walk(), including that item's own flex/grid/list/image semantics.
  function contentStyle(s){return {...s,margin:[0,0,0,0],padding:[0,0,0,0],borders:[],background:'',width:Math.max(0,s.width-s.padding[1]-s.padding[3]-(s.borders[1]?.width||0)-(s.borders[3]?.width||0))}}
  function cellBlocks(el,depth){return walk(el,depth+1,true)}
  function boxed(s){return s.background||s.borders.some(b=>b.width>0&&b.style!=='none')||s.padding.some(v=>v>0)}
  function container(el,s,blocks){stats.tables++;return {kind:'table',tag:'container',id:el.id||'',style:s,rows:[{cells:[{style:{...s,margin:[0,0,0,0]},span:1,row_span:1,blocks,v_align:'top'}]}]}}
  function table(el,depth){
    stats.tables++;const rows=[];
    if(el.caption)rows.push({cells:[{blocks:walk(el.caption,depth+1),style:style(el.caption),span:1,row_span:1}],header:false});
    for(const tr of el.rows){if(hidden(tr))continue;
      rows.push({header:tr.parentElement.tagName==='THEAD',keep:style(tr).keep,cells:[...tr.cells].filter(c=>!hidden(c)).map(td=>({blocks:cellBlocks(td,depth+1),style:style(td),span:Math.min(128,td.colSpan||1),row_span:td.rowSpan===0?Math.max(1,[...el.rows].indexOf(tr.parentElement?.lastElementChild)-[...el.rows].indexOf(tr)+1):td.rowSpan||1,v_align:getComputedStyle(td).verticalAlign}))});
    }
    return {kind:'table',tag:'table',style:style(el),rows};
  }
  function layout(el,depth,s){
    const c=getComputedStyle(el),els=[...el.children].filter(e=>!hidden(e));
    if(els.length<2||[...el.childNodes].some(n=>n.nodeType===3&&n.textContent.trim()))return null;
    if(c.display.includes('flex')&&c.flexDirection.startsWith('column'))return null;
    // A flex row made entirely of single-line text belongs in one paragraph.
    // Intrinsic-width table cells are too fragile when Word rounds font sizes
    // to half points. Native tabs retain date alignment without splitting text.
    const rowLineHeight=Math.max(...els.map(e=>num(getComputedStyle(e).lineHeight)||num(getComputedStyle(e).fontSize)*1.3));
    if(c.display.includes('flex')&&c.flexWrap==='nowrap'&&els.every(e=>!e.querySelector('img,svg,input,table,ul,ol')&&[...e.children].every(ch=>!isBlock(ch))&&e.getBoundingClientRect().height<=rowLineHeight+1)){
      const runs=[],tabs=[],rect=el.getBoundingClientRect(),own=style(el),left=rect.left+own.padding[3]+own.borders[3].width;
      let level=0;
      for(let i=0;i<els.length;i++){
        const e=els[i];
        if(i){
          if(c.justifyContent==='space-between'){const last=i===els.length-1;tabs.push({position:last?s.width:e.getBoundingClientRect().left-left,align:last?'right':'left'});runs.push({text:'\t',style:s})}
          else{const gap=Math.max(0,e.getBoundingClientRect().left-els[i-1].getBoundingClientRect().right);if(gap>0)runs.push({text:' ',style:{...s,size:4/3,letter_spacing:Math.max(0,gap-.37)}})}
        }
        inlines(e,runs);if(/^H[1-6]$/.test(e.tagName))level=Math.max(level,+e.tagName[1]);
      }
      const result=cleanRuns(runs);stats.paragraphs++;stats.characters+=result.reduce((n,r)=>n+(r.text?.length||0),0);
      return {kind:'paragraph',tag:'flex-line',id:el.id||'',style:{...s,line_height:Math.max(...els.map(e=>num(getComputedStyle(e).lineHeight))),no_wrap:false},level,tabs,runs:result};
    }
    const sorted=els.map((e,i)=>{let r=e.getBoundingClientRect();if(!e.children.length&&/^\d{1,3}$/.test(e.textContent.trim())){const range=document.createRange();range.selectNodeContents(e);const width=Math.max(r.width,range.getBoundingClientRect().width+1.5);r={left:r.left,right:r.left+width,top:r.top,bottom:r.bottom,width,height:r.height}}return {e,i,r}}).filter(v=>v.r.width>0).sort((a,b)=>Math.abs(a.r.top-b.r.top)<1?a.r.left-b.r.left:a.r.top-b.r.top);
    if(c.display.includes('flex')&&c.flexWrap==='nowrap'&&c.justifyContent==='space-between'){
      const physical=[...sorted].sort((a,b)=>a.r.left-b.r.left);
      for(let i=0;i<physical.length-1;i++){const v=physical[i],end=physical[i+1].r.left-(num(c.columnGap)||0);if(!boxed(style(v.e))&&v.e.children.length&&end>v.r.right+1)v.r={left:v.r.left,right:end,top:v.r.top,bottom:v.r.bottom,width:end-v.r.left,height:v.r.height}}
    }
    const groups=[];
    if(c.display.includes('flex')&&c.flexWrap==='nowrap')groups.push({items:sorted});
    else for(const item of sorted){
      // Center/baseline aligned items have different top coordinates. Group by
      // overlapping vertical intervals and disjoint horizontal intervals.
      let row=groups.find(g=>g.items.every(v=>Math.min(v.r.right,item.r.right)-Math.max(v.r.left,item.r.left)<1)&&g.items.some(v=>Math.min(v.r.bottom,item.r.bottom)-Math.max(v.r.top,item.r.top)>1));
      if(!row)groups.push(row={items:[]});row.items.push(item);
    }
    if(groups.every(r=>r.items.length===1))return null;
    if(els.some(e=>getComputedStyle(e).gridRowEnd.includes('span')||getComputedStyle(e).gridColumnEnd.includes('span')))warn('GRID_SPAN_APPROXIMATED','跨行跨列 Grid 采用近似的流式表格布局。',el);
    warn('LAYOUT_AS_TABLE','Flex/Grid 横向布局已转换为可编辑表格，复杂布局的间距可能有差异。',el);stats.tables++;
    const own=style(el),rect=el.getBoundingClientRect(),left=rect.left+own.padding[3]+own.borders[3].width,width=s.width;
    const offset=x=>Math.max(0,Math.min(width,x-left));
    const edges=[0,width,...sorted.flatMap(v=>[offset(v.r.left),offset(v.r.right)])].sort((a,b)=>a-b).filter((v,i,a)=>!i||v-a[i-1]>.5);
    if(edges.length>129){warn('DENSE_GRID_FLATTENED','布局网格过密，已展开为连续内容。',el);return null}
    const index=x=>{const v=offset(x);let best=0;for(let i=1;i<edges.length;i++)if(Math.abs(edges[i]-v)<Math.abs(edges[best]-v))best=i;return best};
    const grid=edges.slice(1).map((v,i)=>v-edges[i]),rows=[];
    const spacer=span=>({style:{width:grid.slice(0,span).reduce((a,b)=>a+b,0),padding:[0,0,0,0],margin:[0,0,0,0],borders:[]},span,row_span:1,blocks:[]});
    groups.sort((a,b)=>Math.min(...a.items.map(v=>v.r.top))-Math.min(...b.items.map(v=>v.r.top)));
    for(let gi=0;gi<groups.length;gi++){
      const g=groups[gi],cells=[];let pos=0;
      for(const {e,r} of g.items.sort((a,b)=>a.r.left-b.r.left)){
        const start=index(r.left),end=index(r.right);if(end<=start)continue;
        if(start>pos)cells.push(spacer(start-pos));
        const es=style(e),align=getComputedStyle(e).alignSelf==='auto'?c.alignItems:getComputedStyle(e).alignSelf;
        const image=['IMG','svg'].includes(e.tagName);
        const separate=!image&&['center','end','flex-end','baseline'].includes(align)&&boxed(es);
        const itemStyle=separate?{...contentStyle(es),width:r.width}:{...es,margin:[0,0,0,0],borders:image?[]:es.borders};
        const blocks=cellBlocks(e,depth+1);
        cells.push({style:{...itemStyle,no_wrap:es.no_wrap||(!e.children.length&&/^\d{1,3}$/.test(e.textContent.trim()))},span:end-start,row_span:1,blocks:separate?[container(e,{...es,margin:[0,0,0,0]},blocks)]:blocks,v_align:align==='center'?'middle':['end','flex-end','baseline'].includes(align)?'bottom':'top'});pos=end;
      }
      if(pos<grid.length)cells.push(spacer(grid.length-pos));
      // Layout rows can be nested in a long outer row. Keep their paragraph
      // chain together instead of cantSplit, which can overlap nested rows in
      // document renderers. Native HTML tr constraints are handled separately.
      const groupHeight=Math.max(...g.items.map(v=>v.r.height));
      // Small decorated cards are atomic rows. Tall document columns remain
      // splittable and are segmented by semantic block boundaries in the writer.
      const headingRow=g.items.some(v=>v.e.querySelector('h1,h2,h3,h4,h5,h6'));
      rows.push({cells,keep:groupHeight<640&&(headingRow||g.items.every(v=>boxed(style(v.e))||['IMG','svg'].includes(v.e.tagName)))});
      if(gi<groups.length-1){const gap=Math.min(...groups[gi+1].items.map(v=>v.r.top))-Math.max(...g.items.map(v=>v.r.bottom));if(gap>.5)rows.push({height:gap,exact:true,cells:[spacer(grid.length)]})}
    }
    return {kind:'table',tag:'layout',id:el.id||'',style:s,grid,rows};
  }
  function decoration(el,s){
    if(el.textContent.trim()||el.querySelector('img,svg,input,textarea,select,canvas')||[...el.children].some(e=>e.childElementCount>0)||s.height<=0||s.height>128)return null;
    const rect=el.getBoundingClientRect(),children=[...el.children].filter(e=>!hidden(e)&&style(e).background).map(e=>({s:style(e),r:e.getBoundingClientRect()}));
    if(!s.background&&!children.length)return null;
    const edges=[0,s.width,...children.flatMap(v=>[Math.max(0,v.r.left-rect.left),Math.min(s.width,v.r.right-rect.left)])].sort((a,b)=>a-b).filter((v,i,a)=>!i||v-a[i-1]>.1);
    const grid=edges.slice(1).map((v,i)=>v-edges[i]);
    const cells=grid.map((width,i)=>{const mid=rect.left+edges[i]+width/2,fill=children.find(v=>v.r.left<=mid&&v.r.right>=mid);return {style:{...contentStyle(s),width,background:fill?.s.background||s.background},span:1,row_span:1,blocks:[]}});
    stats.tables++;return {kind:'table',tag:'decoration',id:el.id||'',style:s,grid,rows:[{height:s.height,exact:true,cells}]};
  }
  function walk(el,depth=0,insideCell=false){
    if(depth>maxDepth||++visited>maxNodes)throw Error('HTMLWORD_LIMIT: document depth or block count exceeded');
    if(hidden(el))return [];note(el);const tag=el.tagName.toLowerCase(),c=getComputedStyle(el),s=insideCell?contentStyle(style(el)):style(el);
    if(tag==='table'){const result=[table(el,depth)],chart=chartData(el);if(chart){stats.paragraphs++;result.push({kind:'paragraph',style:contentStyle(s),runs:[chartRun(el,chart)]})}return result}
    const chart=chartData(el);if(chart){stats.paragraphs++;return [{kind:'paragraph',id:el.id||'',style:s,runs:[chartRun(el,chart)]}]}
    if(['img','svg','input','textarea','select'].includes(tag))return [paragraph(el,[el],s)];
    if(tag==='ul'||tag==='ol'){
      if(c.listStyleType==='none')return children(el,depth+1);
      const id=++listID;stats.lists++;const out=[];let first=true;
      for(const li of el.children){if(li.tagName!=='LI'||hidden(li))continue;const parts=children(li,depth+1);let p=parts.find(b=>b.kind==='paragraph');
        if(!p){p=paragraph(li,[]);parts.unshift(p)}
        let level=0;for(let ancestor=el.parentElement;ancestor;ancestor=ancestor.parentElement)if(ancestor.tagName==='UL'||ancestor.tagName==='OL')level++;
        Object.assign(p,{list_id:id,list_level:Math.min(8,level),ordered:tag==='ol',start:first?num(el.getAttribute('start'))||1:0,list_format:c.listStyleType});
        let indent=0;for(let a=el;a;a=a.parentElement)if(a.tagName==='UL'||a.tagName==='OL'){const as=style(a);indent+=as.padding[3]+as.margin[3]}
        p.style.margin[3]+=Math.max(0,indent);p.style.text_indent=-Math.min(Math.max(0,indent),p.style.size*.9);
        if(parts.length){parts[0].style.margin[0]=Math.max(parts[0].style.margin[0],style(li).margin[0]);parts[parts.length-1].style.margin[2]=Math.max(parts[parts.length-1].style.margin[2],style(li).margin[2])}
        first=false;out.push(...parts);
      }return out;
    }
    if(tag==='hr'){return [{kind:'rule',style:s}]}
    if(tag==='canvas'){warn('CANVAS_UNSUPPORTED','Canvas 内容未转换为可编辑正文。',el);return []}
    if(num(c.columnCount)>1){const b=children(el,depth+1);warn('COLUMNS_NATIVE','CSS 多栏转换为 Word 连续分节分栏；实际分页由 Word 计算。',el);return [{kind:'columns',style:s,columns:Math.min(6,num(c.columnCount)),gap:num(c.columnGap),children:b}]}
    if(c.display.includes('flex')||c.display.includes('grid')){const b=layout(el,depth,contentStyle(s));if(b){if(boxed(s))return [container(el,s,[b])];b.style.margin=s.margin;return [b]}}
    const paint=decoration(el,s);if(paint)return [paint];
    const hasBlocks=[...el.children].some(e=>!hidden(e)&&isBlock(e));
    if(!hasBlocks)return [paragraph(el,[...el.childNodes],s)];
    const blocks=children(el,depth);
    if(tag==='figure')for(let i=0;i<blocks.length-1;i++)blocks[i].style.keep_next=true;
    if(boxed(s))return [container(el,s,blocks)];
    if(blocks.length){if(el.id)blocks[0].id=el.id;blocks[0].style.margin[0]=Math.max(blocks[0].style.margin[0],s.margin[0]);blocks[blocks.length-1].style.margin[2]=Math.max(blocks[blocks.length-1].style.margin[2],s.margin[2]);if(s.break_before)blocks[0].style.break_before=true;if(s.break_after)blocks[blocks.length-1].style.break_after=true;if(s.keep_next)blocks[blocks.length-1].style.keep_next=true}
    return blocks;
  }
  const blocks=children(document.body,0);
  if(document.body.getBoundingClientRect().width>innerWidth+2)warn('CONTENT_SCALED','HTML 内容宽于纸张可用区域，表格和图片会按可用宽度缩放。',document.body);
  for(const img of document.images)if(!img.complete||!img.naturalWidth)warn('IMAGE_UNAVAILABLE','部分图片未成功加载；转换结果将使用替代文字或后端下载。',img);
  for(const link of document.querySelectorAll('link[rel="stylesheet"]'))if(!link.sheet)warn('STYLESHEET_UNAVAILABLE','外部样式表未加载，请导入对应 CSS 文件或启用公网资源。',link);
  if(document.querySelector('script'))warn('ACTIVE_CONTENT_REMOVED','脚本不执行，导出输入 HTML 的静态内容。',document.body);
  return {title:document.title||'HTML document',blocks,footer:footerBlocks,warnings:[...warnings],diagnostics,stats};
})()

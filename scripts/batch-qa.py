"""Convert the twelve-case suite and collect evidence, without template fixes.

Run with the bundled Python runtime. Rendering requires --renderer and
--fontconfig. Results describe package/content checks, not pixel parity.
"""
from pathlib import Path
from collections import Counter
from concurrent.futures import ThreadPoolExecutor, as_completed
from html.parser import HTMLParser
from xml.etree import ElementTree as E
import argparse, base64, hashlib, io, json, os, re, shutil, subprocess, sys, time, urllib.request, zipfile

P=argparse.ArgumentParser()
P.add_argument('--base',default='http://127.0.0.1:8787')
P.add_argument('--out',required=True)
P.add_argument('--renderer',required=True)
P.add_argument('--fontconfig',required=True)
P.add_argument('--batch-dir')
P.add_argument('--only',help='Comma-separated sample ids to refresh; merge unchanged prior results')
A=P.parse_args()
HTTP=urllib.request.build_opener(urllib.request.ProxyHandler({}))
ROOT=Path(__file__).resolve().parents[1]
OUT=Path(A.out).resolve();OUT.mkdir(parents=True,exist_ok=True)
BATCH=Path(A.batch_dir).resolve() if A.batch_dir else ROOT/'tools/validation/artifacts/batch'
MANIFEST=json.loads((BATCH/'manifest.json').read_text())
SELECTED=set(A.only.split(',')) if A.only else set(MANIFEST)
if not SELECTED.issubset(MANIFEST):P.error('unknown sample id')
PRIOR=json.loads((BATCH/'results.json').read_text()) if A.only else None
PRIOR_ITEMS={i['id']:i for i in PRIOR['samples']} if PRIOR else {}
NS={'w':'http://schemas.openxmlformats.org/wordprocessingml/2006/main','c':'http://schemas.openxmlformats.org/drawingml/2006/chart','pic':'http://schemas.openxmlformats.org/drawingml/2006/picture','w14':'http://schemas.microsoft.com/office/word/2010/wordml'}
W='{'+NS['w']+'}'

def normalized(s):
    return ''.join(c for c in s.casefold().translate(str.maketrans({'□':'☐','☒':'☑'})) if not c.isspace())

class Source(HTMLParser):
    def __init__(self):
        super().__init__();self.inbody=False;self.text=[];self.tags=Counter();self.images=0;self.links=0;self.values=[];self.colspan=False;self.rowspan=False
        self.charts=[];self.annotated_images=0;self.checkboxes=0;self.checked=0
    def handle_starttag(self,tag,attrs):
        a=dict(attrs)
        if tag=='body':self.inbody=True
        if not self.inbody:return
        self.tags[tag]+=1
        if tag=='img':self.images+=1
        if a.get('data-office-chart'):
            self.charts.append(json.loads(a['data-office-chart']))
            if tag=='img':self.annotated_images+=1
        if tag=='input' and a.get('type')=='checkbox':
            self.checkboxes+=1;self.checked+=int('checked' in a)
        if tag=='a' and a.get('href'):self.links+=1
        if int(a.get('colspan','1'))>1:self.colspan=True
        if int(a.get('rowspan','1'))>1:self.rowspan=True
        if tag=='input' and a.get('value'):
            self.values.append(a['value']);self.text.append(a['value'])
    def handle_endtag(self,tag):
        if tag=='body':self.inbody=False
    def handle_data(self,s):
        if self.inbody:
            self.text.append(s)
            if re.match(r'^\s*[□☐☑☒]\s',s):
                self.checkboxes+=1;self.checked+=int(s.lstrip()[0] in '☑☒')

def package_checks(data,source,required):
    with zipfile.ZipFile(io.BytesIO(data)) as z:
        for name in z.namelist():
            if name.endswith(('.xml','.rels')):E.fromstring(z.read(name))
        raw=z.read('word/document.xml');tree=E.fromstring(raw)
        text=''.join(e.text or '' for e in tree.findall('.//w:t',NS))
        if 'word/footer1.xml' in z.namelist():
            footer=E.fromstring(z.read('word/footer1.xml'))
            text+=''.join(e.text or '' for e in footer.findall('.//w:t',NS))
        expected=Counter(normalized(''.join(source.text)));actual=Counter(normalized(text))
        missing=expected-actual
        cursor=0;actual_sequence=normalized(text)
        for character in normalized(''.join(source.text)):
            cursor=actual_sequence.find(character,cursor)
            if cursor<0:break
            cursor+=1
        errors=[]
        if missing:errors.append('source body characters missing: '+repr(dict(missing)))
        if cursor<0:errors.append('source body character order changed')
        for token in required:
            if normalized(token) not in normalized(text):errors.append('missing key content: '+token)
        if any(t in raw for t in [b'altChunk',b'txbxContent',b'wp:anchor',b'documentProtection']):errors.append('fixed or protected body structure')
        images=len(tree.findall('.//w:drawing',NS));links=len(tree.findall('.//w:hyperlink',NS))
        pictures=len(tree.findall('.//pic:pic',NS));charts=len(tree.findall('.//c:chart',NS));checkboxes=tree.findall('.//w14:checkbox',NS)
        if pictures!=source.images-source.annotated_images:errors.append(f'image count mismatch: {pictures}/{source.images-source.annotated_images}')
        if charts!=len(source.charts):errors.append(f'native chart count mismatch: {charts}/{len(source.charts)}')
        if len(checkboxes)!=source.checkboxes:errors.append('interactive checkbox count mismatch')
        checked=sum(c.find('w14:checked',NS).get('{'+NS['w14']+'}val')=='1' for c in checkboxes)
        if checked!=source.checked:errors.append('checkbox initial states changed')
        for i,spec in enumerate(source.charts,1):
            chart=E.fromstring(z.read(f'word/charts/chart{i}.xml'))
            series=chart.findall('.//c:ser',NS)
            if len(series)!=len(spec['series']):errors.append('chart series count changed')
            for actual_series,expected_series in zip(series,spec['series']):
                actual_values=[float(e.text) for e in actual_series.findall('c:val/c:numRef/c:numCache/c:pt/c:v',NS)]
                if actual_values!=expected_series['values']:errors.append('chart values changed')
                categories=[e.text for e in actual_series.findall('c:cat/c:strRef/c:strCache/c:pt/c:v',NS)]
                if categories!=spec['categories']:errors.append('chart categories changed')
            with zipfile.ZipFile(io.BytesIO(z.read(f'word/embeddings/chart{i}.xlsx'))) as workbook:
                for name in workbook.namelist():
                    if name.endswith(('.xml','.rels')):E.fromstring(workbook.read(name))
                sheet=E.fromstring(workbook.read('xl/worksheets/sheet1.xml'))
                cells={c.get('r'):c for c in sheet.iter('{http://schemas.openxmlformats.org/spreadsheetml/2006/main}c')}
                for j,s in enumerate(spec['series']):
                    values=[float(cells[chr(66+j)+str(k+2)].find('{http://schemas.openxmlformats.org/spreadsheetml/2006/main}v').text) for k in range(len(spec['categories']))]
                    if values!=s['values']:errors.append('editable workbook data changed')
        if links<source.links:errors.append(f'hyperlink count mismatch: {links}/{source.links}')
        if source.colspan and not tree.findall('.//w:gridSpan',NS):errors.append('merged columns missing')
        if source.rowspan and not tree.findall('.//w:vMerge',NS):errors.append('merged rows missing')
        paras=tree.findall('.//w:p',NS)
        count=sum(bool(p.findall('.//w:t',NS)) for p in paras)
        if count<25:errors.append('insufficient native text paragraphs')
        return dict(errors=errors,source_characters=sum(expected.values()),missing_characters=sum(missing.values()),
                    source_body_characters_preserved=not missing,content_coverage=round((sum(expected.values())-sum(missing.values()))/max(1,sum(expected.values()))*100,2),
                    body_character_order_preserved=cursor>=0,
                    native_text_paragraphs=count,native_tables=len(tree.findall('.//w:tbl',NS)),
                    source_tables=source.tags['table'],source_elements=sum(source.tags.values()),source_images=source.images,
                    native_drawings=images,native_pictures=pictures,native_charts=charts,native_checkboxes=len(checkboxes),native_checked=checked,embedded_workbooks=len(source.charts),embedded_images=len([n for n in z.namelist() if n.startswith('word/media/')]),
                    native_hyperlinks=links,native_list_paragraphs=len(tree.findall('.//w:numPr',NS)),
                    native_colspan=bool(tree.findall('.//w:gridSpan',NS)),native_rowspan=bool(tree.findall('.//w:vMerge',NS)),
                    native_columns=len(tree.findall('.//w:cols',NS)),native_editable_body=True)

def engine_hashes():
    return {str(p.relative_to(ROOT)):hashlib.sha256(p.read_bytes()).hexdigest() for p in sorted((ROOT/'internal/convert').glob('*')) if p.is_file() and not p.name.endswith('_test.go')}

EVIDENCE={'created_at':time.strftime('%Y-%m-%dT%H:%M:%S%z'),'engine_before':engine_hashes(),
          'configuration':{'page_size':'A4','margin_mm':20,'page_numbers':True,'media':'print','remote_resources':False},
          'samples':[],'edit_reflow':[],'word_wps_client_validation':False,'pixel_parity_tested':False}
for key,info in MANIFEST.items():
    if key not in SELECTED:
        EVIDENCE['samples'].append(PRIOR_ITEMS[key]);continue
    source=(BATCH/'files'/f'{key}.html').read_text();parsed=Source();parsed.feed(source)
    item={'id':key,'title':info['title'],'features':info['features'],'html_sha256':hashlib.sha256(source.encode()).hexdigest()}
    payload={'html':source,'filename':info['title'],'preview':True,'options':{'page_size':'A4','margin_mm':20,'page_numbers':True,'media':'print'}}
    started=time.monotonic()
    try:
        request=urllib.request.Request(A.base+'/api/convert',json.dumps(payload).encode(),{'Content-Type':'application/json'})
        data=json.load(HTTP.open(request,timeout=120))
        raw=base64.b64decode(data['docx_base64'])
        (BATCH/'files'/f'{key}.docx').write_bytes(raw)
        if not A.batch_dir:
            examples=ROOT/'tools/validation/artifacts/examples';examples.mkdir(parents=True,exist_ok=True)
            (examples/f'{key}.docx').write_bytes(raw)
        folder=OUT/key;folder.mkdir(exist_ok=True)
        (folder/'api.json').write_text(json.dumps({k:v for k,v in data.items() if k not in ['docx_base64','pdf_base64','preview_pages']},ensure_ascii=False,indent=2))
        if data.get('pdf_base64'):(folder/'api.pdf').write_bytes(base64.b64decode(data['pdf_base64']))
        item.update(package_checks(raw,parsed,info['checks']))
        item.update(converted=True,docx_bytes=len(raw),docx_sha256=hashlib.sha256(raw).hexdigest(),stats=data['stats'],warnings=data['warnings'],
                    conversion_ms=data['duration_ms'],request_including_preview_ms=round((time.monotonic()-started)*1000),api_pages=len(data.get('preview_pages',[])),preview_error=data.get('preview_error'))
    except Exception as e:
        item.update(converted=False,errors=[str(e)])
    EVIDENCE['samples'].append(item)
    print('CONVERT',key,'ok' if item['converted'] else 'failed',item.get('api_pages'),item.get('errors'),flush=True)

def render(key,path):
    folder=OUT/key/'render';folder.mkdir(parents=True,exist_ok=True)
    for p in folder.glob('page-*.png'):p.unlink()
    env=os.environ.copy();env['FONTCONFIG_FILE']=str(Path(A.fontconfig).resolve());env['SAL_FONTPATH']='/System/Library/Fonts'
    proc=subprocess.run([sys.executable,A.renderer,str(path.resolve()),'--output_dir',str(folder),'--emit_pdf'],env=env,capture_output=True,text=True,timeout=150)
    if proc.returncode:raise RuntimeError(proc.stdout+'\n'+proc.stderr)
    pages=sorted(folder.glob('page-*.png'),key=lambda p:int(p.stem.split('-')[-1]))
    if not pages:raise RuntimeError('no rendered pages')
    return len(pages)

with ThreadPoolExecutor(max_workers=2) as pool:
    pending={pool.submit(render,i['id'],BATCH/'files'/(i['id']+'.docx')):i for i in EVIDENCE['samples'] if i.get('converted') and i['id'] in SELECTED}
    for future in as_completed(pending):
        item=pending[future]
        try:item['canonical_pages']=future.result();print('RENDER',item['id'],item['canonical_pages'],flush=True)
        except Exception as e:item.setdefault('errors',[]).append('render: '+str(e));print('RENDER FAILED',item['id'],str(e),flush=True)

# Reflow probes modify only native DOCX paragraphs; the source HTML is not rerun.
edit_text='新增内容用于核对完整段落插入后的自然重排和自动分页。This paragraph was inserted directly into native Word content.'
paragraph='<w:p><w:pPr><w:spacing w:after="160"/><w:rPr><w:rFonts w:ascii="Arial" w:eastAsia="PingFang SC"/><w:sz w:val="22"/></w:rPr></w:pPr><w:r><w:rPr><w:rFonts w:ascii="Arial" w:eastAsia="PingFang SC"/><w:sz w:val="22"/></w:rPr><w:t>'+edit_text+'</w:t></w:r></w:p>'
for key in ['batch_annual','batch_quote','batch_handbook']:
    if key not in SELECTED:
        EVIDENCE['edit_reflow'].append(next(i for i in PRIOR['edit_reflow'] if i['id']==key));continue
    source=BATCH/'files'/f'{key}.docx'
    current=next(i for i in EVIDENCE['samples'] if i['id']==key)
    if not current.get('converted') or not current.get('canonical_pages'):continue
    with zipfile.ZipFile(source) as z:files={n:z.read(n) for n in z.namelist()}
    body=files['word/document.xml'].decode()
    # The first table is the top-level masthead. Insert in the document body,
    # after its required trailing paragraph, before the main content.
    index=body.index('</w:tbl>')+len('</w:tbl>');index=body.index('</w:p>',index)+len('</w:p>')
    original=E.fromstring(body);original_text=''.join(e.text or '' for e in original.findall('.//w:t',NS))
    files['word/document.xml']=(body[:index]+paragraph*30+body[index:]).encode()
    edited=OUT/(key+'_edited.docx')
    with zipfile.ZipFile(edited,'w',zipfile.ZIP_DEFLATED) as z:
        for n,data in files.items():z.writestr(n,data)
    tree=E.fromstring(files['word/document.xml']);new_text=''.join(e.text or '' for e in tree.findall('.//w:t',NS))
    check=not Counter(normalized(original_text))-Counter(normalized(new_text))
    baseline=current['canonical_pages']
    pages=render(key+'_edited',edited)
    record={'id':key,'inserted_native_paragraphs':30,'original_content_preserved':check,'before_pages':baseline,'after_pages':pages,'increased_pages':pages>baseline,'method':'direct OOXML paragraph insertion then LibreOffice rendering','client_editing_tested':False}
    EVIDENCE['edit_reflow'].append(record);print('EDIT',json.dumps(record,ensure_ascii=False),flush=True)

EVIDENCE['engine_after']=engine_hashes();EVIDENCE['engine_unchanged']=EVIDENCE['engine_before']==EVIDENCE['engine_after']
(BATCH/'results.json').write_text(json.dumps(EVIDENCE,ensure_ascii=False,indent=2)+'\n')
print('FINISHED',len(EVIDENCE['samples']),'samples',sum(i.get('canonical_pages',0) for i in EVIDENCE['samples']),'pages',flush=True)
if any(i.get('errors') or not i.get('converted') for i in EVIDENCE['samples']):sys.exit(1)

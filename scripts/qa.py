"""HTTP integration + package integrity QA. Uses only Python's standard library."""
import argparse, base64, io, json, pathlib, urllib.request, zipfile
from xml.etree import ElementTree as ET

p=argparse.ArgumentParser()
p.add_argument('--base',default='http://127.0.0.1:8787')
p.add_argument('--out',required=True)
a=p.parse_args()
root=pathlib.Path(__file__).resolve().parents[1]
out=pathlib.Path(a.out);out.mkdir(parents=True,exist_ok=True)
ns={'w':'http://schemas.openxmlformats.org/wordprocessingml/2006/main'}
report=[]
for source in sorted((root/'testdata/html').glob('*.html')):
    payload={'html':source.read_text(),'filename':source.stem,'preview':True,'options':{'page_size':'A4','margin_mm':20,'page_numbers':True}}
    req=urllib.request.Request(a.base+'/api/convert',json.dumps(payload).encode(),{'Content-Type':'application/json'})
    data=json.load(urllib.request.urlopen(req,timeout=100))
    docx=base64.b64decode(data['docx_base64']);(out/(source.stem+'.docx')).write_bytes(docx)
    if data.get('pdf_base64'):(out/(source.stem+'.pdf')).write_bytes(base64.b64decode(data['pdf_base64']))
    with zipfile.ZipFile(io.BytesIO(docx)) as z:
        for name in z.namelist():
            if name.endswith(('.xml','.rels')):ET.fromstring(z.read(name))
        body=z.read('word/document.xml');tree=ET.fromstring(body)
        text=''.join(tree.itertext())
        assert b'altChunk' not in body and b'txbxContent' not in body and b'wp:anchor' not in body
        assert tree.findall('.//w:p',ns),source
        if source.stem=='table':
            assert tree.findall('.//w:vMerge',ns)
            assert tree.findall('.//w:gridSpan',ns)
        if source.stem=='layout':assert any(n.startswith('word/media/') for n in z.namelist())
        if source.stem=='resume_editorial':
            assert len([n for n in z.namelist() if n.startswith('word/media/')])==4
            assert tree.findall('.//w:vMerge',ns) and tree.findall('.//w:gridSpan',ns)
            for token in ['许知夏','Canvas 协作工作台','Mori 内容服务平台','Orbit 运营控制台','能力与方法','工作方式']:
                assert token in text,token
    report.append({'sample':source.stem,'stats':data['stats'],'warnings':data['warnings'],'duration_ms':data['duration_ms'],'bytes':len(docx),'pdf':bool(data.get('pdf_base64')),'rendered_page_count':len(data.get('preview_pages',[]))})
    print(source.stem,json.dumps(report[-1],ensure_ascii=False))
(out/'report.json').write_text(json.dumps(report,ensure_ascii=False,indent=2))

# Add real native paragraphs in the middle of a generated document. The renderer
# must paginate them; no source HTML regeneration or absolute positions are used.
docx=(out/'report.docx').read_bytes()
with zipfile.ZipFile(io.BytesIO(docx)) as z:
    files={n:z.read(n) for n in z.namelist()}
body=files['word/document.xml'].decode()
paragraph='<w:p><w:pPr><w:spacing w:after="180"/></w:pPr><w:r><w:t>新增段落用于验证原生文档自动重排和分页。This is newly inserted editable text.</w:t></w:r></w:p>'
index=body.index('</w:p>')+len('</w:p>')
files['word/document.xml']=(body[:index]+paragraph*60+body[index:]).encode()
with zipfile.ZipFile(out/'report-edited.docx','w',zipfile.ZIP_DEFLATED) as z:
    for n,b in files.items():z.writestr(n,b)
print('Inserted 60 native paragraphs into report-edited.docx')

"""Build static, crawlable documentation using only the Python standard library."""
from pathlib import Path
import argparse, html, json, os, re
from urllib.parse import urlsplit, unquote
from html.parser import HTMLParser

ROOT = Path(__file__).resolve().parents[1]
DOCS = ROOT / 'docs'
BASE = 'https://xinhaolin.github.io/htmlToWord/'
REPO = 'https://github.com/xinhaolin/htmlToWord'
VERSION = 'v0.1.0'
RELEASE = REPO + '/releases/tag/' + VERSION
PAGES = [
 ('guides/getting-started.html','getting-started.md','Convert HTML to Word in Go','Complete Go HTML to DOCX tutorial: installation, editable paragraphs, local assets and a command-line converter.','en'),
 ('guides/native-objects.html','native-objects.md','Native Word charts and checkboxes','Export annotated HTML charts as Office charts with XLSX data, and HTML checkbox inputs as selectable Word controls.','en'),
 ('guides/pagination.html','pagination.md','Editable DOCX pagination and fidelity','How HTMLWord preserves document flow, page settings and editing behavior, with measured validation and CSS limitations.','en'),
 ('articles/html-to-docx-go.html','article-en.md','Converting HTML to editable DOCX in Go','Implementation notes on Chromium layout, Go OOXML, native charts, checkboxes and testing complex HTML to Word conversion.','en'),
 ('articles/html-to-word-go-zh.html','article-zh.md','用 Go 把复杂 HTML 转成可编辑 Word','HTMLWord 的技术实现、图表与复选框原生映射、自动分页验证，以及复杂 HTML 转 Word 的支持边界。','zh-CN'),
]
EXAMPLES = [
 ('batch_annual','Operating review','Native charts, KPI panels and a data table'),
 ('batch_agreement','Service agreement','Clauses, checkboxes and signature areas'),
 ('batch_brand','Brand guidelines','Typography, color swatches and modular layouts'),
 ('batch_catalogue','Device catalogue','Product imagery, specifications and columns'),
 ('batch_event','Event program','Schedules, speakers and editorial sections'),
 ('batch_fieldguide','Field guide','Illustrations, notes and multi-column content'),
 ('batch_handbook','Team handbook','Long text, sidebars and document flow'),
 ('batch_incident','Incident review','Timeline, status panels and action tables'),
 ('batch_plan','Project plan','Milestones, responsibilities and nested layouts'),
 ('batch_quote','Commercial quotation','Pricing tables, totals and agreement details'),
 ('batch_spec','Technical specification','Dense tables, diagrams and selectable checks'),
 ('batch_workbook','Learning workbook','Exercises, writing space and checklists'),
]


def esc(text): return html.escape(str(text), quote=True)

def inline(text):
    text = esc(text)
    text = re.sub(r'`([^`]+)`', r'<code>\1</code>', text)
    text = re.sub(r'\[([^\]]+)\]\(([^\s)]+)\)', r'<a href="\2">\1</a>', text)
    return text

def markdown(source):
    """Render the small, documented Markdown subset used by our guides."""
    out, paragraph, code, in_code, in_list = [], [], [], False, False
    def flush():
        if paragraph:
            out.append('<p>' + inline(' '.join(paragraph)) + '</p>'); paragraph.clear()
    for line in source.splitlines():
        if line.startswith('```'):
            flush()
            if in_list: out.append('</ul>'); in_list = False
            if in_code: out.append('<pre><code>' + esc('\n'.join(code)) + '</code></pre>'); code.clear()
            in_code = not in_code; continue
        if in_code: code.append(line); continue
        if not line.strip():
            flush()
            if in_list: out.append('</ul>'); in_list = False
        elif line.startswith('#') and re.match(r'^#{1,3} ',line):
            flush()
            if in_list: out.append('</ul>'); in_list = False
            level = len(line)-len(line.lstrip('#'))
            out.append(f'<h{level}>' + inline(line[level+1:]) + f'</h{level}>')
        elif line.startswith('- '):
            flush()
            if not in_list: out.append('<ul>'); in_list = True
            out.append('<li>'+inline(line[2:])+'</li>')
        else: paragraph.append(line.strip())
    flush()
    if in_list: out.append('</ul>')
    if in_code: raise ValueError('Unclosed Markdown code fence')
    return '\n'.join(out)

def rel(page, target): return os.path.relpath(target, str(Path(page).parent)).replace(os.sep,'/')

def shell(page,title,description,content,lang='en',kind='WebPage'):
    def link(target): return rel(page,target)
    zh = lang == 'zh-CN'
    alternate = ''
    pairs = {'index.html':'zh/index.html','zh/index.html':'index.html','articles/html-to-docx-go.html':'articles/html-to-word-go-zh.html','articles/html-to-word-go-zh.html':'articles/html-to-docx-go.html'}
    if page in pairs:
        en = pairs[page] if zh else page; cn = page if zh else pairs[page]
        alternate = f'<link rel="alternate" hreflang="en" href="{BASE+en}"><link rel="alternate" hreflang="zh-CN" href="{BASE+cn}"><link rel="alternate" hreflang="x-default" href="{BASE+en}">'
    structured = {'@context':'https://schema.org','@type':kind,'name':title,'description':description,'url':BASE+page,'inLanguage':lang,'isPartOf':{'@type':'WebSite','name':'HTMLWord','url':BASE}}
    if kind == 'SoftwareApplication':
        structured.update(applicationCategory='DeveloperApplication',operatingSystem='macOS, Linux',softwareVersion=VERSION,license=REPO+'/blob/main/LICENSE',codeRepository=REPO)
    if kind == 'TechArticle': structured.update(author={'@type':'Person','name':'xinhaolin','url':'https://github.com/xinhaolin'})
    json_ld = json.dumps(structured,ensure_ascii=False).replace('<', chr(92)+'u003c')
    return f'''<!doctype html>
<html lang="{lang}"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">
<title>{esc(title)} | HTMLWord</title><meta name="description" content="{esc(description)}"><link rel="canonical" href="{BASE+page}">{alternate}
<meta property="og:type" content="{'article' if kind=='TechArticle' else 'website'}"><meta property="og:title" content="{esc(title)}"><meta property="og:description" content="{esc(description)}"><meta property="og:url" content="{BASE+page}"><meta property="og:image" content="{BASE}assets/social-preview.png"><meta property="og:image:alt" content="HTMLWord converts HTML to editable Word documents in Go"><meta name="twitter:card" content="summary_large_image">
<link rel="icon" href="{link('assets/wordmark.svg')}" type="image/svg+xml"><link rel="stylesheet" href="{link('assets/site.css')}"><link rel="alternate" type="application/atom+xml" href="{link('feed.xml')}" title="HTMLWord notes">
<script type="application/ld+json">{json_ld}</script><script src="{link('assets/site.js')}" defer></script></head>
<body><a class="skip" href="#content">{'跳到正文' if zh else 'Skip to content'}</a><header><a class="wordmark" href="{link('index.html')}">HTMLWord<span>Go SDK</span></a><nav aria-label="Main"><a href="{link('guides/getting-started.html')}">{'入门' if zh else 'Quick start'}</a><a href="{link('examples/index.html')}">{'案例' if zh else 'Examples'}</a><a href="{link('articles/html-to-word-go-zh.html' if zh else 'articles/html-to-docx-go.html')}">{'技术文章' if zh else 'How it works'}</a><a href="{link('index.html' if zh else 'zh/index.html')}">{'English' if zh else '简体中文'}</a><a href="{REPO}">GitHub ↗</a></nav></header>
<main id="content">{content}</main><footer><a class="wordmark" href="{link('index.html')}">HTMLWord</a><p>Static HTML. Editable documents. Go SDK.</p><div><a href="{RELEASE}">{VERSION}</a><a href="{REPO}/blob/main/LICENSE">Apache-2.0</a><a href="{REPO}/issues">Issues</a><a href="{REPO}/security/advisories/new">Report a vulnerability</a><a href="{link('validation.md')}">Validation scope</a></div></footer></body></html>'''

def home(page, zh=False):
    r=lambda target: rel(page,target)
    title='用 Go 将 HTML 转成可编辑 Word' if zh else 'HTML to Word in Go. Keep it editable.'
    description='将静态 HTML/CSS 转成可编辑 DOCX。支持原生段落、表格、显式数据图表、复选框和自然分页，需要 Chromium。' if zh else 'Convert static HTML and CSS to editable DOCX with a Go SDK. Native text, tables, annotated charts and checkbox controls. Requires Chromium.'
    content=f'''<section class="hero"><div><p class="eyebrow">HTML → DOCX / {VERSION}</p><h1>{'HTML 转 Word。<br><em>继续自由编辑。</em>' if zh else 'HTML to Word.<br><em>Keep it editable.</em>'}</h1><p class="lead">{description}</p><div class="actions"><a class="button" href="{r('guides/getting-started.html')}">{'开始使用' if zh else 'Start with Go'}</a><a class="text-action" href="{r('examples/index.html')}">{'查看真实导出' if zh else 'Explore real output'} ↗</a></div><div class="install"><code>go get github.com/xinhaolin/htmlToWord@{VERSION}</code><button data-copy="go get github.com/xinhaolin/htmlToWord@{VERSION}" aria-label="Copy install command">Copy</button></div><p class="requirements">Go 1.23+ · macOS / Linux · Chrome / Chromium</p></div><figure class="hero-document"><img src="{r('assets/sdk-report.png')}" width="1275" height="1650" alt="Actual generated DOCX operating report with native Office chart and editable table"><figcaption>{'实际 DOCX 渲染 · LibreOffice' if zh else 'Actual DOCX render · LibreOffice'}</figcaption></figure></section>
<section class="features" aria-label="Capabilities"><article><span>01</span><h2>{'保留文档内容' if zh else 'Native content'}</h2><p>{'文字、列表与合并表格保留为可编辑对象。' if zh else 'Text, lists and merged tables remain editable document content.'}</p></article><article><span>02</span><h2>{'保留原生对象' if zh else 'Office objects'}</h2><p>{'显式图表数据写入原生图表和 XLSX；checkbox 写为内容控件。' if zh else 'Annotated charts include XLSX data. HTML checkboxes become content controls.'}</p></article><article><span>03</span><h2>{'支持后续修改' if zh else 'Room to grow'}</h2><p>{'追加整段文字后，由办公软件重新排版和分页。' if zh else 'Add paragraphs after export and let the office editor reflow the document.'}</p></article></section>
<section class="demo-section"><div><p class="eyebrow">{'真实编辑记录' if zh else 'AFTER EXPORT'}</p><h2>{'看看导出之后。' if zh else 'The document keeps working.'}</h2><p>{'实际 WPS 画面的剪辑演示，展示文字修改、复选框和自然跨页。查看' if zh else 'An edited sequence of actual WPS captures: text editing, checkbox interaction and page reflow. Read the'} <a href="{r('assets/editing-demo.json')}">{'素材与验证记录' if zh else 'capture and verification record'}</a>{'。' if zh else '.'}</p><a href="{BASE}assets/editing-demo.mp4">{'下载演示视频' if zh else 'Download the demo video'} ↗</a></div><video controls preload="none" poster="{r('assets/editing-poster.png')}" aria-label="HTMLWord editing demonstration"><source src="{r('assets/editing-demo.mp4')}" type="video/mp4"><track kind="captions" srclang="en" label="English" src="{r('assets/editing-demo.vtt')}" default></video></section>
<section class="reading"><div><p class="eyebrow">{'可复现起点' if zh else 'REPRODUCIBLE STARTING POINTS'}</p><h2>{'从真实案例开始。' if zh else 'Start with a real document.'}</h2></div><div><a href="{r('examples/index.html')}"><strong>{'12 个复杂案例' if zh else '12 complex examples'}</strong><span>{'报告、合同、规格与手册，附输入和 DOCX。' if zh else 'Reports, agreements, specs and handbooks, with input and DOCX downloads.'}</span></a><a href="{r('guides/native-objects.html')}"><strong>{'图表与复选框' if zh else 'Charts and checkbox controls'}</strong><span>{'原生对象的写法与客户端验证边界。' if zh else 'Explicit data annotations and the office-client validation boundary.'}</span></a><a href="{r('guides/pagination.html')}"><strong>{'分页与支持范围' if zh else 'Pagination and fidelity'}</strong><span>{'了解自然排版、测试证据和 CSS 限制。' if zh else 'Understand document flow, measured evidence and CSS limitations.'}</span></a></div></section>
<section class="boundary"><h2>{'先了解支持边界。' if zh else 'Know the conversion contract.'}</h2><p>{'转换针对静态 HTML；脚本不执行。浏览器与办公软件使用不同的布局模型，复杂 CSS 会产生偏差。图表需要显式数据注解，完整客户端图表数据编辑仍需验证。' if zh else 'This SDK targets static HTML. Scripts do not run. Browser and office layout models differ, so complex CSS can be approximated. Charts require explicit data annotations; complete client-side chart data editing remains to be verified.'}</p><a href="{r('guides/pagination.html')}">{'阅读支持范围' if zh else 'Read the support boundary'} ↗</a></section>'''
    return shell(page,title,description,content,'zh-CN' if zh else 'en','SoftwareApplication')

def gallery():
    cards=[]
    for name,title,description in EXAMPLES:
        cards.append(f'<article class="example"><a href="../assets/example-{name}.png"><img loading="lazy" src="../assets/example-{name}.png" alt="Actual DOCX first page of {esc(title)}"></a><div><h2>{title}</h2><p>{description}</p><a href="{REPO}/blob/v0.1.0/testdata/html/{name}.html">Input HTML</a><a href="{REPO}/releases/download/v0.1.0/{name}.docx">Download DOCX</a></div></article>')
    return shell('examples/index.html','Complex HTML to Word examples','Twelve complex HTML to DOCX examples with actual renders, source HTML, downloadable Word documents and validation scope.',f'''<section class="page-intro"><p class="eyebrow">INPUT / OUTPUT / EVIDENCE</p><h1>Complex pages.<br><em>Real documents.</em></h1><p class="lead">Twelve synthetic fixtures: reports, catalogues, specifications, agreements and handbooks. Each preview is an actual DOCX render; it is not a browser capture.</p><p><a class="button" href="{REPO}/releases/download/v0.1.0/htmlword-examples.zip">Download all inputs + DOCX</a></p><p class="fine">27 original pages · 8 native charts and XLSX workbooks · 8 checkbox controls. Same-environment regression evidence; not universal HTML/Word pixel parity.</p></section><section class="gallery">{''.join(cards)}</section><section class="boundary"><h2>Try the editor you use.</h2><p>Office clients can wrap and render content differently. Download the originals, add paragraphs, test native objects, and report the client version with a minimal reproduction. <a href="../validation.md">Read the full validation record.</a></p></section>''')

def outputs():
    result={'index.html':home('index.html'),'zh/index.html':home('zh/index.html',True),'examples/index.html':gallery()}
    for page,source,title,description,lang in PAGES:
        content=markdown((DOCS/'content'/source).read_text())
        result[page]=shell(page,title,description,'<article class="prose"><p class="eyebrow">'+('TECHNICAL NOTES / 2026-10-22' if page.startswith('articles/') else 'HTMLWORD / GUIDE')+'</p>'+content+'</article>',lang,'TechArticle')
    urls=''.join(f'<url><loc>{BASE+page}</loc></url>' for page in result)
    result['sitemap.xml']='<?xml version="1.0" encoding="UTF-8"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">'+urls+'</urlset>\n'
    result['robots.txt']='User-agent: *\nAllow: /\nSitemap: '+BASE+'sitemap.xml\n'
    result['feed.xml']=f'''<?xml version="1.0" encoding="UTF-8"?><feed xmlns="http://www.w3.org/2005/Atom"><title>HTMLWord technical notes</title><id>{BASE}</id><link href="{BASE}feed.xml" rel="self"/><link href="{BASE}"/><updated>2026-10-22T00:00:00Z</updated><author><name>xinhaolin</name></author><entry><title>Converting HTML to editable DOCX in Go</title><id>{BASE}articles/html-to-docx-go.html</id><link href="{BASE}articles/html-to-docx-go.html"/><updated>2026-10-22T00:00:00Z</updated><summary>Implementation and validation of editable HTML to DOCX conversion in Go.</summary></entry></feed>\n'''
    result['404.html']=shell('404.html','Page not found','Find the HTMLWord documentation, examples and source repository.','<section class="page-intro"><h1>Page not found.</h1><p><a href="'+BASE+'">Open the documentation home</a></p></section>').replace('<meta name="description"','<meta name="robots" content="noindex"><meta name="description"',1)
    result['.nojekyll']=''
    return result

class Links(HTMLParser):
    def __init__(self): super().__init__(); self.targets=[]
    def handle_starttag(self,tag,attrs):
        attrs=dict(attrs)
        for field in ('href','src','poster'):
            if field in attrs: self.targets.append(attrs[field])

def check_links(generated):
    errors=[]
    for page,source in generated.items():
        if not page.endswith('.html'): continue
        parser=Links(); parser.feed(source)
        for target in parser.targets:
            split=urlsplit(target)
            if split.scheme or split.netloc or target.startswith('#'): continue
            path=(DOCS/page).parent/unquote(split.path)
            if not path.exists(): errors.append(f'{page}: missing {target}')
    if errors: raise SystemExit('\n'.join(errors))

if __name__=='__main__':
    args=argparse.ArgumentParser();args.add_argument('--check',action='store_true');args=args.parse_args()
    generated=outputs()
    for page,source in generated.items():
        path=DOCS/page
        if args.check:
            if not path.exists() or path.read_text()!=source: raise SystemExit('Rebuild docs: '+page)
        else: path.parent.mkdir(parents=True,exist_ok=True);path.write_text(source)
    if args.check: check_links(generated)
    print(f'{len(generated)} static outputs '+('verified, including internal links' if args.check else 'built'))

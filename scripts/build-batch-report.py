"""Build the revised report from machine results and explicit manual review.

Prepare screenshots and visual-review.json first. This script never declares
unseen pages reviewed and never substitutes XML integrity for client testing.
"""
from pathlib import Path
from html import escape as h
import hashlib, json

ROOT=Path(__file__).resolve().parents[1]
BATCH=ROOT/'tools/validation/artifacts/batch'
STYLE=(ROOT/'scripts/report-style.css').read_text()

def build():
    data=json.loads((BATCH/'results.json').read_text())
    visual=json.loads((BATCH/'visual-review.json').read_text())
    before=json.loads((BATCH/'baseline-results.json').read_text())
    previous={i['id']:i for i in before['samples']}
    reviews=visual['samples'];items=data['samples']
    for i in items:
        review=reviews[i['id']]
        assert review['reviewed_pages']==list(range(1,i['canonical_pages']+1))
        assert review['docx_sha256']==hashlib.sha256((BATCH/'files'/(i['id']+'.docx')).read_bytes()).hexdigest()
        assert i['source_body_characters_preserved'] and i['body_character_order_preserved'] and not i['errors']
        assert (BATCH/'screenshots'/(i['id']+'-compare.png')).exists()
    pages=sum(i['canonical_pages'] for i in items);oldpages=sum(i['canonical_pages'] for i in before['samples'])
    charts=sum(i['native_charts'] for i in items);checks=sum(i['native_checkboxes'] for i in items)
    nativecases=sum(bool(i['native_charts'] or i['native_checkboxes']) for i in items)
    focus=sum(reviews[i['id']]['group']=='improved' for i in items)
    out=['''<header><div class="wrap"><div class="top"><a href="../">返回转换工作台</a><span>12 种复杂静态 HTML · 2026-10-21 更新</span><a class="button primary" download href="complex-html-word-suite.zip">下载全部样本与报告</a></div><p class="eyebrow">HTML TO WORD / LAYOUT AND NATIVE OFFICE OBJECTS</p><h1>复杂 HTML 转 Word 修复与验证报告</h1><p class="lead">本轮处理长侧栏、卡片成组、尾页和双栏分页，并新增 Office 原生图表与可勾选的复选框。正文继续使用原生段落和表格，支持后续增加内容与自动重排。</p>''']
    out.append(f'<div class="metrics"><div><b>12 / 12</b><span>转换与原生内容检查通过</span></div><div><b>{oldpages} → {pages}</b><span>标准渲染页数</span></div><div><b>{charts}</b><span>原生 Office 图表 含数据工作簿</span></div><div><b>{checks}</b><span>复选框内容控件 保留初始状态</span></div></div></div></header><main class="wrap">')
    out.append('''<div class="callout"><strong>已改善版式，不等于像素级一致</strong><p>此前 7 个重点样本均重新处理。大面积首屏空白、卡片图文拆开、孤立尾页及阅读区双栏失衡已改善。设备手册第 3 页仍较空，字体度量、间距和图表样式仍有差异。真实 WPS 已验证独立探针的复选框点击与保存；图表被识别为原生对象，但客户端“编辑数据”的完整流程尚未通过实测。</p></div>''')
    out.append('''<h2>处理方式与通用范围</h2><div class="panel"><p>转换器没有按案例名称、class 或 id 写分支。本轮依据 DOM 内容关系、计算尺寸及标准 HTML 语义处理以下问题：</p><ol><li>长双栏：短侧栏保留在首段，后续主栏按原宽度转为正文流，让长表正常跨页并重复表头。侧栏背景会在实际内容结束后停止延伸，这是明确的版式适配。</li><li>紧凑卡片和标题行：以原生段落保持关系与表格行分页约束保留图、标题和说明的成组关系；过长内容仍允许跨页。</li><li>多栏：修正连续分节的结束方式，使后续内容能够接在平衡后的双栏下方。</li><li>页尾：短的文档级 footer 转为重复的原生 Word 页脚，消除只放一个品牌标识的尾页；较长或包含复杂结构的 footer 仍留在正文。</li><li>间距：依据实际相邻元素间隔处理正向外边距，减少段距叠加。</li></ol><p class="small">这是一套通用结构策略，但尚未覆盖任意复杂网页。相同输入在 Word、WPS 与 LibreOffice 中仍可能有不同的分页结果。</p></div>''')
    out.append('''<h2>图表与复选框如何使用</h2><div class="panel"><div class="table-wrap"><table><thead><tr><th>对象</th><th>当前实现</th><th>输入与限制</th></tr></thead><tbody><tr><td>数值图表</td><td>折线、纵向柱、横向条、饼图、圆环图；写入 chart XML 和关联的 XLSX 工作簿。</td><td>需要真实类别、系列和数值；JSON 标注图形元素，或给简单数据表标注图表类型。本套 8 张数值图已全部使用原生图表。</td></tr><tr><td>插画与流程图</td><td>保留 9 个图片对象；正文仍为原生文字。</td><td>不能从任意截图可靠恢复底层数值；本轮未实现流程图形状逐项编辑。甘特图保留原生 Word 表格。</td></tr><tr><td>Checkbox</td><td>8 个 w14:checkbox 内容控件；支持勾选和取消，保留 checked 与 disabled。</td><td>识别 input[type=checkbox] 及段首检查符号。静态页面和 PDF 预览中无法点击，需下载 DOCX 到支持控件的办公软件。</td></tr><tr><td>其他表单字段</td><td>保留当前值为原生文字。</td><td>尚未导出交互式下拉或日期控件。</td></tr></tbody></table></div><p>样本只补充了 8 处图表数据标注；去除标注后，12 份 HTML 与修复前逐字相同，未通过改简单模板来掩盖转换差异。</p><pre>&lt;div data-office-chart=\'{"type":"line","categories":["一月","二月"],"series":[{"name":"完成量","values":[12,23]}]}\' style="width:500px;height:200px"&gt;…原视觉内容…&lt;/div&gt;
&lt;table data-office-chart="column"&gt;…第一列类别，其余列为系列数值…&lt;/table&gt;
&lt;label&gt;&lt;input type="checkbox" checked&gt; 已完成确认&lt;/label&gt;</pre><p class="small">已验证 5 种图表的生成与渲染；本套样本实际使用折线、柱形和圆环 3 种。图表类别、数值、缓存与嵌入工作簿逐项核对，外观由办公软件绘制，刻度、线宽与留白可能与原 SVG 不同。</p></div>''')
    out.append(f'''<h2>12 个样本的结果</h2><div class="filters"><button data-filter="all" aria-pressed="true">全部 12 个</button><button data-filter="improved" aria-pressed="false">原重点样本 {focus} 个</button><button data-filter="native" aria-pressed="false">含原生图表或控件 {nativecases} 个</button></div><div class="panel table-wrap"><table><thead><tr><th>案例</th><th>前后页数</th><th>文字段落</th><th>图表 / 图片</th><th>复选框</th><th>观察</th><th>下载</th></tr></thead><tbody>''')
    for i in items:
        key=i['id'];r=reviews[key];isnative=bool(i['native_charts'] or i['native_checkboxes']);label='重点已处理' if r['group']=='improved' else '结构较稳'
        out.append(f'<tr data-group="{r["group"]}" data-native="{str(isnative).lower()}"><td><a href="#{key}">{h(i["title"])}</a></td><td>{previous[key]["canonical_pages"]} → {i["canonical_pages"]}</td><td>{i["native_text_paragraphs"]}</td><td>{i["native_charts"]} / {i["native_pictures"]}</td><td>{i["native_checkboxes"]}</td><td>{label}</td><td><a href="files/{key}.docx" download>Word</a></td></tr>')
    out.append('''</tbody></table></div><p class="small">下方并排图左侧取自先前浏览器截图中的 HTML 原稿，右侧是本轮 DOCX 标准渲染第一页。它是有来源说明的对照合成图，不是本轮工作台界面截图。源视觉标记未改变，见 source-invariance.json。可展开查看本轮全部 Word 页面。</p>''')
    for index,i in enumerate(items,1):
        key=i['id'];r=reviews[key];isnative=bool(i['native_charts'] or i['native_checkboxes']);grade=r['group'];label='重点已处理' if grade=='improved' else '结构较稳'
        out.append(f'<article class="case" id="{key}" data-group="{grade}" data-native="{str(isnative).lower()}"><div class="case-head"><div><p>CASE {index:02d} / {key}</p><h3>{h(i["title"])}</h3><p>{h(i["features"])}</p></div><span class="badge {"focus" if grade=="improved" else "stable"}">{label}</span></div><div class="facts"><span><b>{previous[key]["canonical_pages"]} → {i["canonical_pages"]}</b> 页</span><span><b>{i["native_charts"]}</b> 原生图表</span><span><b>{i["native_pictures"]}</b> 图片</span><span><b>{i["native_checkboxes"]}</b> 复选控件</span><span><b>{i["source_characters"]:,}</b> 个源字符保留</span></div><div class="finding"><strong>本轮结果</strong><span>{h(r["observed_result"])}</span></div><div class="finding"><strong>剩余差异</strong><span>{h(r["remaining_differences"])}</span></div><div class="actions"><a class="button primary" href="http://127.0.0.1:8787/?sample={key}&compare=1&auto=1">实时转换对比</a><a class="button" download href="files/{key}.docx">下载 Word</a><a class="button" download href="files/{key}.html">下载 HTML</a></div><a class="compare" href="screenshots/{key}-compare.png" target="_blank"><img loading="lazy" src="screenshots/{key}-compare.png" alt="{h(i["title"])}源 HTML 与当前 DOCX 第一页对照"></a><p class="caption">左：先前 HTML 浏览器可见区域；右：本轮 DOCX 完整第一页。HTML 为连续页面，显示范围与 Word 分页不同。</p><details><summary>查看本轮全部 {i["canonical_pages"]} 页</summary><div class="issue-gallery">')
        for n in range(1,i['canonical_pages']+1):
            p=f'screenshots/{key}-word-page-{n}.png';out.append(f'<a href="{p}" target="_blank"><img loading="lazy" src="{p}" alt="{h(i["title"])} Word 第 {n} 页">当前 Word 第 {n} 页</a>')
        out.append('</div></details></article>')
    out.append('''<h2>真实 WPS 交互检查</h2><div class="panel"><p>用独立的折线图与两个复选框探针打开 WPS。第二项原先未勾选，实际点击后保存；重新检查保存文件，两个控件均为已勾选，原图表数值 12.5 与 23 保留。图表选择后显示“图表工具”，确认客户端按原生图表识别。</p><div class="issue-gallery"><a href="screenshots/wps-checkbox-toggle.png" target="_blank"><img src="screenshots/wps-checkbox-toggle.png" alt="WPS 中点击复选框并保存后的实际截图">点击复选框并保存的实际界面</a><a href="screenshots/wps-chart-tools.png" target="_blank"><img src="screenshots/wps-chart-tools.png" alt="WPS 将图表识别为原生对象后显示图表工具">原生图表工具实际界面</a></div><p><strong>尚未验证：</strong>WPS 内“编辑数据”未成功进入可操作的数据窗口，因此不能把嵌入工作簿检查写成客户端图表数据编辑验收。Microsoft Word 未安装；12 份样本也未逐份完成真实客户端交互与打印验收。当前交互证据来自独立探针；截图早于勾选符号字体调整，当前导出使用 Arial，控件结构相同。本轮没有扩大到全部样本的客户端验证。</p><div class="actions"><a class="button" download href="client-checks.json">交互检查 JSON</a><a class="button" download href="evidence/wps-saved-probe.docx">WPS 保存后的探针</a></div></div>''')
    out.append('''<h2>连续增加内容与自动重排</h2><div class="panel"><p>从已生成的年报、报价单和入职手册中各插入 30 个原生段落，再由 LibreOffice 重新排版。原有正文保留，三份均从 2 页变为 4 页。追加后的 12 页已逐页查看。这是 DOCX 原生插入与重排检查，不等同于客户端任意位置编辑验收。</p><div class="table-wrap"><table><thead><tr><th>样本</th><th>新增原生段落</th><th>页数变化</th><th>原内容保留</th></tr></thead><tbody>''')
    titles={i['id']:i['title'] for i in items}
    for e in data['edit_reflow']:
        out.append(f'<tr><td>{h(titles[e["id"]])}</td><td>{e["inserted_native_paragraphs"]}</td><td>{e["before_pages"]} → {e["after_pages"]}</td><td>{"是" if e["original_content_preserved"] else "异常"}</td></tr>')
    out.append('''</tbody></table></div></div><h2>检查方法与数据</h2><div class="panel"><ol><li>全部输入采用 A4、20 mm 页边距、打印媒介与页码，资源内嵌，不依赖远程下载。</li><li>12 份重新转换；检查 ZIP/XML、源非空白字符及顺序、关键内容、图片与图表数量、类别与数值、XLSX 工作簿、勾选初始状态、链接及合并表头。源字符检查包含转入原生页脚的文字，不代表所有视觉关系均一致。</li><li>用标准 render_docx.py 与随附 LibreOffice/Poppler 渲染 27 页，并检查 12 页追加段落结果。最终迭代中未改变的页以逐像素相同确认；改变的页重新查看。未执行 HTML 与 Word 的像素差阈值验收。</li><li>Go race 测试、go vet 和脚本语法检查通过；本批转换前后使用同一套转换器源码哈希。</li><li>本轮报告未进行浏览器交互复验；本机页面与附件会通过 HTTP 逐一核对，报告图片的来源分别标明。</li></ol><p class="small">未覆盖任意互联网 HTML、跨系统字体、生产并发与内存曲线。组织、人物、事件和数值均为虚构测试材料。单次本机耗时不能作为吞吐量或生产性能结论。</p><div class="actions"><a class="button" download href="results.json">本轮机器结果</a><a class="button" download href="visual-review.json">逐页视觉记录</a><a class="button" download href="source-invariance.json">HTML 源码对照</a><a class="button" download href="baseline-suite.zip">修复前完整包</a></div><details><summary>复现入口</summary><pre>python scripts/generate-complex-suite.py
python scripts/batch-qa.py --out PATH --renderer PATH_TO_RENDER_DOCX --fontconfig PATH_TO_FONTS_CONF
# 人工检查渲染页并更新 visual-review.json / screenshots
python scripts/build-batch-report.py</pre><p class="small">先前的完整报告与 31 页结果保留在 baseline-suite.zip；本轮结果 JSON 记录实际文件及转换器哈希。</p></details></div>''')
    out.append(f'<footer>生成于 {h(data["created_at"])} · 12 个输入 · {pages} 页本轮 Word · {charts} 个原生图表 · {checks} 个复选框</footer></main>')
    script='''document.querySelectorAll('[data-filter]').forEach(button=>button.addEventListener('click',()=>{const f=button.dataset.filter;document.querySelectorAll('[data-group]').forEach(el=>el.hidden=!(f==='all'||(f==='native'?el.dataset.native==='true':el.dataset.group===f)));document.querySelectorAll('[data-filter]').forEach(el=>el.setAttribute('aria-pressed',String(el===button)))}));'''
    html='<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>复杂 HTML 转 Word 修复报告</title><style>'+STYLE+'</style></head><body>'+''.join(out)+'<script>'+script+'</script></body></html>'
    (BATCH/'report.html').write_text(html)
    print(f'Report built: {len(items)} samples, {pages} pages, {charts} charts, {checks} checkboxes')

if __name__=='__main__':build()

"""Twelve independent static HTML fixtures. All organizations and data are fictional."""
from pathlib import Path
import json
from html import escape

ROOT = Path(__file__).resolve().parents[1]
CSS = '''
*{box-sizing:border-box}body{margin:0;font:12px/1.72 Arial,'PingFang SC',sans-serif;color:#314052;background:white}h1,h2,h3,p,figure,blockquote{margin:0}h1{font-size:32px;line-height:1.28;letter-spacing:-.5px;margin:10px 0 14px}h2{font-size:18px;margin:23px 0 11px;color:var(--ink);break-after:avoid}h3{font-size:13px;margin:10px 0 6px;color:var(--ink);break-after:avoid}p{margin-bottom:9px}a{color:var(--accent);text-decoration:none}ul,ol{padding-left:21px;margin:8px 0 13px}li{margin-bottom:5px}small,.muted{font-size:9px;color:#8190a1}strong{color:var(--ink)}img{max-width:100%;height:auto;display:block}table{border-collapse:collapse;width:100%;font-size:10px;margin:12px 0 17px}th,td{border:1px solid #dce3ea;padding:7px 8px;text-align:left;vertical-align:top}th{background:var(--soft);color:var(--ink)}thead{display:table-header-group}tbody tr:nth-child(even){background:#f8fafb}pre{white-space:pre-wrap;overflow-wrap:anywhere;background:#eef2f6;color:#40516a;padding:13px;font:10px/1.6 Menlo,monospace}code{font:10px Menlo,monospace}.eyebrow{font-size:9px;letter-spacing:2px;color:var(--accent)}.mast{display:flex;justify-content:space-between;gap:14px;padding:12px 0;border-bottom:1px solid #d9e1e9;font-size:8px;letter-spacing:1px;color:#8796a9}.lead{font-size:14px;color:#718096;line-height:1.8}.hero{padding:25px 0 20px}.two{display:grid;grid-template-columns:1.25fr 1fr;gap:20px;margin:16px 0}.equal{grid-template-columns:repeat(2,minmax(0,1fr))}.three{display:grid;grid-template-columns:repeat(3,minmax(0,1fr));gap:13px;margin:16px 0}.four{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:10px;margin:16px 0}.card{padding:14px;border:1px solid #dce4eb;background:white}.card p:last-child{margin-bottom:0}.kpi{padding:13px;background:var(--soft);border-top:3px solid var(--accent)}.kpi b{display:block;font:27px/1.3 Georgia,serif;color:var(--ink)}.kpi span{font-size:9px;color:#8793a2}.note{background:var(--soft);border-left:3px solid var(--accent);padding:13px 15px;margin:15px 0;font-size:11px}.dark{background:var(--ink);color:#dbe5f1;padding:20px}.dark h2,.dark h3,.dark strong{color:white}.tag{display:inline-block;font-size:8px;padding:2px 6px;border:1px solid #d5e0e9;background:var(--soft);margin:0 4px 5px 0;color:var(--accent)}.meta{display:flex;gap:19px;justify-content:space-between;font-size:9px;color:#7c8999;margin:12px 0}.rule{height:4px;background:var(--accent);margin:18px 0}.footer{display:flex;justify-content:space-between;gap:12px;border-top:1px solid #dce3ea;padding:14px 0;margin-top:23px;font-size:8px;color:#8d99a8}.right{text-align:right}.center{text-align:center}.small{font-size:10px}.avoid{break-inside:avoid}.page-break{break-before:page}.split{display:grid;grid-template-columns:144px minmax(0,1fr);gap:22px}.sidebar{background:var(--soft);padding:14px;font-size:10px}.sidebar h3{font-size:11px}.timeline{display:grid;grid-template-columns:72px minmax(0,1fr);gap:14px;margin:14px 0}.timeline>strong{font-size:11px;color:var(--accent)}.listless{list-style:none;padding-left:0}.checkrow{display:flex;gap:7px;align-items:baseline}.step{display:grid;grid-template-columns:30px minmax(0,1fr);gap:10px;margin:14px 0}.step>b{font:23px Georgia,serif;color:var(--accent)}.figure-caption{font-size:8px;color:#8995a3;margin-top:7px}.swatch{height:47px}.swatch-label{padding:8px;font-size:9px}.pill{font-size:9px;letter-spacing:1px;color:var(--accent)}.sheet-title{font:27px Georgia,'Songti SC',serif;color:var(--ink)}
'''
SAMPLES = {}
INFO = {}

def register(key, title, features, body, *, accent='#397384', ink='#233f54', soft='#edf4f6', extra='', checks=()):
    key = 'batch_' + key
    style = f':root{{--accent:{accent};--ink:{ink};--soft:{soft}}}' + CSS + extra
    source = f'<!doctype html><html lang="zh-CN"><head><meta charset="utf-8"><title>{title}</title><style>{style}</style></head><body>{body}</body></html>'
    SAMPLES[key] = source
    INFO[key] = dict(title=title, features=features, note='批量测试新增样本。人物、组织、数值和事件均为虚构；使用同一转换器直接导出。', checks=[title, *checks])

def mast(label, number):
    return f'<div class="mast"><span>QINGYU STUDIO / {label}</span><span>2026 · {number}</span></div>'

def foot(label):
    return f'<footer class="footer"><span>{label} / 青屿工作室</span><span>排版演示资料</span></footer>'

def tags(*items):
    return ''.join(f'<span class="tag">{x}</span>' for x in items)

def rows(items):
    return ''.join('<tr>'+''.join(f'<td>{v}</td>' for v in row)+'</tr>' for row in items)

def chart(kind=0, color='#397384', bg='#edf4f6'):
    if kind == 0:
        shapes = ''.join(f'<path d="M35 {y}H565" stroke="#dbe4e8"/>' for y in [40,80,120,160])
        shapes += '<path d="M40 151L112 125L186 137L260 91L334 99L408 60L482 77L556 30" fill="none" stroke="'+color+'" stroke-width="5"/>'
        shapes += ''.join(f'<circle cx="{x}" cy="{y}" r="5" fill="{color}"/>' for x,y in [(40,151),(112,125),(186,137),(260,91),(334,99),(408,60),(482,77),(556,30)])
    elif kind == 1:
        shapes = ''.join(f'<rect x="{40+i*69}" y="{180-v}" width="40" height="{v}" rx="3" fill="{color}" opacity="{.45+i*.07}"/>' for i,v in enumerate([62,94,76,121,134,110,151,166]))
    elif kind == 2:
        shapes = f'<circle cx="300" cy="108" r="71" fill="none" stroke="#d7e3e9" stroke-width="34"/><circle cx="300" cy="108" r="71" fill="none" stroke="{color}" stroke-width="34" stroke-dasharray="315 446" transform="rotate(-90 300 108)"/>'
    elif kind == 3:
        shapes = f'<rect x="45" y="37" width="125" height="145" rx="10" fill="white" stroke="{color}"/><rect x="238" y="37" width="125" height="145" rx="10" fill="white" stroke="{color}"/><rect x="430" y="37" width="125" height="145" rx="10" fill="white" stroke="{color}"/><path d="M175 108H231m-10-8 10 8-10 8M368 108H423m-10-8 10 8-10 8" fill="none" stroke="{color}" stroke-width="4"/>'
        shapes += ''.join(f'<rect x="{x}" y="{y}" width="{w}" height="8" rx="3" fill="{color}" opacity=".45"/>' for x,y,w in [(64,66,85),(64,90,60),(64,133,80),(256,66,85),(256,90,60),(256,133,80),(448,66,85),(448,90,60),(448,133,80)])
    else:
        shapes = f'<rect x="164" y="33" width="270" height="150" rx="15" fill="white" stroke="{color}" stroke-width="5"/><rect x="180" y="48" width="239" height="115" rx="6" fill="{bg}"/><path d="M270 184v16h60v-16" fill="none" stroke="{color}" stroke-width="5"/><circle cx="300" cy="104" r="32" fill="{color}" opacity=".5"/><path d="M289 90l25 14-25 14Z" fill="white"/>'
    from urllib.parse import quote
    svg = f'<svg xmlns="http://www.w3.org/2000/svg" width="600" height="216" viewBox="0 0 600 216"><rect width="600" height="216" rx="7" fill="{bg}"/>{shapes}</svg>'
    return 'data:image/svg+xml;charset=utf-8,'+quote(svg)

def figure(kind, caption, color='#397384', bg='#edf4f6', native=None):
    if native is None:native=kind<3
    attribute=''
    if native:
        values=[[29,55,43,89,81,120,103,150],[62,94,76,121,134,110,151,166],[315,131]][kind]
        categories=[str(i+1) for i in range(len(values))] if kind<2 else ['长期协作','其他']
        spec=dict(type=['line','column','doughnut'][kind],categories=categories,series=[dict(name=caption,values=values,color=color,colors=[color,'#d7e3e9'])],background=bg,max=180 if kind<2 else 0)
        attribute=' data-office-chart="'+escape(json.dumps(spec,ensure_ascii=False),quote=True)+'"'
    return f'<figure><img{attribute} src="{chart(kind,color,bg)}" alt="{caption}" width="600" height="216"><figcaption class="figure-caption">{caption}</figcaption></figure>'

def build():
    # 01 / editorial business report: charts, KPIs, four-column grid, merged ledger.
    b=mast('OPERATING REVIEW','01')+'''<header class="hero"><span class="eyebrow">ANNUAL OPERATING REVIEW</span><h1>青屿经营回顾与下一年度计划</h1><p class="lead">从交付质量到可持续协作，观察组织能力的结构性变化。</p><div class="meta"><span>报告周期 2026 全年</span><span>编制 经营分析组</span><span>版本 1.0</span></div></header>'''
    b+='<div class="four">'+''.join(f'<div class="kpi"><b>{n}</b><strong>{t}</strong><p><span>{d}</span></p></div>' for n,t,d in [('2,480','服务工时','全年内部统计'),('36','完成项目','含 8 项长期协作'),('94%','按期交付','按计划节点计算'),('18','流程改进','已纳入工作规范')])+'</div>'
    b+='<div class="two"><div><h2>经营摘要</h2><p>本年度工作重点从单点效率转向交付体系。团队将需求澄清、风险发现和质量验证放入日常协作，减少依赖个人经验的重复判断。</p><p>下半年新增的复盘机制覆盖范围评估、文档完整性与反馈闭环。数据用于观察变化，不作为个人绩效排名。</p>'+tags('交付体系','质量治理','知识沉淀')+'</div><div class="note"><h3>管理层关注事项</h3><ol><li>优先处理跨团队等待和反复确认。</li><li>将验收依据写入项目启动材料。</li><li>持续追踪长期维护成本。</li></ol></div></div>'
    b+=figure(0,'图一 月度任务完成趋势 仅为示例',native=True)
    b+='<h2>业务组合与资源分配</h2><table><thead><tr><th rowspan="2">业务单元</th><th colspan="2">年度交付</th><th rowspan="2">下一步重点</th></tr><tr><th>项目数量</th><th>投入占比</th></tr></thead><tbody>'+rows([['协作工具','12','34%','整合高频任务与共享组件'],['内容服务','9','26%','完善审核链路和异常反馈'],['数据体验','8','23%','统一指标解释与追踪口径'],['基础能力','7','17%','持续治理文档和接口约束']])+'</tbody></table>'
    b+='<div class="two equal"><div>'+figure(1,'图二 各季度累计交付量',native=True)+'</div><div>'+figure(2,'图三 长期协作项目占比',native=True)+'</div></div>'
    b+='<h2>行动计划</h2>'+''.join(f'<div class="timeline"><strong>{q}</strong><div><h3>{t}</h3><p>{d}</p></div></div>' for q,t,d in [('第一季度','澄清共同工作标准','统一范围确认、风险登记和验收文档，选择两个项目试运行。'),('第二季度','建立可复用交付单元','梳理高频场景，补齐组件、模板、检查表和维护说明。'),('第三季度','验证协作改进效果','按固定任务追踪等待时长、返工类型与交付可预期程度。')])
    register('annual','青屿经营回顾与下一年度计划','4 列指标 · 3 张图 · 双栏摘要 · 合并表头 · 行动时间线',b+foot('OPERATING REVIEW'),checks=['经营摘要','行动计划'])

    # 02 / formal agreement: hierarchy, clauses, annex and signature boxes.
    b=mast('SERVICE AGREEMENT','02')+'<header class="hero"><span class="eyebrow">SERVICE AGREEMENT / DRAFT</span><h1>协作平台实施服务协议</h1><p class="lead">项目范围 服务交付 验收依据与变更流程</p><div class="meta"><span>编号 QY SA 2026 018</span><span>签订地点 青屿</span><span>版本 讨论稿</span></div></header>'
    b+='<div class="two equal"><div class="card"><h3>委托方</h3><p><strong>青屿内容协作中心</strong></p><p class="small">项目联系人 林澈<br>联系邮箱 client@example.com<br>项目代号 ATLAS</p></div><div class="card"><h3>服务方</h3><p><strong>远岚数字工作室</strong></p><p class="small">项目联系人 周禾<br>联系邮箱 delivery@example.com<br>交付周期 12 周</p></div></div>'
    clauses=[('服务目标与工作边界',['本项目围绕协作平台的配置、迁移方案和使用培训提供实施服务，具体范围以附件清单为准。','双方在启动阶段确认数据字段、访问角色、操作流程和交付环境。未确认的事项记录为待办，不自动扩展为承诺范围。']),('双方协作方式',['委托方提供必要的业务说明与样本数据，并安排具备判断权限的代表参加评审。','服务方按阶段提交文档、配置说明和培训材料，记录已完成事项、依赖及未解决问题。']),('阶段交付与验收',['每个里程碑按约定清单检查。评审意见包含输入条件、观察结果和希望达到的状态。','存在范围差异时，先区分缺陷、改进建议与新增需求，再确认责任人及后续处理计划。']),('范围变更与计划调整',['任何影响工作量、交付时间或验收依据的调整，均进入变更登记表。','双方确认变更的具体内容与影响后再更新计划，并保留前一版本供追踪。']),('资料与访问管理',['项目资料按约定渠道传递，访问范围与任务相关。演示或培训使用经过确认的示例内容。','项目结束时对文档、配置、账号清单与遗留问题进行移交，明确后续维护责任。']),('支持与结束条件',['上线后提供约定周期的使用答疑和缺陷跟踪，新的功能开发另行评估。','双方完成交付清单核对和遗留事项记录后，进入结项归档阶段。'])]
    for i,(title,ps) in enumerate(clauses,1):
        b+=f'<h2>第 {i} 条 {title}</h2><ol>'+''.join(f'<li>{p}</li>' for p in ps)+'</ol>'
    b+='<h2>附件一 阶段成果清单</h2><table><thead><tr><th>阶段</th><th>成果</th><th>验收依据</th><th>负责人</th></tr></thead><tbody>'+rows([['启动','范围说明与任务地图','双方评审记录','项目经理'],['设计','配置与迁移方案','字段映射及权限矩阵','实施顾问'],['验证','试运行与问题清单','固定任务与检查结果','双方代表'],['移交','培训及运维文档','文档目录与确认记录','交付负责人']])+'</tbody></table>'
    b+='<div class="two equal"><div class="card"><h3>委托方确认</h3><p>授权代表 __________________</p><p>日期 ______________________</p><p>签章区域</p></div><div class="card"><h3>服务方确认</h3><p>授权代表 __________________</p><p>日期 ______________________</p><p>签章区域</p></div></div>'
    register('agreement','协作平台实施服务协议','条款层级 · 双方信息卡 · 阶段成果表 · 签署区',b+foot('SERVICE AGREEMENT'),accent='#655a8b',ink='#38324e',soft='#f1eff7',checks=['第 6 条','附件一'])

    # 03 / specification: sidebar navigation, nested acceptance table and code.
    b=mast('PRODUCT REQUIREMENTS','03')+'<header class="hero"><span class="eyebrow">ATLAS / REQUIREMENTS</span><h1>团队知识空间产品需求规格</h1><p class="lead">统一阅读 搜索 编辑与权限反馈的关键任务</p></header>'
    b+='<div class="split"><aside class="sidebar"><h3>文档信息</h3><p>负责人 产品体验组<br>状态 评审中<br>目标版本 2.4</p><h3>章节索引</h3><ol><li><a href="#goals">目标与范围</a></li><li><a href="#flows">任务流程</a></li><li><a href="#features">功能清单</a></li><li><a href="#acceptance">验收矩阵</a></li></ol><h3>评审角色</h3><p>产品 设计<br>研发 测试<br>内容运营</p></aside><main><h2 id="goals">目标与范围</h2><p>使团队成员能够从统一入口找到可信资料，理解文档状态，并在遇到权限或版本问题时获得明确反馈。</p><div class="note"><strong>成功条件</strong><ul><li>关键操作提供可恢复的结果和错误信息。</li><li>相同任务在不同入口保持一致的状态说明。</li><li>权限范围与内容可见性保持对应。</li></ul></div><h2 id="flows">任务流程</h2>'+figure(3,'阅读入口 内容操作与反馈三阶段流程')+'<ol><li>成员进入团队空间，按目录或关键词寻找内容。</li><li>阅读并检查文档的状态、维护者和更新时间。</li><li>执行编辑、收藏或分享，获得可核对的结果。</li></ol><h2 id="features">功能清单</h2><table><thead><tr><th>编号</th><th>场景</th><th>优先级</th><th>输出与边界</th></tr></thead><tbody>'
    feature_names=['空间首页','目录展开','关键词搜索','筛选组合','文档阅读','收藏列表','最近访问','新建草稿','自动保存','编辑冲突','链接分享','权限申请','归档恢复','批量移动','导入文档','导出内容','操作日志','空状态','失败重试','键盘导航','移动浏览','状态通知','版本历史','模板选择']
    b+=rows([[f'F{i+1:02}',n,'P0' if i<9 else 'P1', ['保留任务上下文与明确的操作反馈','权限变化后刷新可见内容','失败时保留输入并允许重试'][i%3]] for i,n in enumerate(feature_names)])+'</tbody></table><h2 id="acceptance">验收矩阵</h2><table><thead><tr><th rowspan="2">任务</th><th colspan="2">验证条件</th></tr><tr><th>输入与操作</th><th>预期行为</th></tr></thead><tbody><tr><td>冲突恢复</td><td><table><tr><th>成员</th><th>版本</th></tr><tr><td>甲</td><td>V12</td></tr><tr><td>乙</td><td>V13</td></tr></table></td><td>说明版本差异，保留本地输入并提供合并入口。</td></tr><tr><td>权限变更</td><td>阅读中撤销访问权限</td><td>停止显示敏感内容并说明可申请的访问方式。</td></tr></tbody></table><h3>示例状态结构</h3><pre>{\n  "state": "conflict",\n  "local_revision": 12,\n  "remote_revision": 13,\n  "recoverable": true\n}</pre></main></div>'
    register('spec','团队知识空间产品需求规格','导航侧栏 · 24 项长表 · 嵌套表格 · 原生锚点 · JSON 代码',b+foot('PRODUCT REQUIREMENTS'),accent='#4969a3',ink='#293e62',soft='#edf2fa',checks=['F24','验收矩阵','recoverable'])

    # 04 / identity manual: visual art, swatches, type scale and before/after.
    b=mast('IDENTITY GUIDELINES','04')+'<header class="hero"><span class="eyebrow">NORTHLINE STUDIO</span><h1>北线品牌视觉规范</h1><p class="lead">建立清晰 克制并可持续维护的视觉语言</p></header>'
    b+='<div class="two equal"><div class="dark"><p class="eyebrow">THE WORDMARK</p><p style="font:43px/1.3 Georgia,serif;letter-spacing:2px;color:white;margin:22px 0">northline.</p><p>让方向清楚，让表达留有空间。</p></div><div><h2>品牌气质</h2><p>采用稳定的字重、宽松的留白和有限的强调色。视觉重点服务于信息层级，使内容在不同媒介中保持可辨识性。</p>'+tags('CLEAR','CALM','CONSISTENT')+'</div></div>'
    b+='<h2>基础色与辅助色</h2><div class="three">'+''.join(f'<div class="card" style="padding:0"><div class="swatch" style="background:{c}"></div><div class="swatch-label"><strong>{n}</strong><p>{c}<br>{d}</p></div></div>' for n,c,d in [('深松绿','#244c45','标题与主标识'),('鼠尾草','#a8b8a1','图表与次级图形'),('暖砂白','#f4efe3','背景与内容底色'),('铜棕','#aa7955','小面积强调'),('雾灰','#8c9793','辅助说明'),('墨黑','#26332e','长文正文')])+'</div>'
    b+='<h2>字体与信息层级</h2><div class="card"><p style="font-size:31px;line-height:1.4;color:#244c45">让重要信息自然显现</p><p style="font-size:18px;line-height:1.5">清晰标题与舒适阅读节奏</p><p>正文用于解释上下文、条件与行动，避免仅靠装饰表达层级。</p><p class="muted">Caption / 说明文字 9 px / 建议至少保留可读对比</p></div>'
    b+='<h2>图形与应用</h2><div class="two equal"><div>'+figure(0,'线性图形在数据场景的应用','#244c45','#f4efe3')+'</div><div>'+figure(4,'数字媒介中的视觉比例','#aa7955','#f4efe3')+'</div></div><table><thead><tr><th>场景</th><th>主元素</th><th>使用规则</th></tr></thead><tbody>'+rows([['网页标题','品牌字标与短标题','保留适当留白，避免重复强调'],['项目封面','标题 时间与责任方','建立一个清晰的阅读入口'],['数据图表','有限强调色','颜色之外使用标签或位置表达'],['说明文档','分级标题和辅助文字','保持信息顺序和排版一致性']])+'</tbody></table>'
    b+='<div class="two equal"><div class="note"><h3>推荐做法</h3><ul><li>根据内容需要选择图形和颜色。</li><li>控制每页视觉重心的数量。</li><li>使用真实文本检查最终排版。</li></ul></div><div class="card"><h3>需要避免</h3><ul><li>同时使用过多的高饱和强调色。</li><li>压缩字宽或拉伸品牌字标。</li><li>只用颜色区分重要状态。</li></ul></div></div>'
    register('brand','北线品牌视觉规范','六色色板 · 字体层级 · 深色品牌卡 · 图形应用 · 双栏规范',b+foot('IDENTITY GUIDELINES'),accent='#6f8874',ink='#244c45',soft='#f4efe3',checks=['基础色与辅助色','需要避免'])

    # 05 / quotation: financial-looking fictional ledger with 28 line items.
    b=mast('QUOTATION AND STATEMENT','05')+'<header class="hero"><div class="two"><div><span class="eyebrow">COMMERCIAL PROPOSAL</span><h1>工作空间升级项目报价单</h1><p class="lead">配置清单与阶段结算说明</p></div><div class="note"><p>报价编号 QY Q 2026 032<br>出具日期 2026 10 20<br>有效版本 第 3 版</p></div></div></header>'
    b+='<div class="three"><div class="card"><h3>采购单位</h3><p class="small">青屿学习中心<br>项目联系人 陈禾<br>buyer@example.com</p></div><div class="card"><h3>服务单位</h3><p class="small">北线协作工作室<br>项目联系人 陆宁<br>quote@example.com</p></div><div class="card"><h3>交付方式</h3><p class="small">分阶段检查与移交<br>金额单位 人民币元<br>价格均为虚构演示</p></div></div>'
    items=['需求梳理','任务地图','空间规划','账号配置','角色设计','目录迁移','内容清理','模板整理','表单配置','权限核对','接口联调','导入验证','搜索配置','操作说明','培训材料','培训演示','试运行','问题跟踪','反馈分析','配置调整','文档移交','维护说明','项目管理','质量检查','设计走查','版本确认','验收记录','结项归档']
    prices=[800+(i%6)*120 for i in range(28)];units=[1+i%3 for i in range(28)];total=sum(a*b for a,b in zip(prices,units))
    b+='<table><thead><tr><th rowspan="2">序号</th><th rowspan="2">服务项目</th><th colspan="2">计价明细</th><th rowspan="2">小计</th></tr><tr><th>数量</th><th>单价</th></tr></thead><tbody>'+rows([[f'{i+1:02}',name,units[i],f'{prices[i]:,}.00',f'{prices[i]*units[i]:,}.00'] for i,name in enumerate(items)])+f'<tr><td colspan="4" class="right"><strong>合计</strong></td><td><strong>{total:,}.00</strong></td></tr></tbody></table>'
    b+='<div class="two"><div><h2>阶段结算</h2><ol><li>启动与范围确认后完成第一阶段记录。</li><li>试运行完成后按已确认成果核对。</li><li>移交时检查文档目录与遗留事项。</li></ol></div><div class="note"><h3>说明</h3><p>本清单仅展示报价结构与排版。新增项目先登记范围、计量方式和计划影响，再更新报价版本。</p></div></div><div class="two equal"><div class="card"><h3>采购方确认</h3><p>签字 __________________<br>日期 __________________</p></div><div class="card"><h3>服务方确认</h3><p>签字 __________________<br>日期 __________________</p></div></div>'
    register('quote','工作空间升级项目报价单','三方信息卡 · 28 行明细 · 横纵合并表头 · 合计与确认区',b+foot('QUOTATION'),accent='#a8743e',ink='#655034',soft='#f7f0e6',checks=['28','结项归档',f'{total:,}.00'])

    # 06 / conference: asymmetric hero, parallel-track schedule and speaker grid.
    b=mast('CONFERENCE PROGRAM','06')+'<header class="hero"><div class="two"><div><span class="eyebrow">CLEAR SYSTEMS FORUM 2026</span><h1>清晰系统技术交流日</h1><p class="lead">从设计决策到可靠交付<br>一次关于复杂产品的共同讨论</p>'+tags('11 月 06 日','青屿交流空间','三条平行议程')+'</div><div>'+figure(3,'论坛视觉图形','#6655a2','#f0edf8')+'</div></div></header>'
    b+='<h2>上午共同议程</h2><table><thead><tr><th>时间</th><th>议题</th><th>形式</th><th>主持</th></tr></thead><tbody>'+rows([['09:00—09:30','签到与案例展览','自由交流','会务组'],['09:30—10:10','复杂系统中的清晰表达','主题分享','林砚'],['10:10—10:50','把验证放进交付过程','案例讨论','周青'],['10:50—11:10','茶歇与交流','开放空间','会务组'],['11:10—12:00','团队协作的边界与接口','圆桌','陈宁']])+'</tbody></table><h2>下午平行议程</h2><table><thead><tr><th rowspan="2">时间</th><th colspan="3">分会场</th></tr><tr><th>设计与体验</th><th>研发与质量</th><th>产品与协作</th></tr></thead><tbody>'+rows([['13:30—14:10','信息架构工作坊','文档结构与兼容性','需求澄清与范围管理'],['14:20—15:00','可访问性检查','自动化验证边界','证据与决策记录'],['15:10—15:50','设计系统维护','跨平台渲染差异','面向长期维护的交付'],['16:00—16:40','复杂表单体验','性能测量与复现','协作接口设计']])+'<tr><td>16:40—17:10</td><td colspan="3">共同回顾 交流发现与后续实践</td></tr></tbody></table>'
    b+='<h2>分享嘉宾</h2><div class="two equal">'+''.join(f'<div class="card"><div class="pill">SPEAKER 0{i+1}</div><h3>{n}</h3><p class="muted">{r}</p><p class="small">{d}</p>'+tags(*t)+'</div>' for i,(n,r,d,t) in enumerate([('林砚','产品设计负责人','关注多角色产品中的任务模型、导航和信息表达。',['信息架构','设计研究']),('周青','质量工程负责人','关注可复现的问题描述、验证范围与交付依据。',['自动化','质量策略']),('陈宁','技术负责人','关注跨平台文档结构、性能与兼容性。',['系统设计','文档技术']),('陆禾','协作实践研究员','关注跨团队责任边界、异步评审与知识维护。',['协作','团队实践'])]))+'</div><div class="note"><h3>参会提示</h3><ul><li>平行议程可按兴趣选择，座位安排以现场引导为准。</li><li>工作坊建议携带可离线使用的案例材料。</li><li>所有嘉宾、地点和时间均为排版演示内容。</li></ul></div>'
    register('event','清晰系统技术交流日','平行日程合并表 · 四位嘉宾卡 · 图文首屏 · 参会提示',b+foot('CONFERENCE PROGRAM'),accent='#7561a9',ink='#3c335c',soft='#f2eff8',checks=['下午平行议程','共同回顾'])

    # 07 / travel-style guide, completely fictional locations.
    b=mast('FIELD STUDY GUIDE','07')+'<header class="hero"><span class="eyebrow">THREE DAYS IN QINGYU</span><h1>青屿城市观察研学行程</h1><p class="lead">用三天观察公共空间 信息服务与城市日常</p><div class="meta"><span>研学小组 16 人</span><span>行程版本 02</span><span>全部地点为虚构</span></div></header>'
    b+='<div class="two"><div>'+figure(3,'三个观察主题的路线示意','#b2734d','#f8efe8')+'</div><div class="note"><h3>观察目标</h3><ul><li>记录人们如何寻找信息与完成任务。</li><li>区分空间设计、信息设计和服务反馈。</li><li>用证据支持讨论，避免仅凭印象归因。</li></ul></div></div>'
    stops=[('DAY 01','公共空间与导向系统',[('09:00','北线广场','观察入口、路径与标识层级'),('11:00','青屿展馆','记录首次到访者的信息需求'),('14:00','云岸图书馆','检查目录、检索与借阅反馈'),('16:30','共学工坊','整理照片、观察笔记与疑问')]),('DAY 02','服务流程与任务体验',[('09:30','风桥服务中心','沿完整任务记录接触点'),('11:30','南汀市集','观察交易、排队与异常处理'),('14:30','微光社区','访谈不同角色的使用情境'),('17:00','共学工坊','绘制服务蓝图与改进假设')]),('DAY 03','信息表达与方案交流',[('09:00','溪石文化馆','分析内容叙事与阅读路径'),('11:00','开放设计室','尝试低成本的信息重组'),('14:00','共学工坊','原型制作与小组走查'),('16:00','青屿交流厅','汇报观察证据与方案边界')])]
    for day,title,events in stops:
        b+=f'<h2>{day} {title}</h2><table><thead><tr><th>时间</th><th>地点</th><th>观察任务</th></tr></thead><tbody>'+rows(events)+'</tbody></table><div class="three"><div class="card"><h3>记录</h3><p class="small">用行为、时间和上下文描述所见，区分事实与推断。</p></div><div class="card"><h3>讨论</h3><p class="small">选择一项具体任务，比较不同角色的目标和约束。</p></div><div class="card"><h3>产出</h3><p class="small">形成一页观察记录，保留未验证的问题。</p></div></div>'
    b+='<h2>随身清单</h2><div class="two equal"><ul><li>纸笔与离线地图资料</li><li>充电设备与文件备份</li><li>已获允许的记录工具</li></ul><ul><li>任务观察表和小组安排</li><li>适合步行的个人用品</li><li>项目联系清单</li></ul></div>'
    register('fieldguide','青屿城市观察研学行程','三天日程 · 12 个观察任务 · 路线图 · 三栏任务卡 · 双栏清单',b+foot('FIELD STUDY GUIDE'),accent='#b2734d',ink='#704b36',soft='#f8efe8',checks=['DAY 03','随身清单'])

    # 08 / project plan: editable Gantt cells plus risk and RACI matrices.
    b=mast('DELIVERY PLAN','08')+'<header class="hero"><span class="eyebrow">ATLAS / DELIVERY OFFICE</span><h1>协作空间迁移项目交付计划</h1><p class="lead">范围 依赖 里程碑与协作责任的统一视图</p></header>'
    b+='<div class="three"><div class="kpi"><b>08</b><strong>计划周数</strong><p class="muted">含试运行与移交</p></div><div class="kpi"><b>12</b><strong>工作包</strong><p class="muted">按可验收成果组织</p></div><div class="kpi"><b>04</b><strong>关键里程碑</strong><p class="muted">每阶段检查依赖与边界</p></div></div><h2>工作分解与排期</h2><table class="gantt"><thead><tr><th>工作包</th><th>责任方</th>'+''.join(f'<th>W{i}</th>' for i in range(1,9))+'</tr></thead><tbody>'
    tasks=[('范围确认','产品',1,1),('内容盘点','运营',1,2),('字段映射','实施',2,3),('角色矩阵','产品',2,3),('迁移工具','研发',3,5),('样本导入','实施',4,4),('任务验证','测试',4,6),('培训材料','运营',5,6),('分批迁移','实施',6,7),('使用答疑','支持',6,8),('遗留关闭','全体',7,8),('文档移交','交付',8,8)]
    for name,owner,start,end in tasks:
        b+=f'<tr><td>{name}</td><td>{owner}</td>'+''.join(f'<td style="background:{"#a7c1bc" if start<=w<=end else "#fff"};text-align:center">{"●" if start<=w<=end else "—"}</td>' for w in range(1,9))+'</tr>'
    b+='</tbody></table><h2>里程碑与检查依据</h2><table><thead><tr><th>节点</th><th>成果</th><th>进入条件</th><th>退出依据</th></tr></thead><tbody>'+rows([['M1 / W2','范围基线','资料与代表确认','范围和依赖记录完成'],['M2 / W4','小范围试迁移','字段映射和权限确认','样本任务通过检查'],['M3 / W7','分批迁移完成','遗留问题已有处理计划','内容和权限核对完成'],['M4 / W8','项目移交','使用支持及文档准备','责任交接与结项记录']])+'</tbody></table><h2>责任矩阵</h2><table><thead><tr><th>工作项</th><th>产品</th><th>研发</th><th>实施</th><th>测试</th></tr></thead><tbody>'+rows([['范围与验收','A','C','R','C'],['迁移与回滚','C','A','R','C'],['任务验证','C','R','C','A'],['使用培训','A','C','R','C']])+'</tbody></table><p class="muted">A 最终负责 R 执行 C 咨询。每项工作保留一个最终负责人。</p><h2>风险登记</h2><div class="two equal"><div class="note"><h3>数据与权限差异</h3><p>先以固定样本验证字段映射和访问范围，再确定迁移批次。保留原始记录与回退说明。</p></div><div class="note"><h3>跨团队依赖延迟</h3><p>每周检查输入是否到位。依赖变化时更新里程碑影响，避免把等待误记为完成。</p></div></div>'
    register('plan','协作空间迁移项目交付计划','8 周原生甘特表 · 12 工作包 · RACI 矩阵 · 风险卡',b+foot('DELIVERY PLAN'),accent='#547d76',ink='#2b514b',soft='#edf4f1',extra='.gantt{font-size:8px}.gantt th,.gantt td{padding:7px 4px}.gantt th:first-child{width:24%}',checks=['文档移交','责任矩阵','风险登记'])

    # 09 / incident review: native code, timeline and recovery decision table.
    b=mast('INCIDENT REVIEW','09')+'<header class="hero"><span class="eyebrow">OPERATIONS / REVIEW 014</span><h1>文档预览任务积压故障复盘</h1><p class="lead">记录可观察事实 分析依赖并建立后续验证计划</p></header><div class="note"><strong>事件摘要</strong><p>示例环境中部分预览任务等待时间上升。处理过程包含流量观察、执行池检查、恢复操作与结果核对。以下事件和数据均为虚构。</p></div>'
    b+='<div class="three"><div class="kpi"><b>42 min</b><strong>观察窗口</strong></div><div class="kpi"><b>18%</b><strong>示例超时比例</strong></div><div class="kpi"><b>0</b><strong>确认数据丢失</strong></div></div>'+figure(0,'图一 示例队列等待趋势','#aa6650','#f7eeea',native=True)+'<h2>事件时间线</h2>'
    b+=''.join(f'<div class="timeline"><strong>{t}</strong><div><h3>{title}</h3><p>{desc}</p></div></div>' for t,title,desc in [('10:02','发现延迟上升','监控中预览等待时间增加，错误率变化晚于等待时间变化。'),('10:08','核对执行池状态','任务仍在入队，渲染进程占用执行槽较长，排队任务逐步累积。'),('10:16','限制新任务并保留输入','降低同时提交的任务数，向调用方返回明确的等待与重试信息。'),('10:28','恢复完成任务回收','检查执行槽释放和进程清理，对完成与失败任务进行核对。'),('10:44','验证恢复与登记后续','固定任务重复运行，确认等待回落并记录尚未验证的假设。')])
    b+='<h2>观察证据与解释边界</h2><table><thead><tr><th>观察</th><th>支持的判断</th><th>不能证明的内容</th></tr></thead><tbody>'+rows([['队列长度持续增加','到达速率高于当前处理速率','单凭队列无法确定根因'],['部分进程完成时间变长','执行时长分布发生变化','不能直接归因于内存泄漏'],['限制入队后等待回落','资源竞争有所缓解','不能代替长期容量验证']])+'</tbody></table><h2>示例日志片段</h2><pre>10:08:03 queue pending=24 active=2\n10:08:07 render task=demo-014 elapsed_ms=18620\n10:16:12 admission state=limited retry_after=5\n10:44:01 probe status=complete content_check=passed</pre><h2>后续行动</h2><ol><li><strong>补齐生命周期检查</strong><ul><li>成功、失败和取消都释放执行槽。</li><li>结束后清理临时目录与独立进程。</li></ul></li><li><strong>验证资源上限</strong><ul><li>以不同页数和图片数量建立固定负载。</li><li>同时记录等待、处理耗时与错误原因。</li></ul></li><li><strong>改进调用方反馈</strong><ul><li>区分转换完成与预览失败。</li><li>提供可判断是否重试的错误信息。</li></ul></li></ol>'
    register('incident','文档预览任务积压故障复盘','事件时间线 · 趋势图 · 证据矩阵 · 日志代码 · 嵌套行动列表',b+foot('INCIDENT REVIEW'),accent='#aa6650',ink='#654237',soft='#f7eeea',checks=['content_check=passed','后续行动'])

    # 10 / product catalogue: six illustrated product cards and comparison table.
    b=mast('PRODUCT CATALOGUE','10')+'<header class="hero"><span class="eyebrow">NORTHLINE / WORKSPACE COLLECTION</span><h1>北线协作设备选型手册</h1><p class="lead">围绕专注 交流与内容呈现组织工作空间</p></header><div class="two"><div class="dark"><h2>从任务选择设备</h2><p>先确认使用者、空间、协作方式与维护条件，再比较配置。以下型号和参数仅用于排版测试。</p>'+tags('个人专注','小组协作','共享展示')+'</div><div>'+figure(4,'设备视觉概念','#4a7898','#edf4f8')+'</div></div><h2>产品系列</h2><div class="two equal">'
    products=[('LINE 24','个人内容终端','24 英寸 / 标准支架','用于资料阅读、内容编辑与日常协作。',4),('LINE 27','多任务工作终端','27 英寸 / 高度调整','用于数据核对、多窗口浏览与设计走查。',4),('ARC HUB','小组协作中枢','共享接入 / 桌面布置','用于小组讨论中的内容共享和任务记录。',3),('ARC VIEW','公共信息终端','横竖切换 / 共享展示','用于公共区域信息、日程和状态展示。',1),('FOCUS MINI','轻量任务终端','紧凑机身 / 固定部署','用于单项业务任务和受限空间。',2),('FOCUS PRO','协作展示终端','多源连接 / 远程协作','用于混合会议与跨地点内容讨论。',0)]
    for i,(name,role,meta,desc,kind) in enumerate(products):
        b+='<article class="card">'+figure(kind,name+' 产品示意','#4a7898','#edf4f8')+f'<p class="pill">SERIES 0{i+1}</p><h3>{name}</h3><p><strong>{role}</strong></p><p class="muted">{meta}</p><p class="small">{desc}</p>'+tags('可维护','模块化','示例型号')+'</article>'
    b+='</div><h2>配置比较</h2><table><thead><tr><th>系列</th><th>典型任务</th><th>部署形式</th><th>检查重点</th></tr></thead><tbody>'+rows([[p[0],p[1],['个人桌面','共享桌面','公共区域'][i%3],['阅读距离与信息密度','接入与协作流程','维护与权限范围'][i%3]] for i,p in enumerate(products)])+'</tbody></table><h2>选择与交付</h2><div class="three"><div class="card"><h3>01 明确任务</h3><p class="small">记录关键操作和真实使用情境。</p></div><div class="card"><h3>02 现场验证</h3><p class="small">用固定任务检查显示与交互效果。</p></div><div class="card"><h3>03 维护移交</h3><p class="small">明确配置、资料和后续责任。</p></div></div>'
    register('catalogue','北线协作设备选型手册','6 个图文产品卡 · 双栏网格 · 参数比较表 · 三步选型',b+foot('PRODUCT CATALOGUE'),accent='#4a7898',ink='#2c4a64',soft='#edf4f8',checks=['FOCUS PRO','配置比较'])

    # 11 / handbook: different colored navigation rail and static form controls.
    b=mast('TEAM HANDBOOK','11')+'<header class="hero"><span class="eyebrow">WELCOME TO QINGYU</span><h1>新成员协作与入职手册</h1><p class="lead">从第一天到第一个月 建立清晰的工作联系</p></header><div class="split"><aside class="sidebar"><h3>联系窗口</h3><p>入职协调 韩宁<br>团队伙伴 周澈<br>技术支持 林禾</p><h3>工作资料</h3><ul><li>任务空间</li><li>团队目录</li><li>知识文档</li><li>支持记录</li></ul><div class="note"><strong>有问题时</strong><p>先说明你要完成的任务、当前状态和阻碍，再一起确认下一步。</p></div></aside><main><h2>第一天 建立连接</h2><p>了解团队正在解决的问题，认识与你协作的角色，确认资料入口与支持方式。无需在第一天理解所有系统和历史决策。</p><div class="card"><h3>开始前核对</h3><p><input type="checkbox" checked> 已收到团队与项目介绍</p><p><input type="checkbox" checked> 已确认工作资料入口</p><p><input type="checkbox"> 已完成账号与权限检查</p><p><input type="checkbox"> 已安排与团队伙伴的交流</p></div><h2>第一周 了解任务与边界</h2>'+figure(3,'了解任务 共同实践与反馈的循环','#8a6b43','#f6f0e6')+'<ol><li>选择一项真实任务，记录输入、输出与完成依据。</li><li>和伙伴共同完成一次小范围交付，观察协作接口。</li><li>保留尚未理解的问题，在固定交流时间核对。</li></ol><h2>第一个月 完成一次交付</h2><table><thead><tr><th>阶段</th><th>建议活动</th><th>可核对产出</th></tr></thead><tbody>'+rows([['第 1 周','观察与结对','任务记录与问题清单'],['第 2 周','独立完成小任务','结果、检查与说明'],['第 3 周','参与共同评审','取舍与反馈记录'],['第 4 周','回顾与计划','后续学习与协作安排']])+'</tbody></table><h2>日常协作约定</h2><div class="note"><h3>让沟通可继续</h3><ul><li>说明背景、目标、约束与需要回应的事项。</li><li>区分事实、判断和待确认的信息。</li><li>决定发生变化时，更新相关任务和文档。</li></ul></div><h2>个人学习记录</h2><table><tr><th>成员</th><td><input value="周以宁"></td></tr><tr><th>当前任务</th><td><input value="熟悉项目知识空间与交付流程"></td></tr><tr><th>希望获得的帮助</th><td><textarea>理解验收方式与跨团队协作接口。</textarea></td></tr></table></main></div>'
    register('handbook','新成员协作与入职手册','导航侧栏 · 勾选状态 · 输入值 · 三阶段计划 · 图文与表格',b+foot('TEAM HANDBOOK'),accent='#8a6b43',ink='#5c4b34',soft='#f6f0e6',checks=['个人学习记录','周以宁','理解验收方式'])

    # 12 / workbook: two-column reading section, nested numbering, forced appendix.
    b=mast('LEARNING WORKBOOK','12')+'<header class="hero"><span class="eyebrow">DOCUMENT ENGINEERING / WORKSHOP</span><h1>结构化文档设计训练册</h1><p class="lead">用任务 例子与检查依据练习清晰表达</p><div class="meta"><span>建议用时 90 分钟</span><span>形式 小组实践</span><span>材料 静态示例</span></div></header><h2>阅读材料</h2><div class="reading"><p>文档的结构不仅决定外观，也决定后续如何编辑、检索和维护。标题、段落、列表与表格应各自表达明确的内容关系。</p><p>当一个段落包含多个判断时，可以先明确主张，再给出证据与适用条件。过度依赖颜色或位置，会使内容在不同呈现环境中失去意义。</p><p>将复杂页面转为流式文档时，需要识别可连续阅读的内容与独立的视觉材料。图示可以作为图片保留，正文则应尽量保持原生结构。</p><p>测试应分别检查内容完整性、视觉结果、编辑行为与运行边界。生成成功只证明某一步完成，不能自动推导其他步骤已经通过。</p></div><h2>练习一 识别内容结构</h2><ol type="I"><li><strong>整理一段需求说明</strong><ol type="a"><li>写出目标、使用角色和主要约束。</li><li>标注哪些信息是事实，哪些仍需验证。</li></ol></li><li><strong>设计一个任务清单</strong><ol type="a"><li>每项任务有明确的动作与完成依据。</li><li>保留依赖、责任方和需要反馈的问题。</li></ol></li><li><strong>检查一张复杂表格</strong><ol type="a"><li>解释表头与行列之间的关系。</li><li>判断合并单元格是否帮助理解。</li></ol></li></ol><h2>练习二 对比三种表达</h2><div class="three"><div class="card"><h3>段落</h3><p class="small">适合解释观点、原因、条件与联系。</p></div><div class="card"><h3>列表</h3><p class="small">适合并列事项、顺序步骤与检查项。</p></div><div class="card"><h3>表格</h3><p class="small">适合按相同字段比较多组信息。</p></div></div><h2>练习三 完成验证记录</h2><table><thead><tr><th>观察对象</th><th>检查方式</th><th>观察结果</th><th>未覆盖边界</th></tr></thead><tbody>'+rows([['正文结构','检查段落与列表','填写观察','填写边界'],['图片位置','查看实际渲染','填写观察','填写边界'],['分页行为','追加完整段落','填写观察','填写边界'],['复杂表格','核对合并与跨页','填写观察','填写边界']])+'</tbody></table><h2>自我检查</h2><ul class="listless"><li>□ 标题能够让读者理解具体主题。</li><li>□ 每个结论有对应的事实或证据。</li><li>□ 测试结果与未验证范围分别记录。</li><li>□ 文档在内容增加后仍能继续阅读。</li></ul><section class="page-break"><h2>参考讨论与延伸练习</h2><div class="note"><p><strong>讨论方向：</strong>结构表达内容关系，版式帮助阅读。两者需要在真实输入与后续编辑中共同验证。</p></div><h3>延伸任务</h3><ol start="4"><li>将同一段材料分别整理为段落、列表和表格。</li><li>追加一倍内容，观察布局与分页是否仍然合理。</li><li>记录一个看似成功但证据不足的结论，并补齐验证方式。</li></ol><pre>检查对象 → 输入条件 → 实际观察 → 判断依据 → 验证边界</pre><h3>小组回顾</h3><p>选择一项有分歧的观察，说明各自的证据，再约定下一步如何验证。</p></section>'
    register('workbook','结构化文档设计训练册','CSS 双栏阅读 · 罗马与字母编号 · 检查表 · 原生表格 · 强制分页附录',b+foot('LEARNING WORKBOOK'),accent='#4d7091',ink='#2e4d69',soft='#edf3f7',extra='.reading{column-count:2;column-gap:23px;font-size:11px}.reading p{margin-bottom:11px}',checks=['参考讨论与延伸练习','验证边界'])

def main():
    build()
    assert len(SAMPLES) == 12
    batch = ROOT / 'tools/validation/artifacts/batch'
    (batch / 'files').mkdir(parents=True,exist_ok=True)
    (batch / 'screenshots').mkdir(exist_ok=True)
    for key,source in SAMPLES.items():
        (ROOT/'testdata/html'/f'{key}.html').write_text(source)
        (batch/'files'/f'{key}.html').write_text(source)
    script='// Generated by scripts/generate-complex-suite.py. Fictional static fixtures.\n'
    script+='const batchSampleKeys='+json.dumps(list(SAMPLES),ensure_ascii=False)+';\n'
    script+='Object.assign(complexSampleInfo,'+json.dumps(INFO,ensure_ascii=False)+');\n'
    script+='Object.assign(complexSamples,'+json.dumps(SAMPLES,ensure_ascii=False)+');\n'
    (ROOT/'tools/playground/internal/server/web/batch-samples.js').write_text(script)
    (batch/'manifest.json').write_text(json.dumps(INFO,ensure_ascii=False,indent=2)+'\n')
    print('Generated 12 distinct complex HTML samples')

if __name__ == '__main__':
    main()

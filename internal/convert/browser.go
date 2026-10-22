package convert

import (
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/xinhaolin/htmlToWord/internal/cdp"
)

//go:embed extract.js
var extractJS string

// PublicResource rejects local/private destinations. Remote resources are opt-in.
func PublicResource(ctx context.Context, raw string) error {
	u, e := url.Parse(raw)
	if e != nil {
		return e
	}
	if u.Scheme != "https" && u.Scheme != "http" {
		return fmt.Errorf("不支持资源协议 %q", u.Scheme)
	}
	if u.User != nil || u.Hostname() == "" {
		return fmt.Errorf("无效资源地址")
	}
	if p := u.Port(); p != "" && p != "80" && p != "443" {
		return fmt.Errorf("只允许 80/443 端口的远程资源")
	}
	ips, e := net.DefaultResolver.LookupIPAddr(ctx, u.Hostname())
	if e != nil {
		return e
	}
	if len(ips) == 0 {
		return fmt.Errorf("资源域名无地址")
	}
	for _, a := range ips {
		ip := a.IP
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() || ip.IsMulticast() || !ip.IsGlobalUnicast() {
			return fmt.Errorf("不允许本机或私网资源")
		}
	}
	return nil
}

func Analyze(ctx context.Context, html, css string, o Options) (Document, error) {
	return AnalyzeWithRuntime(ctx, html, css, o, Runtime{Browser: cdp.Options{Path: os.Getenv("HTMLWORD_CHROME"), NoSandbox: os.Getenv("HTMLWORD_NO_SANDBOX") == "1"}})
}

func AnalyzeWithRuntime(ctx context.Context, html, css string, o Options, execution Runtime) (Document, error) {
	var doc Document
	var browser *cdp.Browser
	var warnings []string
	var resourceDiagnostics []Diagnostic
	var mu sync.Mutex
	var e error
	browser, e = cdp.StartWithOptions(ctx, execution.Browser, func(m cdp.Message) {
		if m.Method != "Fetch.requestPaused" {
			return
		}
		var ev struct {
			RequestID string `json:"requestId"`
			Request   struct {
				URL string `json:"url"`
			} `json:"request"`
		}
		if json.Unmarshal(m.Params, &ev) != nil {
			return
		}
		go func() {
			if execution.LoadResource != nil {
				data, mime, err := execution.LoadResource(ctx, ev.Request.URL)
				if err == nil {
					_ = browser.Call(ctx, m.SessionID, "Fetch.fulfillRequest", map[string]any{"requestId": ev.RequestID, "responseCode": 200, "responseHeaders": []map[string]string{{"name": "Content-Type", "value": mime}}, "body": base64.StdEncoding.EncodeToString(data)}, nil)
				} else {
					mu.Lock()
					if len(warnings) < 200 {
						warnings = append(warnings, "资源未加载："+ev.Request.URL+" ("+err.Error()+")")
						resourceDiagnostics = append(resourceDiagnostics, Diagnostic{Code: "RESOURCE_UNAVAILABLE", Message: err.Error(), Resource: ev.Request.URL})
					}
					mu.Unlock()
					_ = browser.Call(ctx, m.SessionID, "Fetch.failRequest", map[string]any{"requestId": ev.RequestID, "errorReason": "BlockedByClient"}, nil)
				}
				return
			}
			allow := false
			if o.AllowRemote {
				check, cancel := context.WithTimeout(ctx, 3*time.Second)
				allow = PublicResource(check, ev.Request.URL) == nil
				cancel()
			}
			if allow {
				_ = browser.Call(ctx, m.SessionID, "Fetch.continueRequest", map[string]any{"requestId": ev.RequestID}, nil)
			} else {
				mu.Lock()
				if len(warnings) < 25 {
					warnings = append(warnings, "远程资源未加载："+ev.Request.URL)
				}
				mu.Unlock()
				_ = browser.Call(ctx, m.SessionID, "Fetch.failRequest", map[string]any{"requestId": ev.RequestID, "errorReason": "BlockedByClient"}, nil)
			}
		}()
	})
	if e != nil {
		return doc, e
	}
	defer browser.Close()
	var target struct {
		TargetID string `json:"targetId"`
	}
	if e = browser.Call(ctx, "", "Target.createTarget", map[string]any{"url": "about:blank"}, &target); e != nil {
		return doc, e
	}
	var attached struct {
		SessionID string `json:"sessionId"`
	}
	if e = browser.Call(ctx, "", "Target.attachToTarget", map[string]any{"targetId": target.TargetID, "flatten": true}, &attached); e != nil {
		return doc, e
	}
	session := attached.SessionID
	for _, call := range []struct {
		method string
		params any
	}{
		{"Page.enable", map[string]any{}}, {"Emulation.setDeviceMetricsOverride", map[string]any{"width": int(o.ContentPX() + .5), "height": 1100, "deviceScaleFactor": 1, "mobile": false}},
		{"Emulation.setEmulatedMedia", map[string]any{"media": o.Media}},
		{"Fetch.enable", map[string]any{"patterns": []map[string]any{{"urlPattern": "http://*"}, {"urlPattern": "https://*"}, {"urlPattern": "file://*"}}}},
	} {
		if e = browser.Call(ctx, session, call.method, call.params, nil); e != nil {
			return doc, e
		}
	}
	h, _ := json.Marshal(html)
	c, _ := json.Marshal(css)
	f, _ := json.Marshal(o.Font)
	remote := ""
	if o.AllowRemote || execution.LoadResource != nil {
		remote = " http: https:"
	}
	csp := "default-src 'none'; img-src data: blob:" + remote + "; style-src 'unsafe-inline'" + remote + "; font-src data:" + remote + "; script-src 'none'; connect-src 'none'; frame-src 'none'; base-uri 'none'; form-action 'none'"
	policy, _ := json.Marshal(csp)
	// Detached DOMParser does not execute scripts. Strip active elements before
	// writing the document; CSP is placed before all user-provided resources.
	expression := fmt.Sprintf(`(() => {const d=new DOMParser().parseFromString(%s,'text/html');const active=d.querySelectorAll('script,iframe,object,embed,base,meta[http-equiv]');const removed=active.length;active.forEach(x=>x.remove());for(const e of d.querySelectorAll('*'))for(const a of [...e.attributes])if(a.name.toLowerCase().startsWith('on')||a.name.toLowerCase()==='srcdoc')e.removeAttribute(a.name);const policy=d.createElement('meta');policy.httpEquiv='Content-Security-Policy';policy.content=%s;d.head.prepend(policy);const base=d.createElement('style');base.textContent='html{width:%gpx}body{margin:0;font-family:'+%s+';font-size:%gpx}img,svg{max-width:100%%}*,*::before,*::after{animation:none!important;transition:none!important;caret-color:transparent!important}';policy.after(base);const extra=d.createElement('style');extra.textContent=%s;d.head.append(extra);return {html:'<!doctype html>'+d.documentElement.outerHTML,removed};})()`, string(h), string(policy), o.ContentPX(), string(f), o.FontSize*96/72, string(c))
	var clean struct {
		HTML    string `json:"html"`
		Removed int    `json:"removed"`
	}
	if e = browser.Eval(ctx, session, expression, &clean); e != nil {
		return doc, e
	}
	var frame struct {
		FrameTree struct {
			Frame struct {
				ID string `json:"id"`
			} `json:"frame"`
		} `json:"frameTree"`
	}
	if e = browser.Call(ctx, session, "Page.getFrameTree", map[string]any{}, &frame); e != nil {
		return doc, e
	}
	if execution.BaseURL != "" {
		// Install a trusted base before resolving resources in a second document.
		// User supplied base elements were removed by the sanitizer above.
		// Keep standards mode: outerHTML alone drops the doctype and changes table metrics.
		base, _ := json.Marshal(execution.BaseURL)
		if e = browser.Eval(ctx, session, fmt.Sprintf(`(() => {const d=new DOMParser().parseFromString(%s,'text/html');d.querySelector('meta[http-equiv="Content-Security-Policy"]').content=d.querySelector('meta[http-equiv="Content-Security-Policy"]').content.replace("base-uri 'none'","base-uri http: https:");const b=d.createElement('base');b.href=%s;d.head.prepend(b);return '<!doctype html>'+d.documentElement.outerHTML;})()`, mustJSON(clean.HTML), string(base)), &clean.HTML); e != nil {
			return doc, e
		}
	}
	if e = browser.Call(ctx, session, "Page.setDocumentContent", map[string]any{"frameId": frame.FrameTree.Frame.ID, "html": clean.HTML}, nil); e != nil {
		return doc, e
	}
	settings, _ := json.Marshal(map[string]any{"maxNodes": execution.MaxNodes, "maxDepth": execution.MaxDepth, "maxImages": execution.MaxImages, "footerMode": execution.FooterMode, "symbolCheckboxes": execution.SymbolCheckboxes, "explicitSemantics": execution.ExplicitSemantics})
	if e = browser.Eval(ctx, session, `globalThis.__htmlwordPolicy=`+string(settings), nil); e != nil {
		return doc, e
	}
	// fonts.ready and readyState can settle before setDocumentContent starts a
	// linked stylesheet. Arm listeners before reconnecting unresolved links, then
	// settle fonts/images. The resolver cache coalesces duplicate resource loads.
	if e = browser.Eval(ctx, session, `Promise.race([
  (async () => {
    await Promise.all([...document.querySelectorAll('link[rel~="stylesheet"]')].map(link => {
      if (link.sheet) return Promise.resolve();
      return new Promise(resolve => {
        const pending = link.cloneNode(true);
        pending.addEventListener('load', resolve, {once: true});
        pending.addEventListener('error', resolve, {once: true});
        link.replaceWith(pending);
      });
    }));
    await Promise.all([document.fonts.ready, ...[...document.images].map(i => i.decode().catch(() => {}))]);
    return true;
  })(),
  new Promise(resolve => setTimeout(() => resolve(false), 8000))
])`, nil); e != nil {
		return doc, e
	}
	if e = browser.Eval(ctx, session, extractJS, &doc); e != nil {
		return doc, e
	}
	mu.Lock()
	doc.Warnings = append(doc.Warnings, warnings...)
	doc.Diagnostics = append(doc.Diagnostics, resourceDiagnostics...)
	mu.Unlock()
	if clean.Removed > 0 {
		doc.Warnings = append(doc.Warnings, "脚本、嵌入网页及活动内容已移除，仅转换静态 HTML。")
		doc.Diagnostics = append(doc.Diagnostics, Diagnostic{Code: "ACTIVE_CONTENT_REMOVED"})
	}
	if len(doc.Blocks) == 0 {
		return doc, fmt.Errorf("HTML 中没有可转换的可见内容")
	}
	seen := map[string]bool{}
	unique := doc.Warnings[:0]
	for _, w := range doc.Warnings {
		w = strings.TrimSpace(w)
		if !seen[w] {
			seen[w] = true
			unique = append(unique, w)
		}
	}
	doc.Warnings = unique
	return doc, nil
}

func mustJSON(value string) string { raw, _ := json.Marshal(value); return string(raw) }

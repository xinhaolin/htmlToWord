# Publishing kit

Primary URL: https://xinhaolin.github.io/htmlToWord/
Repository: https://github.com/xinhaolin/htmlToWord
Release: https://github.com/xinhaolin/htmlToWord/releases/tag/v0.1.0

## Chinese developer-community post

Title: 用 Go 把复杂 HTML 转成可编辑 Word：原生图表、复选框和自动分页

正文见 [中文版技术文章](../content/article-zh.md)。开头附仓库、文档和 24 秒真实编辑画面的剪辑演示；保留 Chromium 依赖、图表显式数据和客户端验证边界。结尾邀请使用者提供最小 HTML 及办公软件版本复测，反馈转换问题。

## Go community post

Title: HTMLWord: a Go SDK for static HTML/CSS to editable DOCX

I built HTMLWord to export static HTML/CSS into editable, flowing Word documents. Chromium measures layout; Go writes native paragraphs, tables, annotated Office charts with embedded XLSX data, and checkbox content controls.

It requires separately installed Chrome/Chromium and currently supports macOS and Linux. It does not claim arbitrary HTML/Word pixel parity. There are 12 complex synthetic examples, input/output downloads and an edited sequence of actual WPS editing captures. Full chart Edit Data interaction and Microsoft Word acceptance remain to be verified.

Source: https://github.com/xinhaolin/htmlToWord
Guide and examples: https://xinhaolin.github.io/htmlToWord/

I'd appreciate small reproduction cases and feedback from the office clients you use.

## Show HN

Title: Show HN: HTMLWord – Convert HTML to editable DOCX in Go

URL: https://github.com/xinhaolin/htmlToWord

First comment: use the Go community post above, plus the direct release download and reproducible CLI example. Show HN requires something people can actually try. Do not submit just the article or a sign-up page. Follow the community's current rules and answer technical questions directly.

## Search and sharing

Use the product name and the task together: HTML to DOCX in Go; HTML to Word Go SDK; editable Word export; native Office charts and checkbox controls. Write descriptive, distinct titles and link the guides from actual examples. The project Pages URL is a subpath: docs/robots.txt is not the origin-level robots.txt for xinhaolin.github.io. A custom domain and Search Console verification require owner-controlled DNS/account access.

## First-week measurement

Record each publication URL and date, repository unique visitors, net new stars, identifiable trial feedback, bug reports and repeat users. Leave unavailable metrics unknown. Stars, issue activity and package reverse-dependencies are different signals; pkg.go.dev is not an SDK download counter.

Suggested review: day 1 repository/docs/release; day 2 Chinese article; day 3 Go community; day 4 Show HN if ready; days 5–7 reply, fix reproducible issues and compare referral quality. This is a review cadence, not a scheduled automation.

Awesome Go is a later submission target: check its current maturity, version and test/documentation requirements before opening a PR.

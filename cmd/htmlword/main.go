// Command htmlword converts one static HTML file using the public SDK.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/xinhaolin/htmlToWord"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	input := flag.String("input", "", "HTML file (required)")
	output := flag.String("output", "document.docx", "DOCX destination")
	css := flag.String("css", "", "additional CSS file")
	browser := flag.String("chrome", "", "Chrome/Chromium executable")
	assets := flag.String("assets", "", "explicit asset directory (caller controls symlinks and permissions)")
	base := flag.String("base-url", "https://assets.invalid/", "base URL for relative assets")
	strict := flag.Bool("strict", false, "reject warning diagnostics")
	autoFooter := flag.Bool("auto-footer", false, "promote short HTML footer to repeated Word footer")
	symbols := flag.Bool("symbol-checkboxes", false, "convert leading square symbols into checkbox controls")
	deterministic := flag.Bool("deterministic", false, "fix package creation time")
	remote := flag.Bool("allow-remote", false, "allow public HTTP(S) resources")
	page := flag.String("page", "A4", "A4, A3 or Letter")
	margin := flag.Float64("margin", 20, "page margin in millimetres")
	landscape := flag.Bool("landscape", false, "landscape page")
	flag.Parse()
	if *input == "" {
		flag.Usage()
		return fmt.Errorf("-input is required")
	}
	html, err := os.ReadFile(*input)
	if err != nil {
		return err
	}
	var extra []byte
	if *css != "" {
		extra, err = os.ReadFile(*css)
		if err != nil {
			return err
		}
	}
	config := htmlword.Config{BrowserPath: *browser}
	if *assets != "" && *remote {
		return fmt.Errorf("choose either -assets or -allow-remote")
	}
	if *assets != "" {
		config.Resolver = htmlword.FSResolver(os.DirFS(*assets))
	}
	if *remote {
		config.Resolver = htmlword.HTTPResolver{}
	}
	converter, err := htmlword.New(config)
	if err != nil {
		return err
	}
	defer converter.Close()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	file, err := os.CreateTemp(filepath.Dir(*output), ".htmlword-*.docx")
	if err != nil {
		return err
	}
	temp := file.Name()
	defer os.Remove(temp)
	defer file.Close()
	mode := htmlword.FooterPreserve
	if *autoFooter {
		mode = htmlword.FooterAuto
	}
	report, err := converter.Convert(ctx, file, htmlword.Input{HTML: string(html), CSS: string(extra), BaseURL: *base}, htmlword.Options{Page: &htmlword.PageSetup{Size: htmlword.PageSize(*page), MarginMM: *margin, Landscape: *landscape}, Strict: *strict, FooterMode: mode, SymbolCheckboxes: *symbols, Deterministic: *deterministic})
	if err != nil {
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	if err = os.Rename(temp, *output); err != nil {
		return err
	}
	return json.NewEncoder(os.Stderr).Encode(report)
}

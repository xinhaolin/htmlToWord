// Run from the repository root: go run ./examples/basic
package main

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/xinhaolin/htmlToWord"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	converter, err := htmlword.New(htmlword.Config{})
	if err != nil {
		return err
	}
	defer converter.Close()
	var output bytes.Buffer
	report, err := converter.Convert(context.Background(), &output, htmlword.Input{
		HTML:  `<h1>Project notes</h1><p>Editable text that flows across pages.</p><p><input type="checkbox" checked> Reviewed</p>`,
		Title: "Project notes",
	}, htmlword.Options{})
	if err != nil {
		return err
	}
	if err := os.WriteFile("document.docx", output.Bytes(), 0644); err != nil {
		return err
	}
	fmt.Printf("Wrote %d bytes; %d checkbox; %d diagnostics\n", report.Bytes, report.Stats.Checkboxes, len(report.Diagnostics))
	return nil
}

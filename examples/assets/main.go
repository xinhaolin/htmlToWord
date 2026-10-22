// This example serves CSS from an in-memory filesystem, without network access.
package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing/fstest"

	"github.com/xinhaolin/htmlToWord"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run() error {
	assets := fstest.MapFS{"style.css": &fstest.MapFile{Data: []byte(`h1{color:#365e72}p{font-size:14px}`)}}
	converter, err := htmlword.New(htmlword.Config{Resolver: htmlword.FSResolver(assets)})
	if err != nil {
		return err
	}
	defer converter.Close()
	var out bytes.Buffer
	_, err = converter.Convert(context.Background(), &out, htmlword.Input{
		HTML:    `<link rel="stylesheet" href="style.css"><h1>Local assets</h1><p>No external requests are needed.</p>`,
		BaseURL: "https://assets.invalid/",
	}, htmlword.Options{})
	if err != nil {
		return err
	}
	return os.WriteFile("assets.docx", out.Bytes(), 0644)
}

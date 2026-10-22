package htmlword_test

import (
	"bytes"
	"context"
	"github.com/xinhaolin/htmlToWord"
)

func ExampleConverter_Convert() {
	converter, err := htmlword.New(htmlword.Config{})
	if err != nil {
		return
	}
	defer converter.Close()
	var output bytes.Buffer
	report, err := converter.Convert(context.Background(), &output, htmlword.Input{HTML: "<h1>Project report</h1><p>Editable content.</p>"}, htmlword.Options{})
	_, _ = report, err
}

package convert

import (
	"context"
	"github.com/xinhaolin/htmlToWord/internal/cdp"
	"time"
)

// Runtime is internal, per-conversion execution policy. It is never serialized.
type Runtime struct {
	Browser                       cdp.Options
	BaseURL                       string
	LoadResource                  func(context.Context, string) ([]byte, string, error)
	MaxNodes, MaxDepth, MaxImages int
	MaxOutputBytes                int64
	CreatedAt                     *time.Time
	FooterMode                    string
	SymbolCheckboxes              bool
	ExplicitSemantics             bool
}

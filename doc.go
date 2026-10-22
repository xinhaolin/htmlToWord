// Package htmlword converts static HTML and CSS into editable, flowing DOCX.
//
// Chromium computes the input layout; Go writes native Word paragraphs, tables,
// charts and checkbox controls. No HTTP server or office application is needed
// for conversion. External resources are disabled unless Config.Resolver is set.
// Callers must close converters they create. See README.md for runtime support.
package htmlword

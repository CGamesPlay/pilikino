package parser

import (
	mathjax "github.com/litao91/goldmark-mathjax"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

func Parse(input []byte) (ast.Node, error) {
	markdown := goldmark.New(
		goldmark.WithExtensions(mathjax.MathJax, extension.Table),
	)
	context := parser.NewContext()
	reader := text.NewReader(input)
	doc := markdown.Parser().Parse(reader, parser.WithContext(context))
	return doc, nil
}

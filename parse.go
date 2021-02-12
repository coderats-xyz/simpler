package simpler

import (
	"fmt"

	"github.com/alecthomas/participle"
)

type metaData struct {
	Key   string `"-" "-" @Ident ":"`
	Value string `@Ident`
}

type metaParser struct {
	parser *participle.Parser
}

func newMetaParser() (*metaParser, error) {
	parser, err := participle.Build(&metaData{})
	if err != nil {
		return nil, fmt.Errorf("Error creating a parser: %v", err)
	}

	return &metaParser{parser}, nil
}

func (m *metaParser) parseMeta(input string) (*metaData, bool, error) {
	ast := &metaData{}
	err := m.parser.ParseString(input, ast)
	if err != nil {
		return nil, false, fmt.Errorf(`Error parsing string "%s": %v`, input, err)
	}

	return ast, true, nil
}

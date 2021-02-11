package simpler

import (
	"fmt"

	"github.com/alecthomas/participle"
)

type MetaData struct {
	Key   string `"-" "-" @Ident ":"`
	Value string `@Ident`
}

func parseMeta(input string) (*MetaData, bool, error) {
	parser, err := participle.Build(&MetaData{})
	if err != nil {
		return nil, false, fmt.Errorf("Error creating a parser: %v", err)
	}

	ast := &MetaData{}
	err = parser.ParseString(input, ast)
	if err != nil {
		return nil, false, fmt.Errorf(`Error parsing string "%s": %v`, input, err)
	}

	return ast, true, nil
}

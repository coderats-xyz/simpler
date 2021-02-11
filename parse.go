package simpler

import (
	"fmt"

	"github.com/alecthomas/participle"
)

type metaData struct {
	Key   string `"-" "-" @Ident ":"`
	Value string `@Ident`
}

func parseMeta(input string) (*metaData, bool, error) {
	parser, err := participle.Build(&metaData{})
	if err != nil {
		return nil, false, fmt.Errorf("Error creating a parser: %v", err)
	}

	ast := &metaData{}
	err = parser.ParseString(input, ast)
	if err != nil {
		return nil, false, fmt.Errorf(`Error parsing string "%s": %v`, input, err)
	}

	return ast, true, nil
}

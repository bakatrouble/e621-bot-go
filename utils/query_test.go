package utils

import (
	"testing"

	"github.com/alecthomas/repr"
)

func TestParser(t *testing.T) {
	//repr.Println(QueryParser)
	expr, err := QueryParser.ParseString("", "-a b {c | d} -{e | f} rating:s")
	if err != nil {
		panic(err)
	}

	repr.Println(expr)
}

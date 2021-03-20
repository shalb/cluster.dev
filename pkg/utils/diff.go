package utils

import (
	"fmt"
	"strings"

	"github.com/kylelemons/godebug/pretty"
	"github.com/shalb/cluster.dev/pkg/colors"
)

type emptyStruct struct{}

func Diff(structA, structB interface{}, colored bool) string {
	if structA == nil {
		structA = emptyStruct{}
	}
	if structB == nil {
		structB = emptyStruct{}
	}
	GreenColor := "%s"
	RedColor := "%s"
	if colored {
		GreenColor = colors.Green.Sprint("%s")
		RedColor = colors.Red.Sprint("%s")
	}
	diffs := make([]string, 0)
	// Compare, join result to string and add colors.
	for _, s := range strings.Split(pretty.Compare(structA, structB), "\n") {
		switch {
		case strings.HasPrefix(s, "+"):
			diffs = append(diffs, fmt.Sprintf(GreenColor, s))
		case strings.HasPrefix(s, "-"):
			diffs = append(diffs, fmt.Sprintf(RedColor, s))
		default:
			diffs = append(diffs, s)
		}
	}
	return strings.Join(diffs, "\n")
}

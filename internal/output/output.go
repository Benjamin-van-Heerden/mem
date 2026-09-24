package output

import (
	"fmt"
	"io"
	"strings"
)

func Heading(w io.Writer, title string) {
	fmt.Fprintf(w, "\n%s\n%s\n%s\n\n", strings.Repeat("=", 80), title, strings.Repeat("=", 80))
}

func Section(w io.Writer, title string) {
	fmt.Fprintf(w, "\n%s\n%s\n%s\n\n", strings.Repeat("-", 70), title, strings.Repeat("-", 70))
}

func File(w io.Writer, title string) {
	fmt.Fprintf(w, "\n%s\n# %s\n%s\n\n", strings.Repeat("#", 50), title, strings.Repeat("#", 50))
}

func Instruction(w io.Writer, lines ...string) {
	Section(w, "⚠️ AGENT INSTRUCTION")
	for _, line := range lines {
		fmt.Fprintln(w, line)
	}
}

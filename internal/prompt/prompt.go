package prompt

import (
	"fmt"
	"strings"
)

const systemPlain = `You write POSIX awk programs for text streams. Rules:
- Output ONLY the awk program text. No markdown fences, no prose before or after.
- Assume input is read from stdin line by line (standard awk record mode).
- Use $1, $2, ... for fields. NF is the number of fields on the current line.
- Prefer concise, correct awk. Use BEGIN/END when needed for aggregates.
- If the user asks for grouping or sums, use associative arrays.
- Print results with print or printf as appropriate.`

const systemExplain = `You write POSIX awk programs for text streams. Rules:
- Output ONLY the awk program, but every non-empty line of awk must be followed
  immediately by one comment line starting with # that briefly explains that line.
- No markdown fences, no prose before or after the program.
- Same awk semantics as a normal program: the executable awk lines must form valid awk
  when comment lines starting with # are included (awk treats # as comment to EOL).`

// Build returns system and user messages for the LLM.
func Build(naturalLanguage, fieldSep string, explain bool, correction string) (system, user string) {
	if explain {
		system = systemExplain
	} else {
		system = systemPlain
	}
	var b strings.Builder
	b.WriteString("Task (plain English):\n")
	b.WriteString(strings.TrimSpace(naturalLanguage))
	b.WriteString("\n")
	if fieldSep != "" {
		fmt.Fprintf(&b, "Field separator (-F): %q (escaped for clarity)\n", fieldSep)
	} else {
		b.WriteString("Field separator: whitespace (default awk)\n")
	}
	if correction != "" {
		b.WriteString("\nYour previous awk failed validation or execution. Fix it.\n")
		b.WriteString(correction)
		b.WriteString("\n")
	}
	return system, b.String()
}

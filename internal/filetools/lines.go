package filetools

import "bytes"

// rawLine is one physical line of a file, split so that every byte of the
// original content can be reconstructed exactly: text holds the line
// content without its terminator, term holds the terminator that followed
// it ("\n", "\r\n", or "" for a final line with no trailing newline).
type rawLine struct {
	text string
	term string
}

// splitLines splits data into rawLines without losing any byte: joinLines
// on the result always reproduces data exactly.
func splitLines(data []byte) []rawLine {
	if len(data) == 0 {
		return nil
	}
	var lines []rawLine
	start := 0
	for i := 0; i < len(data); i++ {
		if data[i] != '\n' {
			continue
		}
		end := i
		term := "\n"
		if end > start && data[end-1] == '\r' {
			end--
			term = "\r\n"
		}
		lines = append(lines, rawLine{text: string(data[start:end]), term: term})
		start = i + 1
	}
	if start < len(data) {
		lines = append(lines, rawLine{text: string(data[start:]), term: ""})
	}
	return lines
}

func joinLines(lines []rawLine) []byte {
	var buf bytes.Buffer
	for _, l := range lines {
		buf.WriteString(l.text)
		buf.WriteString(l.term)
	}
	return buf.Bytes()
}

// fileTerminator returns the dominant line terminator used by lines,
// falling back to "\n" for a file with no terminated line at all (empty or
// a single line without a trailing newline).
func fileTerminator(lines []rawLine) string {
	for _, l := range lines {
		if l.term != "" {
			return l.term
		}
	}
	return "\n"
}

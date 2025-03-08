package bruparser

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/rs/zerolog/log"
)

func getCleanedLines(reader io.Reader) ([]string, error) {
	// Read the file line by line and collect meaningful lines.
	lines := make([]string, 0)
	scanner := bufio.NewReader(reader)
	for {
		line, err := scanner.ReadString('\n')
		if err == io.EOF {
			log.Trace().Msg("EOF")
			break
		}

		if err != nil {
			return nil, fmt.Errorf("error reading file: %w", err)
		}

		// Skip empty lines, comments and annotations
		if isEmptyLine(line) || isComment(line) || isAnnotation(line) {
			continue
		}

		// Parse the line
		lines = append(lines, line)
	}
	return lines, nil
}

func isEmptyLine(line string) bool {
	return len(strings.TrimSpace(line)) == 0
}

// Comments are lines starting with "#"
// Ref: https://github.com/brulang/bru-lang?tab=readme-ov-file#comments
func isComment(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "#")
}

// Annotations are used to provide additional information about a key-value pair. An annotation starts with (@) and ends with a newline (\n).
// Ref: https://github.com/brulang/bru-lang?tab=readme-ov-file#annotations
func isAnnotation(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "@")
}

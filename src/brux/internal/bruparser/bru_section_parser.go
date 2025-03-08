package bruparser

import (
	"errors"
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

type _State string

const (
	_sectionStart   _State = "sectionStart"
	_sectionRunning _State = "sectionRunning"
)

// Section is a struct that represents a Bru section
// Example:
//
//	get {
//	 url: https://example.com
//	 body: json
//	 auth: none
//	}
type _Section struct {
	sectionName   string
	sectionValues map[string]string
	sectionData   string
}

var (
	ErrInvalidSectionStart = errors.New("invalid section start")
	ErrInvalidKeyValuePair = errors.New("invalid key value pair")
)

func getSections(lines []string) ([]_Section, error) {
	sectionList := make([]_Section, 0)
	nextState := _sectionStart
	currentSectionName := ""
	currentSectionData := ""
	currentSectionValues := make(map[string]string)
	for _, line := range lines {
		switch nextState {
		case _sectionStart:
			if !strings.HasSuffix(strings.TrimSpace(line), "{") {
				return nil, fmt.Errorf("%w: '%s'", ErrInvalidSectionStart, line)
			}
			currentSectionName = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(line), "{"))
			nextState = _sectionRunning
		case _sectionRunning:
			if strings.TrimSpace(line) == "}" && strings.HasPrefix(line, "}") {
				nextState = _sectionStart
				sectionList = append(sectionList, _Section{
					sectionName:   currentSectionName,
					sectionValues: currentSectionValues,
					sectionData:   currentSectionData,
				})
				currentSectionName = ""
				currentSectionData = ""
				currentSectionValues = make(map[string]string)
			} else if strings.TrimSpace(line) == "{" || currentSectionData != "" {
				currentSectionData += line + "\n"
			} else {
				// parse key value pair
				keyValue := strings.SplitN(line, ":", 2)
				if len(keyValue) < 2 {
					return nil, fmt.Errorf("invalid key value pair: '%s': %w", line, ErrInvalidKeyValuePair)
				}
				currentSectionValues[strings.TrimSpace(keyValue[0])] = strings.TrimSpace(keyValue[1])
			}
		}

		log.Trace().
			Str("line", line).
			Str("section", currentSectionName).
			Str("state", string(nextState)).
			Int("values", len(currentSectionValues)).
			Msg("section")
	}
	return sectionList, nil
}

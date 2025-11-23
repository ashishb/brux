package bruparser

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/samber/lo"
)

// BruFile is a struct that represents a Bru file
type BruFile struct {
	meta     *_Meta
	req      *_Request
	headers  map[string]string
	bodyJson *string

	// This section is present in the "env" files
	vars map[string]string
}

type _Meta struct {
	name    string
	reqType string // only "http" for now
	seq     string
}

type _Request struct {
	httpMethod string // only "get" for now
	url        string
	body       string // only "json" for now
	auth       string // only "none" for now
}

var (
	ErrUnknownSectionName            = errors.New("unknown section name")
	ErrUnsupportedNetworkRequestType = errors.New("unsupported request type")
	ErrTemplateVariablesFound        = errors.New("template variables found")
)

// NewBruFile creates a new BruFile object from the given reader
func NewBruFile(reader io.Reader) (*BruFile, error) {
	lines, err := getCleanedLines(reader)
	if err != nil {
		return nil, err
	}

	var metaSection *_Meta
	var reqSection *_Request
	headers := make(map[string]string)
	var bodyJson *string
	vars := make(map[string]string)
	sections, err := getSections(lines)
	if err != nil {
		return nil, fmt.Errorf("error getting sections: %w", err)
	}

	for _, section := range sections {
		log.Debug().
			Any("section", section.sectionName).
			Any("values", section.sectionValues).
			Msg("section")
		switch section.sectionName {
		case "meta":
			metaSection = &_Meta{
				name:    section.sectionValues["name"],
				reqType: section.sectionValues["type"],
				seq:     section.sectionValues["seq"],
			}
			if metaSection.reqType != "http" {
				return nil, fmt.Errorf("%w: '%s'", ErrUnsupportedNetworkRequestType, metaSection.reqType)
			}
		case "get", "head", "post":
			urlStr := section.sectionValues["url"]
			reqSection = &_Request{
				httpMethod: section.sectionName,
				body:       section.sectionValues["body"],
				auth:       section.sectionValues["auth"],
				url:        urlStr,
			}
		case "headers":
			for k, v := range section.sectionValues {
				headers[k] = v
			}
		case "vars":
			for k, v := range section.sectionValues {
				vars[k] = v
			}
		case "body:json":
			bodyJson = &section.sectionData
		default:
			return nil, fmt.Errorf("%w: '%s'", ErrUnknownSectionName, section.sectionName)
		}
	}

	return &BruFile{
		meta:     metaSection,
		req:      reqSection,
		headers:  headers,
		bodyJson: bodyJson,
		vars:     vars,
	}, nil
}

func (f BruFile) HttpMethod() string {
	return strings.ToUpper(f.req.httpMethod)
}

func (f BruFile) URL() (*string, error) {
	u1 := replaceVariables(f.req.url, f.vars)
	if hasUnreplacedVariables(u1) {
		return nil, fmt.Errorf("%w: '%s'", ErrTemplateVariablesFound, u1)
	}
	return lo.ToPtr(u1), nil
}

func (f BruFile) RequestBody() (io.Reader, error) {
	if f.req.body == "json" && f.bodyJson != nil {
		newJson := replaceVariables(*f.bodyJson, f.vars)
		if hasUnreplacedVariables(newJson) {
			return nil, fmt.Errorf("%w: '%s'", ErrTemplateVariablesFound, newJson)
		}
		return bytes.NewReader([]byte(newJson)), nil
	}

	return nil, nil
}

func (f BruFile) Headers() (http.Header, error) {
	h := make(http.Header)
	for k, v := range f.headers {
		k1 := replaceVariables(k, f.vars)
		v1 := replaceVariables(v, f.vars)
		if hasUnreplacedVariables(k1) {
			return nil, fmt.Errorf("%w: '%s'", ErrTemplateVariablesFound, k1)
		}
		if hasUnreplacedVariables(v1) {
			return nil, fmt.Errorf("%w: '%s'", ErrTemplateVariablesFound, v1)
		}
		h.Set(k1, v1)
	}
	return h, nil
}

func (f BruFile) Variables() map[string]string {
	return f.vars
}

func (f BruFile) SetVariables(variables map[string]string) {
	if f.vars == nil {
		f.vars = make(map[string]string)
	}

	for k, v := range variables {
		f.vars[k] = v
	}

	log.Debug().
		Int("variables", len(f.vars)).
		Msg("variables set")
}

func replaceVariables(str string, vars map[string]string) string {
	if len(vars) == 0 {
		log.Debug().
			Str("str", str).
			Msg("no variables to replace")
		return str
	}

	for k, v := range vars {
		log.Trace().
			Str("key", k).
			Str("value", v).
			Msg("replacing variable")
		s2 := strings.ReplaceAll(str, "{{"+k+"}}", v)
		s2 = strings.ReplaceAll(s2, "{{process.env."+k+"}}", v)
		if s2 != str {
			log.Debug().
				Str("key", k).
				Msg("replaced variable")
			str = s2
		}
	}
	return str
}

func hasUnreplacedVariables(str string) bool {
	return strings.Contains(str, "{{")
}

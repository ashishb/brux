package brurunner

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/ashishb/brux/src/brux/internal/bruparser"
)

func Run(ctx context.Context, cfg Config) error {
	bruFile, err := cfg.getBruFile()
	if err != nil {
		return fmt.Errorf("could not get bru file: %w", err)
	}

	return run(ctx, cfg, bruFile)
}

func run(ctx context.Context, cfg Config, bruObj *bruparser.BruFile) error {
	u1, err := bruObj.URL()
	if err != nil {
		return fmt.Errorf("could not get URL: %w", err)
	}
	reqBody, err := bruObj.RequestBody()
	if err != nil {
		return fmt.Errorf("could not get request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, bruObj.HttpMethod(), *u1, reqBody)
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	log.Debug().
		Str("request", req.URL.String()).
		Str("method", req.Method).
		Msg("requesting")

	req.Header, err = bruObj.Headers()
	if err != nil {
		return fmt.Errorf("could not get headers: %w", err)
	}

	client := &http.Client{
		Timeout: 300 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("could not make request: %w", err)
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("could not read response body: %w", err)
	}

	log.Debug().
		Int("response", len(data)).
		Msg("response received")
	return cfg.maybeSaveOutput(data)
}

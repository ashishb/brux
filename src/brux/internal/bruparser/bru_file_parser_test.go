package bruparser

import (
	"bytes"
	_ "embed"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ashishb/brux/src/brux/internal/logger"
)

//go:embed testdata/simple_get.bru
var _simpleGet []byte

func TestNewBruFile1(t *testing.T) {
	t.Parallel()
	logger.ConfigureLogging()
	bruFile, err := NewBruFile(bytes.NewReader(_simpleGet))
	if err != nil {
		t.Errorf("NewBruFile() error = %v", err)
	}

	require.NotNil(t, bruFile.meta)
	require.NotNil(t, bruFile.req)
	require.NotEmpty(t, bruFile.headers)

	require.Equal(t, "Send request to example.com", bruFile.meta.name)
	require.Equal(t, "http", bruFile.meta.reqType)
	require.Equal(t, "1", bruFile.meta.seq)

	require.Equal(t, "get", bruFile.req.httpMethod)
	require.Equal(t, "http://example.com/", bruFile.req.url)
	require.Equal(t, "json", bruFile.req.body)
	require.Equal(t, "none", bruFile.req.auth)
	require.Equal(t, "application/json", bruFile.headers["Content-Type"])
}

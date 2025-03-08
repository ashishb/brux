package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	_defaultLogLevel = zerolog.InfoLevel
)

var _initLogging sync.Once

func init() {
	ConfigureLogging()
}

// ConfigureLogging configures ZeroLog's logging config with good defaults
func ConfigureLogging() {
	_initLogging.Do(configureLogging)
}

func configureLogging() {
	colorLogOutput := !strings.EqualFold(os.Getenv("COLOR_LOG_OUTPUT"), "false")
	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logLevel := getLogLevel()
	zerolog.SetGlobalLevel(logLevel)

	if colorLogOutput {
		// Pretty printing is a bit inefficient for production
		output := zerolog.ConsoleWriter{Out: os.Stderr}
		output.FormatTimestamp = func(t any) string {
			number, ok := t.(json.Number)
			if !ok {
				panic(fmt.Sprintf("Expected type %s", t))
			}

			ms, err := number.Int64()
			if err != nil {
				panic(err)
			}

			return time.Unix(ms, 0).In(time.Local).Format("03:04:05PM")
		}
		log.Logger = log.Output(output)
		log.Logger = log.With().Caller().Logger()
	}

	zerolog.TimestampFunc = func() time.Time {
		return time.Now().In(time.Local)
	}
	zerolog.CallerMarshalFunc = func(_ uintptr, file string, line int) string {
		// Use just the filename and not the full file path for logging
		fields := strings.Split(file, "/")
		return fields[len(fields)-1] + ":" + strconv.Itoa(line)
	}

	log.Info().
		Stringer("logLevel", logLevel).
		Bool("colorLogOutput", colorLogOutput).
		Msg("Configured logging")
}

func getLogLevel() zerolog.Level {
	logLevelStr := strings.TrimSpace(os.Getenv("LOG_LEVEL"))
	if len(logLevelStr) == 0 {
		return _defaultLogLevel
	}
	switch strings.ToUpper(logLevelStr) {
	case "TRACE":
		return zerolog.TraceLevel
	case "DEBUG":
		return zerolog.DebugLevel
	case "INFO":
		return zerolog.InfoLevel
	case "ERROR":
		return zerolog.ErrorLevel
	case "WARN":
		return zerolog.WarnLevel
	case "FATAL":
		return zerolog.FatalLevel
	default:
		panic("Unexpected log level: " + logLevelStr)
	}
}

package brurunner

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/hashicorp/go-envparse"
	"github.com/rs/zerolog/log"

	"github.com/ashishb/brux/src/brux/internal/bruparser"
)

const _BrunoEnvironmentsDirName = "environments"

type Config struct {
	bruFilePath     string
	environmentName string

	saveOutput     bool
	outputFilePath string

	prettyPrint bool
}

var ErrEmptyBruFilePath = errors.New("empty bru file path")

func NewConfig(bruFilePath string, saveOutput bool, outputFilePath string, envName string, prettyPrint bool) (*Config, error) {
	if bruFilePath == "" {
		return nil, ErrEmptyBruFilePath
	}
	if !fileExists(bruFilePath) {
		return nil, fmt.Errorf("file does not exist '%s': %w", bruFilePath, os.ErrNotExist)
	}

	return &Config{
		bruFilePath:     bruFilePath,
		saveOutput:      saveOutput,
		outputFilePath:  outputFilePath,
		environmentName: envName,
		prettyPrint:     prettyPrint,
	}, nil
}

func (cfg Config) getBruFile() (*bruparser.BruFile, error) {
	if !fileExists(cfg.bruFilePath) {
		return nil, fmt.Errorf("file does not exist: %s, %w", cfg.bruFilePath, os.ErrNotExist)
	}

	f, err := os.Open(cfg.bruFilePath)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %w", err)
	}

	defer f.Close()
	bruFile, err := bruparser.NewBruFile(f)
	if err != nil {
		return nil, fmt.Errorf("could not parse file: %w", err)
	}

	variables, err := cfg.getVariables()
	if err != nil {
		return nil, fmt.Errorf("could not get variables: %w", err)
	}

	bruFile.SetVariables(variables)
	log.Info().
		Str("file", cfg.bruFilePath).
		Any("bruFile", bruFile).
		Msg("file parsed successfully")
	return bruFile, nil
}

func (cfg Config) maybeSaveOutput(data []byte) error {
	if !cfg.saveOutput {
		return nil
	}

	if cfg.outputFilePath == "" {
		fileName := strings.TrimSuffix(path.Base(cfg.bruFilePath), path.Ext(cfg.bruFilePath))
		fileName = regexp.MustCompile(`[^a-zA-Z0-9-]+`).ReplaceAllString(fileName, "-")
		fileName = strings.ReplaceAll(fileName, "--", "-")
		fileName = strings.Trim(fileName, "-")
		extension := mimetype.Detect(data).Extension()
		suffix := getHash(data) + extension
		cfg.outputFilePath = path.Join(
			os.TempDir(), strings.Join([]string{"bru", "output", fileName, cfg.environmentName, suffix}, "-"))
	}
	if cfg.prettyPrint {
		data = maybePrettyPrint(data)
	}
	if err := os.WriteFile(cfg.outputFilePath, data, 0o600); err != nil {
		return fmt.Errorf("could not write file: %w", err)
	}

	log.Info().
		Str("outputFilePath", cfg.outputFilePath).
		Msg("output saved")
	return nil
}

func (cfg Config) getVariables() (map[string]string, error) {
	var1, err := cfg.getVariablesFromBruEnvironment()
	if err != nil {
		return nil, err
	}

	var2, err := cfg.getVariablesFromEnvFile()
	if err != nil {
		return nil, err
	}

	variables := make(map[string]string, len(var1)+len(var2))
	for k, v := range var1 {
		variables[k] = v
	}
	for k, v := range var2 {
		variables[k] = v
	}
	return variables, nil
}

func (cfg Config) getVariablesFromBruEnvironment() (map[string]string, error) {
	if cfg.environmentName == "" {
		return make(map[string]string), nil
	}

	parentDir := path.Dir(cfg.bruFilePath)
	for {
		envDir := path.Join(parentDir, _BrunoEnvironmentsDirName)
		if !dirExists(envDir) {
			parentDir = path.Dir(parentDir)
			if parentDir == "/" {
				break
			}
			continue
		}
		file := path.Join(envDir, cfg.environmentName+".bru")
		if fileExists(file) {
			return getVariablesFromFile(file)
		}

		if isBrunoCollectionRootDir(parentDir) {
			log.Warn().
				Str("dir", parentDir).
				Str("envName", cfg.environmentName).
				Msg("reached top of bruno collection dir but 'environments' dir not found")
			break
		}
	}

	return make(map[string]string), nil
}

func (cfg Config) getVariablesFromEnvFile() (map[string]string, error) {
	parentDir := path.Dir(cfg.bruFilePath)
	for {
		log.Debug().
			Str("dir", parentDir).
			Str("envName", cfg.environmentName).
			Msg("searching for '.env' file")
		envFile := path.Join(parentDir, ".env")
		if fileExists(envFile) {
			return getVariablesFromEnvFile(envFile)
		}
		if isBrunoCollectionRootDir(parentDir) {
			log.Info().
				Str("envDir", parentDir).
				Str("envName", cfg.environmentName).
				Msg("reached top of bruno collection dir but '.env' file not found")
			break
		}
		currentDir, err := filepath.Abs(parentDir)
		if err != nil {
			return nil, fmt.Errorf("could not get absolute path of '%s': %w", parentDir, err)
		}
		parentDir = path.Dir(currentDir)
		if parentDir == "/" {
			break
		}
	}
	return make(map[string]string), nil
}

func isBrunoCollectionRootDir(parentDir string) bool {
	return fileExists(path.Join(parentDir, "bruno.json"))
}

func maybePrettyPrint(data []byte) []byte {
	if len(data) == 0 {
		return nil
	}

	mimeType := mimetype.Detect(data)
	if mimeType == nil {
		return data
	}

	if mimeType.Is("application/json") {
		var buf bytes.Buffer
		if err := json.Indent(&buf, data, "", "  "); err != nil {
			return data
		}
		return buf.Bytes()
	}
	return data
}

func getVariablesFromFile(filePath string) (map[string]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %w", err)
	}

	defer f.Close()
	bruFile, err := bruparser.NewBruFile(f)
	if err != nil {
		return nil, fmt.Errorf("could not parse file: %w", err)
	}
	log.Info().
		Str("file", filePath).
		Any("variables", bruFile.Variables()).
		Msg("Loading environment variables")
	return bruFile.Variables(), nil
}

func getVariablesFromEnvFile(filePath string) (map[string]string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("could not open file: %w", err)
	}

	defer f.Close()
	envVars, err := envparse.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("could not parse file: %w", err)
	}
	log.Info().
		Str("file", filePath).
		Int("variables", len(envVars)).
		Msg("Loading environment variables")
	return envVars, nil
}

func getHash(data []byte) string {
	h := sha256.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))[:8]
}

func fileExists(filePath string) bool {
	stat, err := os.Stat(filePath)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		log.Debug().
			Str("filePath", filePath).
			Msg("file does not exist")
		return false
	}
	return !stat.IsDir()
}

func dirExists(dirPath string) bool {
	stat, err := os.Stat(dirPath)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		log.Info().
			Str("dir", dirPath).
			Msg("dir does not exist")
		return false
	}
	return stat.IsDir()
}

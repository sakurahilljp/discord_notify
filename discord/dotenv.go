package discord

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"
)

// LoadDotEnv reads a .env file and sets environment variables if they are not already set.
// It will not overwrite existing environment variables.
func LoadDotEnv(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return loadDotEnvFromReader(f, false)
}

// LoadDotEnvIfExists loads a .env file if it exists, otherwise ignores and returns nil.
func LoadDotEnvIfExists(filename string) error {
	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return LoadDotEnv(filename)
}

func loadDotEnvFromReader(r io.Reader, overwrite bool) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, val, ok := parseEnvLine(line)
		if !ok || key == "" {
			continue
		}

		if overwrite || os.Getenv(key) == "" {
			_ = os.Setenv(key, val)
		}
	}
	return scanner.Err()
}

func parseEnvLine(line string) (string, string, bool) {
	idx := strings.Index(line, "=")
	if idx < 0 {
		return "", "", false
	}

	key := strings.TrimSpace(line[:idx])
	// Strip "export " prefix if present
	key = strings.TrimPrefix(key, "export ")
	key = strings.TrimSpace(key)

	val := strings.TrimSpace(line[idx+1:])

	// Handle quoted values
	if len(val) >= 2 {
		if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) ||
			(strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
			val = val[1 : len(val)-1]
			return key, val, true
		}
	}

	// Handle inline comment after non-quoted value
	if commentIdx := strings.Index(val, "#"); commentIdx >= 0 {
		val = strings.TrimSpace(val[:commentIdx])
	}

	return key, val, true
}

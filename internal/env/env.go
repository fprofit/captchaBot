package env

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func Load() error {
	file, err := os.Open(".env")
	if err != nil {
		return fmt.Errorf("open env file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Пустая строка.
		if line == "" {
			continue
		}

		// Комментарий.
		if strings.HasPrefix(line, "#") {
			continue
		}

		key, value, err := parseLine(line)
		if err != nil {
			return err
		}

		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf(
				"set environment variable %q: %w",
				key,
				err,
			)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read env file: %w", err)
	}

	return nil
}

func parseLine(line string) (string, string, error) {
	key, value, ok := strings.Cut(line, "=")
	if !ok {
		return "", "", fmt.Errorf(
			"invalid env line: %q",
			line,
		)
	}

	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)

	if key == "" {
		return "", "", fmt.Errorf(
			"empty environment variable name",
		)
	}

	value = unquote(value)

	return key, value, nil
}

func unquote(value string) string {
	if len(value) < 2 {
		return value
	}

	if (value[0] == '"' && value[len(value)-1] == '"') ||
		(value[0] == '\'' && value[len(value)-1] == '\'') {
		return value[1 : len(value)-1]
	}

	return value
}

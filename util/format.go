package util

import (
	"encoding/json"
	"errors"

	"gopkg.in/yaml.v3"
)

const (
	// FormatUnknown represents an unknown data format
	FormatUnknown = iota

	// FormatText represents a raw string
	FormatText

	// FormatJSON represents data formatted as JSON
	FormatJSON

	// FormatYaml represents data formatted as YAML
	FormatYaml
)

// FormatData formats the given value with the given format
func FormatData(data interface{}, format int) (string, error) {
	switch format {
	case FormatYaml:
		d, err := yaml.Marshal(&data)
		if err != nil {
			return "", err
		}

		return string(d), nil

	case FormatJSON:
		d, err := json.MarshalIndent(&data, "", "    ")
		if err != nil {
			return "", err
		}

		return string(d), nil

	case FormatText:
		switch v := data.(type) {
		case string:
			return v, nil
		}

		return "", errors.New("data cannot be formatted as text, it is not a string")
	}

	return "", errors.New("unknown format")
}

// ParseData parses a string of bytes into the given format
func ParseData(data []byte, format int) (interface{}, error) {
	switch format {
	case FormatYaml:
		var d interface{}

		err := yaml.Unmarshal(data, &d)
		if err != nil {
			return nil, err
		}

		return d, nil

	case FormatJSON:
		var d interface{}

		err := json.Unmarshal(data, &d)
		if err != nil {
			return nil, err
		}

		return d, nil

	case FormatText:
		return string(data), nil
	}

	return "", errors.New("unknown format")
}

// FormatToName returns a text representation of a format code
func FormatToName(format int) string {
	switch format {
	case FormatYaml:
		return "yaml"

	case FormatJSON:
		return "json"

	case FormatText:
		return "text"

	default:
		return "unknown"
	}
}

// NameToFormat returns a number representation of a format name
func NameToFormat(name string) int {
	switch name {
	case "yaml":
		return FormatYaml

	case "json":
		return FormatJSON

	case "text":
		return FormatText

	default:
		return FormatUnknown
	}
}

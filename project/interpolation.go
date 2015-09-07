package project

import (
	"bytes"
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
)

func parseVariable(line string, pos int, mapping func(string) string) (string, int, bool) {
	var buffer bytes.Buffer

	for ; pos < len(line); pos++ {
		c := line[pos]

		switch {
		case c == '_' || (c >= 'A' && c <= 'Z'):
			buffer.WriteByte(c)
		default:
			return mapping(buffer.String()), pos - 1, true
		}
	}

	return mapping(buffer.String()), pos, true
}

func parseVariableWithBraces(line string, pos int, mapping func(string) string) (string, int, bool) {
	var buffer bytes.Buffer

	for ; pos < len(line); pos++ {
		c := line[pos]

		switch {
		case c == '}':
			bufferString := buffer.String()
			if bufferString == "" {
				return "", 0, false
			} else {
				return mapping(buffer.String()), pos, true
			}
		case c == '_' || (c >= 'A' && c <= 'Z'):
			buffer.WriteByte(c)
		default:
			return "", 0, false
		}
	}

	return "", 0, false
}

func parseInterpolationExpression(line string, pos int, mapping func(string) string) (string, int, bool) {
	c := line[pos]

	switch {
	case c == '$':
		return "$", pos, true
	case c == '{':
		return parseVariableWithBraces(line, pos+1, mapping)
	case c >= 'A' && c <= 'Z':
		return parseVariable(line, pos, mapping)
	default:
		return "", 0, false
	}

	return "", pos, true
}

func parseLine(line string, mapping func(string) string) (string, bool) {
	var buffer bytes.Buffer

	for pos := 0; pos < len(line); pos++ {
		c := line[pos]
		switch {
		case c == '$':
			var replaced string
			var success bool

			replaced, pos, success = parseInterpolationExpression(line, pos+1, mapping)

			if !success {
				return "", false
			}

			buffer.WriteString(replaced)
		default:
			buffer.WriteByte(c)
		}
	}

	return buffer.String(), true
}

func interpolate(option, service string, data *interface{}, mapping func(string) string) error {
	switch typedData := (*data).(type) {
	case string:
		var success bool
		*data, success = parseLine(typedData, mapping)

		if !success {
			return fmt.Errorf("Invalid interpolation format for \"%s\" option in service \"%s\": \"%s\"", option, service, typedData)
		}
	case []interface{}:
		for k, v := range typedData {
			err := interpolate(option, service, &v, mapping)

			if err != nil {
				return err
			}

			typedData[k] = v
		}
	case map[interface{}]interface{}:
		for k, v := range typedData {
			err := interpolate(option, service, &v, mapping)

			if err != nil {
				return err
			}

			typedData[k] = v
		}
	}

	return nil
}

func Interpolate(config *rawServiceMap) error {
	for k, v := range *config {
		for k2, v2 := range v {
			err := interpolate(k2, k, &v2, func(s string) string {
				value := os.Getenv(s)

				if value == "" {
					logrus.Warnf("The %s variable is not set. Substituting a blank string.", s)
				}

				return value
			})

			if err != nil {
				return err
			}

			(*config)[k][k2] = v2
		}
	}

	return nil
}

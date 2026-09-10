package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

var taylorToolNamePattern = regexp.MustCompile(`(?m)^\s*name:\s*"([^"]+)",\s*$`)

func TestTaylorToolsSchemaMatchesFrozenGolden(t *testing.T) {
	repositoryRoot := repositoryRoot(t)
	source, err := os.ReadFile(filepath.Join(repositoryRoot, "taylor-tools.ts"))
	if err != nil {
		t.Fatalf("read taylor-tools.ts: %v", err)
	}
	actual, err := parseTaylorToolSchemas(string(source))
	if err != nil {
		t.Fatalf("parse taylor-tools.ts: %v", err)
	}
	golden, err := os.ReadFile(filepath.Join(repositoryRoot, "internal", "tools", "testdata", "params_schema.json"))
	if err != nil {
		t.Fatalf("read frozen schema golden: %v", err)
	}
	var want map[string][]string
	if err := json.Unmarshal(golden, &want); err != nil {
		t.Fatalf("decode frozen schema golden: %v", err)
	}
	if !reflect.DeepEqual(actual, want) {
		t.Fatalf("taylor-tools.ts schema changed\nwant: %#v\ngot: %#v", want, actual)
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate test source")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(source), "..", ".."))
}

func parseTaylorToolSchemas(source string) (map[string][]string, error) {
	matches := taylorToolNamePattern.FindAllStringSubmatchIndex(source, -1)
	if len(matches) == 0 {
		return nil, fmt.Errorf("no tool names found")
	}
	schemas := make(map[string][]string, len(matches)+1)
	for index, match := range matches {
		name := source[match[2]:match[3]]
		end := len(source)
		if index+1 < len(matches) {
			end = matches[index+1][0]
		}
		toolSource := source[match[1]:end]
		objectStart, err := typeObjectStart(toolSource, "parameters: Type.Object(")
		if err != nil {
			return nil, fmt.Errorf("%s parameters: %w", name, err)
		}
		keys, _, err := typeObjectKeys(toolSource, objectStart)
		if err != nil {
			return nil, fmt.Errorf("%s parameters: %w", name, err)
		}
		schemas[name] = keys
		if name == "apply_patch" {
			hunksStart, err := typeObjectStart(toolSource, "hunks: Type.Array(Type.Object(")
			if err != nil {
				return nil, fmt.Errorf("apply_patch hunks: %w", err)
			}
			hunkKeys, _, err := typeObjectKeys(toolSource, hunksStart)
			if err != nil {
				return nil, fmt.Errorf("apply_patch hunks: %w", err)
			}
			schemas["apply_patch.hunks"] = hunkKeys
		}
	}
	return schemas, nil
}

func typeObjectStart(source, prefix string) (int, error) {
	index := strings.Index(source, prefix)
	if index < 0 {
		return 0, fmt.Errorf("Type.Object not found")
	}
	start := strings.Index(source[index+len(prefix):], "{")
	if start < 0 {
		return 0, fmt.Errorf("Type.Object opening brace not found")
	}
	return index + len(prefix) + start, nil
}

func typeObjectKeys(source string, start int) ([]string, int, error) {
	depth := 1
	keys := make([]string, 0)
	for index := start + 1; index < len(source); index++ {
		if source[index] == '"' || source[index] == '\'' {
			index = skipString(source, index)
			continue
		}
		switch source[index] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return keys, index, nil
			}
		default:
			if depth != 1 || !isIdentifierStart(source[index]) {
				continue
			}
			end := index + 1
			for end < len(source) && isIdentifierPart(source[end]) {
				end++
			}
			after := end
			for after < len(source) && (source[after] == ' ' || source[after] == '\t' || source[after] == '\r' || source[after] == '\n') {
				after++
			}
			if after < len(source) && source[after] == ':' {
				keys = append(keys, source[index:end])
			}
			index = end - 1
		}
	}
	return nil, 0, fmt.Errorf("Type.Object closing brace not found")
}

func skipString(source string, start int) int {
	quote := source[start]
	for index := start + 1; index < len(source); index++ {
		if source[index] == '\\' {
			index++
			continue
		}
		if source[index] == quote {
			return index
		}
	}
	return len(source) - 1
}

func isIdentifierStart(character byte) bool {
	return character == '_' || character >= 'a' && character <= 'z' || character >= 'A' && character <= 'Z'
}

func isIdentifierPart(character byte) bool {
	return isIdentifierStart(character) || character >= '0' && character <= '9'
}

package pirpc

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// TestTCPIRPC001 enforces INV-9 (spec.md section 10): no source in
// internal/pirpc may construct Pi's host-level bash RPC command. The check
// runs over every Go file, including tests, so a future fixture must remain
// under testdata rather than becoming a source-file exception.
func TestTCPIRPC001(t *testing.T) {
	t.Run("clean package source", func(t *testing.T) {
		if err := checkPackageForBashRPCCommand("."); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("forbidden fixture", func(t *testing.T) {
		path := filepath.Join("testdata", "forbidden_bash_rpc.txt")
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read fixture: %v", err)
		}
		violations, err := findBashRPCCommands(path, source)
		if err != nil {
			t.Fatalf("check fixture: %v", err)
		}
		if len(violations) != 1 {
			t.Fatalf("fixture violations = %#v, want exactly one", violations)
		}
		if !strings.Contains(violations[0].String(), "forbidden_bash_rpc.txt:3") {
			t.Fatalf("fixture violation = %s, want file:line", violations[0])
		}
	})

	t.Run("forbidden string assembly", func(t *testing.T) {
		source := []byte(`package fixture

const forbiddenCommand = "{\"type\":" + "\"bash\"}"
`)
		violations, err := findBashRPCCommands("assembly.go", source)
		if err != nil {
			t.Fatalf("check assembled command: %v", err)
		}
		if len(violations) != 1 || violations[0].line != 3 {
			t.Fatalf("assembly violations = %#v, want one violation at line 3", violations)
		}
	})
}

type bashRPCViolation struct {
	file string
	line int
}

func (v bashRPCViolation) String() string {
	return fmt.Sprintf("%s:%d: INV-9 forbids a Pi bash RPC command", v.file, v.line)
}

func checkPackageForBashRPCCommand(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("list pirpc source: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".go" {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		source, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		violations, err := findBashRPCCommands(path, source)
		if err != nil {
			return err
		}
		if len(violations) > 0 {
			return fmt.Errorf("%s", violations[0])
		}
	}
	return nil
}

func findBashRPCCommands(filename string, source []byte) ([]bashRPCViolation, error) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, source, 0)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", filename, err)
	}
	constants := stringConstants(file)
	var violations []bashRPCViolation
	ast.Inspect(file, func(node ast.Node) bool {
		expr, ok := node.(ast.Expr)
		if !ok {
			return true
		}
		// Identifiers are resolved while evaluating a surrounding string
		// expression. Reporting them directly would duplicate a violation
		// already reported at the literal or concatenation that defines it.
		switch expr.(type) {
		case *ast.Ident, *ast.ParenExpr:
			return true
		}
		value, ok := stringValue(expr, constants)
		if !ok || !isBashRPCCommand(value) {
			return true
		}
		position := fset.Position(expr.Pos())
		violations = append(violations, bashRPCViolation{file: position.Filename, line: position.Line})
		return false
	})
	return violations, nil
}

func stringConstants(file *ast.File) map[string]string {
	constants := make(map[string]string)
	for _, declaration := range file.Decls {
		general, ok := declaration.(*ast.GenDecl)
		if !ok || general.Tok != token.CONST {
			continue
		}
		for _, specification := range general.Specs {
			valueSpec, ok := specification.(*ast.ValueSpec)
			if !ok || len(valueSpec.Names) != len(valueSpec.Values) {
				continue
			}
			for index, name := range valueSpec.Names {
				if value, ok := stringValue(valueSpec.Values[index], constants); ok {
					constants[name.Name] = value
				}
			}
		}
	}
	return constants
}

func stringValue(expr ast.Expr, constants map[string]string) (string, bool) {
	switch value := expr.(type) {
	case *ast.BasicLit:
		if value.Kind != token.STRING {
			return "", false
		}
		decoded, err := strconv.Unquote(value.Value)
		return decoded, err == nil
	case *ast.BinaryExpr:
		if value.Op != token.ADD {
			return "", false
		}
		left, leftOK := stringValue(value.X, constants)
		right, rightOK := stringValue(value.Y, constants)
		return left + right, leftOK && rightOK
	case *ast.Ident:
		constant, ok := constants[value.Name]
		return constant, ok
	case *ast.ParenExpr:
		return stringValue(value.X, constants)
	default:
		return "", false
	}
}

func isBashRPCCommand(value string) bool {
	var command map[string]json.RawMessage
	if err := json.Unmarshal([]byte(value), &command); err != nil {
		return false
	}
	var commandType string
	if err := json.Unmarshal(command["type"], &commandType); err != nil {
		return false
	}
	return commandType == "bash"
}

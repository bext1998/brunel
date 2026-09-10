package tools

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

var updateSchemaSnapshot = flag.Bool("update", false, "update tools schema snapshot")

func TestParameterSchemaSnapshot(t *testing.T) {
	snapshot := map[string][]string{
		"list_files":        jsonFieldNames(reflect.TypeOf(ListFilesParams{})),
		"search_text":       jsonFieldNames(reflect.TypeOf(SearchTextParams{})),
		"read_file":         jsonFieldNames(reflect.TypeOf(ReadFileParams{})),
		"apply_patch":       jsonFieldNames(reflect.TypeOf(ApplyPatchParams{})),
		"apply_patch.hunks": jsonFieldNames(reflect.TypeOf(PatchHunk{})),
		"create_file":       jsonFieldNames(reflect.TypeOf(CreateFileParams{})),
		"write_file":        jsonFieldNames(reflect.TypeOf(WriteFileParams{})),
		"run_powershell":    jsonFieldNames(reflect.TypeOf(RunPowerShellParams{})),
		"workspace_diff":    jsonFieldNames(reflect.TypeOf(WorkspaceDiffParams{})),
	}
	actual, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		t.Fatalf("marshal schema snapshot: %v", err)
	}
	actual = append(actual, '\n')
	path := filepath.Join("testdata", "params_schema.json")
	if *updateSchemaSnapshot {
		if err := os.WriteFile(path, actual, 0o600); err != nil {
			t.Fatalf("write schema snapshot: %v", err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read schema snapshot: %v", err)
	}
	// The snapshot's contract is the field-name lists, not byte-exact line
	// endings: a core.autocrlf checkout leaves the committed file CRLF on
	// Windows while the freshly marshalled bytes are always LF.
	normalize := func(b []byte) string { return strings.ReplaceAll(string(b), "\r\n", "\n") }
	if normalize(actual) != normalize(want) {
		t.Fatalf("parameter schema changed; run go test ./internal/tools -update\nwant:\n%s\nactual:\n%s", want, actual)
	}
}

func jsonFieldNames(typ reflect.Type) []string {
	fields := make([]string, 0, typ.NumField())
	for i := 0; i < typ.NumField(); i++ {
		name := strings.Split(typ.Field(i).Tag.Get("json"), ",")[0]
		fields = append(fields, name)
	}
	return fields
}

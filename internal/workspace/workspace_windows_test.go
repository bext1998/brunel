//go:build windows

package workspace

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestBindRejectsInvalidRoots(t *testing.T) {
	if _, err := Bind(""); !errors.Is(err, ErrWorkspaceInvalid) {
		t.Fatalf("Bind(empty) error = %v, want E_WORKSPACE_INVALID", err)
	}

	missing := filepath.Join(t.TempDir(), "missing")
	if _, err := Bind(missing); !errors.Is(err, ErrWorkspaceInvalid) {
		t.Fatalf("Bind(missing) error = %v, want E_WORKSPACE_INVALID", err)
	} else if code := ErrorCode(err); code != "E_WORKSPACE_INVALID" {
		t.Fatalf("ErrorCode(Bind(missing)) = %q", code)
	}

	file := filepath.Join(t.TempDir(), "not-a-directory")
	if err := os.WriteFile(file, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Bind(file); !errors.Is(err, ErrWorkspaceInvalid) {
		t.Fatalf("Bind(file) error = %v, want E_WORKSPACE_INVALID", err)
	}
}

func TestResolveRootAndNestedNewTargetStayWithinFinalRoot(t *testing.T) {
	ws := newWorkspace(t)
	resolvedRoot, err := ws.Resolve(".")
	if err != nil {
		t.Fatal(err)
	}
	if !samePath(resolvedRoot, ws.finalRoot) {
		t.Fatalf("Resolve(\\\".\\\") = %q, want final root %q", resolvedRoot, ws.finalRoot)
	}
	if err := os.Mkdir(filepath.Join(ws.Root(), "existing"), 0o700); err != nil {
		t.Fatal(err)
	}
	resolved, err := ws.Resolve(filepath.Join("existing", "new", "nested", "file.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if !containsPath(ws.finalRoot, resolved) {
		t.Fatalf("Resolve(nested new target) = %q, outside final root %q", resolved, ws.finalRoot)
	}
}

func TestResolveRejectsAbsoluteAndLexicalEscape(t *testing.T) {
	ws := newWorkspace(t)
	outside := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(outside, []byte("outside"), 0o600); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{outside, `C:relative.txt`, `..\outside.txt`} {
		if _, err := ws.Resolve(path); !errors.Is(err, ErrPathEscape) {
			t.Errorf("Resolve(%q) error = %v, want E_PATH_ESCAPE", path, err)
		}
	}
	contents, err := os.ReadFile(outside)
	if err != nil {
		t.Fatal(err)
	}
	if string(contents) != "outside" {
		t.Fatalf("outside file changed to %q", contents)
	}
}

func TestResolveBlocksSymlinkAndCommonPrefixEscape(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "work")
	outside := filepath.Join(parent, "workspace")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(outside, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(outside, "secret.txt"), []byte("outside"), 0o600); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "escape")
	if err := os.Symlink(outside, link); err != nil {
		t.Skipf("cannot create Windows symlink: %v", err)
	}
	ws, err := Bind(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ws.Resolve(filepath.Join("escape", "secret.txt")); !errors.Is(err, ErrPathEscape) {
		t.Fatalf("Resolve(symlink escape) error = %v, want E_PATH_ESCAPE", err)
	}
}

func TestContainsPathRequiresComponentBoundary(t *testing.T) {
	if containsPath(`\\?\C:\work`, `\\?\C:\workspace`) {
		t.Fatal("common-prefix sibling must not be contained by workspace root")
	}
	if !containsPath(`\\?\C:\work`, `\\?\C:\work\file.txt`) {
		t.Fatal("workspace child must be contained by workspace root")
	}
}

func TestResolveBlocksJunctionEscape(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "root")
	outside := filepath.Join(parent, "outside")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(outside, 0o700); err != nil {
		t.Fatal(err)
	}
	junction := filepath.Join(root, "escape")
	command := exec.Command("cmd", "/c", "mklink", "/J", junction, outside)
	if output, err := command.CombinedOutput(); err != nil {
		t.Skipf("cannot create junction: %v (%s)", err, output)
	}
	ws, err := Bind(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ws.Resolve(filepath.Join("escape", "new.txt")); !errors.Is(err, ErrPathEscape) {
		t.Fatalf("Resolve(junction escape) error = %v, want E_PATH_ESCAPE", err)
	}
}

func TestResolveSupportsNonASCIIAndCaseAlias(t *testing.T) {
	ws := newWorkspace(t)
	dir := filepath.Join(ws.Root(), "資料夾")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(file, []byte("ok"), 0o600); err != nil {
		t.Fatal(err)
	}
	resolved, err := ws.Resolve(filepath.Join("資料夾", "FILE.TXT"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.EqualFold(filepath.Base(resolved), "file.txt") {
		t.Fatalf("Resolve() = %q, want final file path", resolved)
	}
}

func TestResolveRejectsChangedRootIdentity(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "root")
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	ws, err := Bind(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(root, filepath.Join(parent, "old-root")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(root, 0o700); err != nil {
		t.Fatal(err)
	}
	if _, err := ws.Resolve("file.txt"); !errors.Is(err, ErrWorkspaceUnbound) {
		t.Fatalf("Resolve() error = %v, want E_WORKSPACE_UNBOUND", err)
	}
}

func newWorkspace(t *testing.T) *Workspace {
	t.Helper()
	ws, err := Bind(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return ws
}

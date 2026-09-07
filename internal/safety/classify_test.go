package safety

import "testing"

func TestClassifyPowerShellAutoForOrdinaryCommands(t *testing.T) {
	auto := []string{
		"go test ./...",
		"go vet ./...",
		"go build ./...",
		"npm run build",
		"npm test",
		"golangci-lint run",
		"git status",
		"git diff",
		"git log --oneline -5",
		"Remove-Item .\\build\\output.txt", // single, non-forced, non-recursive delete
		"Move-Item a.txt .\\archive\\",    // rename into a directory, no -Force, no wildcard
	}
	for _, cmd := range auto {
		if _, confirm := classifyPowerShell(cmd, `C:\ws`); confirm {
			t.Errorf("classifyPowerShell(%q) wants AUTO, got CONFIRM", cmd)
		}
	}
}

func TestClassifyPowerShellConfirmsEachRepresentativeCategory(t *testing.T) {
	confirm := []string{
		`Remove-Item -Recurse -Force .\build`,
		`Clear-Content .\log.txt`,
		`Move-Item .\a.txt .\b.txt -Force`,
		`git commit -m "wip"`,
		`git push origin main`,
		`git reset --hard HEAD~1`,
		`git clean -fd`,
		`git rebase main`,
		`npm install left-pad`,
		`pip install requests`,
		`winget install Git.Git`,
		`Invoke-WebRequest -Uri https://example.com`,
		`Invoke-RestMethod -Uri https://example.com`,
		`wget https://example.com`,
		`Start-Process notepad.exe`,
		`Start-Job { Get-Process }`,
	}
	for _, cmd := range confirm {
		if _, ok := classifyPowerShell(cmd, `C:\ws`); !ok {
			t.Errorf("classifyPowerShell(%q) wants CONFIRM, got AUTO", cmd)
		}
	}
}

func TestClassifyPowerShellFlagsOnlyOutOfWorkspaceAbsolutePaths(t *testing.T) {
	if _, ok := classifyPowerShell(`Get-Content C:\ws\a.txt`, `C:\ws`); ok {
		t.Error("in-workspace absolute path should not require confirmation")
	}
	if _, ok := classifyPowerShell(`Get-Content C:\Windows\System32\config`, `C:\ws`); !ok {
		t.Error("out-of-workspace absolute path should require confirmation")
	}
	if _, ok := classifyPowerShell(`Get-Content \\server\share\file.txt`, `C:\ws`); !ok {
		t.Error("UNC path outside workspace should require confirmation")
	}
	// No workspace root configured: nothing to compare against, so the
	// absolute-path check is disabled rather than flagging everything.
	if _, ok := classifyPowerShell(`Get-Content C:\Windows\System32\config`, ""); ok {
		t.Error("empty workspace root should disable the absolute-path check, not flag everything")
	}
}

// Command separators (`;`, `|`, `&`) must not let a high-risk flag glued
// to a separator slip past token matching (PR #27 review issue 1).
func TestClassifyPowerShellCommandSeparatorsDoNotBypassFlags(t *testing.T) {
	confirm := []string{
		`Remove-Item .\build -Recurse; Write-Output done`,
		`Remove-Item .\build -Force& Write-Output done`,
		`Remove-Item .\build -Recurse | Measure-Object`,
		`Move-Item .\a.txt .\b.txt -Force; git status`,
		`git commit -m "wip"; go test ./...`,
		`npm install left-pad; npm test`,
		`Start-Process notepad.exe; Write-Output started`,
	}
	for _, cmd := range confirm {
		if _, ok := classifyPowerShell(cmd, `C:\ws`); !ok {
			t.Errorf("classifyPowerShell(%q) wants CONFIRM, got AUTO", cmd)
		}
	}
}

// Windows accepts forward slashes in absolute paths, and `..` segments
// must be resolved before the workspace containment check (PR #27 review
// issue 2).
func TestClassifyPowerShellDetectsForwardSlashAndDotDotEscapePaths(t *testing.T) {
	confirm := []string{
		`Get-Content C:/Windows/win.ini`,
		`Get-Content C:\ws\..\Windows\win.ini`,
		`Get-Content C:/ws/../Windows/win.ini`,
		`Copy-Item C:\ws\..\secret.txt D:\elsewhere`,
	}
	for _, cmd := range confirm {
		if _, ok := classifyPowerShell(cmd, `C:\ws`); !ok {
			t.Errorf("classifyPowerShell(%q) wants CONFIRM, got AUTO", cmd)
		}
	}
	// The same forms pointing *inside* the workspace stay AUTO.
	auto := []string{
		`Get-Content C:/ws/a.txt`,
		`Get-Content C:\ws\sub\..\a.txt`,
	}
	for _, cmd := range auto {
		if _, ok := classifyPowerShell(cmd, `C:\ws`); ok {
			t.Errorf("classifyPowerShell(%q) wants AUTO, got CONFIRM", cmd)
		}
	}
}

// Overwrite and bulk-move forms that do not use -Force must still be
// confirmed (PR #27 review issue 3): a wildcard move is bulk, and a
// copy/move onto an existing destination overwrites it.
func TestClassifyPowerShellConfirmsBulkMoveAndOverwriteWithoutForce(t *testing.T) {
	confirm := []string{
		`Move-Item .\*.txt .\archive\`,
		`Copy-Item .\replacement.txt .\existing.txt`,
		`Move-Item *.go .\archive\`,
	}
	for _, cmd := range confirm {
		if _, ok := classifyPowerShell(cmd, `C:\ws`); !ok {
			t.Errorf("classifyPowerShell(%q) wants CONFIRM, got AUTO", cmd)
		}
	}
	// A single, non-wildcard move/copy that does not name an existing
	// destination stays AUTO (the ordinary rename case).
	if _, ok := classifyPowerShell(`Move-Item a.txt .\archive\`, `C:\ws`); ok {
		t.Error("plain single-file move into a directory (trailing separator) without -Force or wildcard should stay AUTO")
	}
}

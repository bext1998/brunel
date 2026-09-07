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
		"Move-Item a.txt b.txt",            // no -Force
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

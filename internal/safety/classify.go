package safety

import (
	"regexp"
	"strings"
)

// classify decides AUTO vs CONFIRM for one tool call (spec.md §6.2).
// Every file tool - list_files, search_text, read_file, workspace_diff,
// apply_patch, create_file, write_file - is always AUTO: reads carry no
// risk, and writes are already exact and hash-guarded elsewhere
// (workspace + internal/filetools own that precondition, Issue #5). Only
// run_powershell's command text is inspected against the six
// representative CONFIRM categories from spec.md §6.2, checked in the
// order they are listed there; the first match wins and its category name
// becomes the reason shown in the approval prompt.
func (g *Gate) classify(call ToolCall) (Risk, string) {
	if call.Tool != "run_powershell" {
		return RiskAuto, ""
	}
	if reason, confirm := classifyPowerShell(call.Command, g.WorkspaceRoot); confirm {
		return RiskConfirm, reason
	}
	return RiskAuto, ""
}

// Representative command-form patterns. These are deliberately simple,
// best-effort token/substring checks - spec.md §6.2 explicitly disclaims
// full PowerShell semantic coverage ("不得宣稱涵蓋完整 PowerShell") and its
// test plan only requires the listed representative commands to classify
// correctly, not language completeness.
var (
	deleteVerbs      = map[string]bool{"remove-item": true, "ri": true, "del": true, "erase": true, "rd": true, "rmdir": true}
	deleteForceFlags = map[string]bool{"-recurse": true, "-force": true}
	clearVerbs       = map[string]bool{"clear-content": true, "clear-item": true}
	overwriteVerbs   = map[string]bool{"move-item": true, "copy-item": true}

	gitStateChangingSubcommands = map[string]bool{"commit": true, "push": true, "clean": true, "rebase": true}

	installTools    = map[string]bool{"npm": true, "npx": true, "pip": true, "pip3": true, "winget": true, "choco": true, "dotnet": true, "yarn": true, "pnpm": true}
	installVerbs    = map[string]bool{"install": true, "i": true, "add": true, "update": true, "upgrade": true, "ci": true}
	networkCommands = map[string]bool{"invoke-webrequest": true, "invoke-restmethod": true, "iwr": true, "irm": true, "curl": true, "curl.exe": true, "wget": true}
	backgroundVerbs = map[string]bool{"start-process": true, "start-job": true}

	// Matches a Windows drive-letter absolute path or a UNC share
	// anywhere in the command text, e.g. C:\Users\x or \\server\share.
	reAbsoluteWindowsPath = regexp.MustCompile(`(?i)[A-Z]:\\[^\s"'|]*|\\\\[^\s"'|]+`)
)

// classifyPowerShell returns (reason, true) if command matches one of the
// six representative CONFIRM categories, checked in spec order.
func classifyPowerShell(command, workspaceRoot string) (string, bool) {
	tokens := tokenize(command)

	if containsAny(tokens, deleteVerbs) && containsAny(tokens, deleteForceFlags) {
		return "recursive or forced delete", true
	}
	if containsAny(tokens, clearVerbs) {
		return "clears file content", true
	}
	if containsAny(tokens, overwriteVerbs) && tokens["-force"] {
		return "bulk move or overwrite", true
	}

	if tokens["git"] {
		if containsAny(tokens, gitStateChangingSubcommands) {
			return "git state-changing command", true
		}
		if tokens["reset"] && tokens["--hard"] {
			return "git state-changing command", true
		}
	}

	if containsAny(tokens, installTools) && containsAny(tokens, installVerbs) {
		return "installs or updates dependencies or packages", true
	}

	if containsAny(tokens, networkCommands) {
		return "network transmission command", true
	}

	if containsAny(tokens, backgroundVerbs) {
		return "background process or job", true
	}

	if path, ok := firstOutOfWorkspaceAbsolutePath(command, workspaceRoot); ok {
		return "references an absolute path outside the workspace: " + path, true
	}

	return "", false
}

// tokenize splits command on whitespace into a lowercase token set for
// membership checks. It is intentionally simple (no quote-aware parsing):
// classification here is best-effort, not a PowerShell parser.
func tokenize(command string) map[string]bool {
	fields := strings.Fields(command)
	tokens := make(map[string]bool, len(fields))
	for _, f := range fields {
		tokens[strings.ToLower(strings.Trim(f, `"'`))] = true
	}
	return tokens
}

func containsAny(tokens map[string]bool, set map[string]bool) bool {
	for t := range set {
		if tokens[t] {
			return true
		}
	}
	return false
}

// firstOutOfWorkspaceAbsolutePath returns the first absolute Windows path
// found in command that does not fall under workspaceRoot. An empty
// workspaceRoot disables this check (nothing to compare against).
func firstOutOfWorkspaceAbsolutePath(command, workspaceRoot string) (string, bool) {
	if strings.TrimSpace(workspaceRoot) == "" {
		return "", false
	}
	root := strings.TrimRight(workspaceRoot, `\/`)
	for _, match := range reAbsoluteWindowsPath.FindAllString(command, -1) {
		trimmed := strings.TrimRight(match, `\/`)
		if !isWithinRoot(root, trimmed) {
			return match, true
		}
	}
	return "", false
}

func isWithinRoot(root, path string) bool {
	if strings.EqualFold(root, path) {
		return true
	}
	if len(path) <= len(root) || !strings.EqualFold(path[:len(root)], root) {
		return false
	}
	sep := path[len(root)]
	return sep == '\\' || sep == '/'
}

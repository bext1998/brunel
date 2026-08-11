# Phase 0 Environment Snapshot

Captured: 2026-08-12 (Asia/Taipei)

| Component | Observed version / state |
|---|---|
| Node | `v24.16.0` |
| npm | `11.18.0` |
| Git | `2.53.0.windows.2` |
| PowerShell | `7.6.4` |
| Pi package | `@earendil-works/pi-coding-agent`, `0.84.1` |
| Worktree | `D:\AgentCoding\Brunel-pi-spike` |
| Branch | `agent/pi-spike-issue-24` |

## Credential availability

The check was performed by `recon-credentials.ps1`. It reports presence/state only and never prints secret values.

- `OPENROUTER_API_KEY`: present.
- `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`, `GEMINI_API_KEY`: absent.
- `~/.pi/agent/settings.json`: present; no provider names detected by the non-secret probe.
- Windows Credential Manager: a matching provider-related entry was present.

The OpenRouter credential was sufficient for Pi startup/model selection in the Gate 0 clean-environment probe. The exact credential value was not read into any repository file.

## Installation note

`npm install -g @earendil-works/pi-coding-agent` completed with exit code 0. npm emitted a cleanup warning for a pre-existing temporary package directory (`EPERM`) and an allow-scripts warning; `pi --version` returned `0.84.1` afterward.

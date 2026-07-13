# Brunel — 下一步行動

> 最後同步：2026-07-14

## 下一個 Session 目標

Issue #6（F-5 PowerShell Job Object 執行器）已在本機完成實作與測試，待 review 並決定是否開 PR；之後可續接 #4 stale-read 防護。

## 優先行動

1. Review `internal/exec`（分支 `maze/2026-07-14-ec533b`）：Job Object 綁定順序、逃逸視窗處理、資源上限 API 設計是否符合預期，決定是否開 PR。
2. PR 合併後，開始 #4（F-4），以 workspace guard 為前置條件實作 stale-read 與 AC-5。
3. `internal/exec` 的 `Options` 要求呼叫端明確帶入 Timeout／MaxProcesses／MaxMemoryBytes／MaxOutputBytes；實際數字待 OQ-4 裁決，之後串接 `tools`/`safety` 時需留意不得自行填入預設值。

## 阻塞與待決策

- AC-5 受 #4 未實作阻塞。
- OQ-4（benchmark 硬性上限）未裁決，`internal/exec` 呼叫端仍無法帶入具體數字。

## 參考

- `docs/spec.md` §0、§5.1、§7.3～§7.5、§10.3、§11.2、§13、§15.1～§15.5、§15.8
- `MAZE_PROJECT.md`

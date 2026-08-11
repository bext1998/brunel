# Brunel — 下一步行動

> 最後同步：2026-08-12

## 下一個 Session 目標

#24 Pi Compatibility Spike 已完成於 `agent/pi-spike-issue-24`，等待指揮者 QA；Gate 0 仍需在未安裝 Git Bash 的 Windows 環境補測。

## 優先行動

1. 由指揮者驗收 `spikes/pi-compatibility/REPORT.md`、Gate logs 與 branch diff。
2. 若要完成 Route B 判斷，在沒有 Git Bash 的 Windows VM/runner 重做 Gate 0。
3. 之後串接 `internal/exec` 到 `tools`/`safety` 時，`Options` 的 Timeout／MaxProcesses／MaxMemoryBytes／MaxOutputBytes 需呼叫端明確帶入，不得自行填入預設值；實際數字待 OQ-4 裁決。

## 阻塞與待決策

- OQ-4（benchmark 硬性上限）未裁決，`internal/exec` 呼叫端與未來 benchmark（#15）仍無法帶入具體數字。

## 參考

- `docs/spec.md` §0、§5.1、§7.3～§7.5、§10.3、§11.2、§13、§15.1～§15.5、§15.8
- `MAZE_PROJECT.md`

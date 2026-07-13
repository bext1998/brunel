# Brunel — 下一步行動

> 最後同步：2026-07-14

## 下一個 Session 目標

#3（Workspace）與 #6（F-5 exec）皆已合併並關閉；下一個未阻塞的 P1 項目是 #5（F-4 stale-read 防護，AC-5），以 workspace guard 為前置條件，現已具備。

## 優先行動

1. 開始 #5（F-4）：實作 stale-read hash 前置條件與原子寫入，對應 AC-5。
2. 之後串接 `internal/exec` 到 `tools`/`safety` 時，`Options` 的 Timeout／MaxProcesses／MaxMemoryBytes／MaxOutputBytes 需呼叫端明確帶入，不得自行填入預設值；實際數字待 OQ-4 裁決。

## 阻塞與待決策

- OQ-4（benchmark 硬性上限）未裁決，`internal/exec` 呼叫端與未來 benchmark（#15）仍無法帶入具體數字。

## 參考

- `docs/spec.md` §0、§5.1、§7.3～§7.5、§10.3、§11.2、§13、§15.1～§15.5、§15.8
- `MAZE_PROJECT.md`

# Brunel — 下一步行動

> 最後同步：2026-09-05

## 下一個 Session 目標

#5（F-4 stale-read hash 防護與原子寫入）的核心邏輯已在 `internal/filetools` 完成並通過單元測試，PR 待 review／合併。下一個 Session 應先確認 #5 的 review 結果並視需要修正，其餘重點仍是把 ADR-002 的架構轉向拆解成可執行的 GitHub Issues，並釐清 #8（F-7 Provider）／#9（F-8 Agent Loop）哪些部分仍要做、哪些改由 Pi 承接。

## 優先行動

1. 追蹤 #5 的 PR review：處理任何 review 意見；合併後同步關閉 #5 並更新 STATUS.md／NEXT_ACTION.md。
2. 之後處理 #7 的 AUTO／CONFIRM 安全入口（目前唯一無相依阻塞的正式子項）；#4（F-3 工具，需接上 `internal/filetools` 與 `workspace.Workspace`）與 #2（F-1 CLI/TUI）的相依阻塞待 #7 解除後再排。
3. 用 `maze-spec-to-issues` 或等效流程，把 ADR-002 的「後續需要」拆成具體 Issues：(a) 未安裝 Git Bash 的 Windows VM/runner 補測 Gate 0、(b) Taylor RPC client 的 `bash` command allowlist/lint 防線設計與實作、(c) Route B 正式整合（取代 #8/#9 原本的 Provider Adapter / Agent Loop 範圍）。
4. 檢視 `docs/spec.md` §5（架構與公開介面）、§9（Contract），依 ADR-002 修訂範圍（哪些改由 Pi 提供、哪些仍需 Go 自建的 Taylor tools 邊界），完成後再排入 #8/#9 後續工作。
5. 規劃 #2 前將 Go module 基線由 1.22 同步至 1.25.x，並引入 Bubble Tea v2；此項需另行實作授權，且與 Route B 無關（Host 層仍是 Go）。

## 阻塞與待決策

- Route B 正式整合範圍尚未拆成 Issues；`docs/spec.md` §5／§9 的詳細內容需在拆 Issue 前先決定去留。
- 無 Alpha 1 硬阻塞；`docs/spec.md` §16 的 Open Questions 依各自裁決前行為處理。

## 參考

- `docs/spec.md` §4～§6、§8～§16
- `docs/adr/ADR-002-pi-agent-runtime.md`
- `MAZE_PROJECT.md`

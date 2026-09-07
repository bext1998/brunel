# Brunel — 下一步行動

> 最後同步：2026-09-07

## 下一個 Session 目標

#5（PR #26）與 #7（本次，PR 待建立）的核心邏輯都已完成並等待 review／合併。下一個 Session 應先確認兩者的 review 結果並視需要修正；合併後即解除 #4（F-3 工具，需接上 `internal/filetools` 與 `internal/safety`）與 #2（F-1 CLI/TUI，需提供 `safety.Approver` 的 TUI／純文字實作）的相依阻塞。其餘重點仍是把 ADR-002 的架構轉向拆解成可執行的 GitHub Issues，並釐清 #8（F-7 Provider）／#9（F-8 Agent Loop）哪些部分仍要做、哪些改由 Pi 承接。

## 優先行動

1. 追蹤 #5（PR #26）與 #7 的 PR review：處理任何 review 意見；合併後同步關閉對應 Issue 並更新 STATUS.md／NEXT_ACTION.md。
2. #4（F-3：8 工具 schema 與安全入口接線）：待 #7 合併後排入，把 `internal/filetools`（#5）與 `internal/safety`（#7）接上 `workspace.Workspace` 與實際工具呼叫路徑。
3. #2（F-1 CLI/TUI）：待 #4 完成基礎工具閉環後排入；需先將 Go module 基線由 1.22 同步至 1.25.x 並引入 Bubble Tea v2，另行實作授權。
4. 用 `maze-spec-to-issues` 或等效流程，把 ADR-002 的「後續需要」拆成具體 Issues：(a) 未安裝 Git Bash 的 Windows VM/runner 補測 Gate 0、(b) Taylor RPC client 的 `bash` command allowlist/lint 防線設計與實作、(c) Route B 正式整合（取代 #8/#9 原本的 Provider Adapter / Agent Loop 範圍）。
5. 檢視 `docs/spec.md` §5（架構與公開介面）、§9（Contract），依 ADR-002 修訂範圍（哪些改由 Pi 提供、哪些仍需 Go 自建的 Taylor tools 邊界），完成後再排入 #8/#9 後續工作。

## 阻塞與待決策

- Route B 正式整合範圍尚未拆成 Issues；`docs/spec.md` §5／§9 的詳細內容需在拆 Issue 前先決定去留。
- 無 Alpha 1 硬阻塞；`docs/spec.md` §16 的 Open Questions 依各自裁決前行為處理。

## 參考

- `docs/spec.md` §4～§6、§8～§16
- `docs/adr/ADR-002-pi-agent-runtime.md`
- `MAZE_PROJECT.md`

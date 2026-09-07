# Brunel — 下一步行動

> 最後同步：2026-09-07

## 下一個 Session 目標

#5（PR #26）、#7（PR #27）與 #8（本次，PR 待建立）的核心邏輯都已完成並等待 review／合併。下一個 Session 應先確認三者的 review 結果並視需要修正；合併後即解除 #4（F-3 工具，需接上 `internal/filetools`／`internal/safety`）與 #9（F-8 Agent Loop，需接上 `internal/pirpc` 並實作真正的 Pi RPC 子行程管理與 event 轉譯、`taylor-tools.ts`）的相依阻塞。

## 優先行動

1. 追蹤 #5（PR #26）、#7（PR #27）與 #8 的 PR review：處理任何 review 意見；合併後同步關閉對應 Issue 並更新 STATUS.md／NEXT_ACTION.md。
2. #4（F-3：8 工具 schema 與安全入口接線）：待 #7 合併後排入，把 `internal/filetools`（#5）與 `internal/safety`（#7）接上 `workspace.Workspace` 與實際工具呼叫路徑。
3. #9（F-8 Agent Loop／EventSink／context，需依 v1.3 §4 矩陣重新拆解範圍）：待 #4、#8 合併後排入，實作 `taylor-tools.ts`、Pi RPC 子行程啟動與生命週期管理、RPC event → `agent.Event` 轉譯，並落實 INV-9（禁止送出 `bash` RPC command，`internal/pirpc` 已建立 TC-PIRPC-001 原始碼掃描保護）。
4. #2（F-1 CLI/TUI）：待 #4 完成基礎工具閉環後排入；需先將 Go module 基線由 1.22 同步至 1.25.x 並引入 Bubble Tea v2，另行實作授權；同時需提供 `safety.Approver` 的 TUI／純文字實作。
5. 用 `maze-spec-to-issues` 或等效流程，把 ADR-002 的「後續需要」拆成具體 Issues：(a) 未安裝 Git Bash 的 Windows VM/runner 補測 Gate 0（OQ-9）、(b) `taylor-tools.ts` 實作（與 #9 合併排入或獨立拆分，視範圍大小裁決）。

## 阻塞與待決策

- Route B 正式整合範圍尚未拆成 Issues；`docs/spec.md` §5／§9 的詳細內容需在拆 Issue 前先決定去留。
- 無 Alpha 1 硬阻塞；`docs/spec.md` §16 的 Open Questions 依各自裁決前行為處理。

## 參考

- `docs/spec.md` §4～§6、§8～§16
- `docs/adr/ADR-002-pi-agent-runtime.md`
- `MAZE_PROJECT.md`

# Brunel — 決策紀錄

> 格式：每條決策包含「時間、決策內容、原因、影響範圍」；最新決策置頂。

## 決策紀錄

### 2026-08-12 — 放棄零依賴單檔 exe 需求，採用 Pi 作為 Model-facing Agent Runtime（Route B）

**決策**：放棄 ADR-001 硬需求 (a)「乾淨 Windows x64 環境下載單一 `brunel.exe` 即可執行，不需預裝任何 runtime」（spec.md G-1 現行文字為「無需預裝 Go 或 Node.js」，兩者皆與本決策衝突）。改採 [ADR-002](docs/adr/ADR-002-pi-agent-runtime.md)：以 [earendil-works/pi](https://github.com/earendil-works/pi) 作為 model-facing Agent Runtime（Provider abstraction、Agent Loop、Tool-call lifecycle），透過 RPC 模式被 Go Host 呼叫；Go 繼續持有 Workspace boundary、Safety 決策模型、PowerShell 7 執行器與 Windows Job Object 這幾項 Host-only authority。ADR-001 標記為部分 Superseded，硬需求 (b)（Job Object）、Go 1.25.x／Bubble Tea v2 工具鏈基線與程序控制部分維持有效。

**原因**：Issue #24 的 5-Gate Spike（分支 `agent/pi-spike-issue-24`，已 push 至遠端、未合併，可拋棄原型）證實 Gate 1（Model Tool Authority）、Gate 3（Windows Process Authority）、Gate 4（Dual Runtime Complexity）皆 Pass，Gate 2（RPC Side Channel）Pass 但需 Taylor RPC client 自建 `bash` command 禁止清單；Gate 0（Windows Runtime Requirement）因測試機已裝 Git Bash 只完成 Partial。使用者評估後認為「零依賴單檔 exe」在現實中本來就難以完全達成（Brunel 本身已要求 `pwsh` 7 另外安裝），繼續以此為硬約束不合理；放棄該需求後，原本選擇 Go native 而非 Node/Pi 的兩個理由之一（零依賴）不再成立，繼續自建多供應商 Agent Loop 的長期維護負擔不再有相抵收益。

**影響範圍**：`docs/adr/ADR-001-runtime-language.md`（狀態部分 Superseded）、新增 `docs/adr/ADR-002-pi-agent-runtime.md`、`docs/spec.md` G-1 與修訂記錄（標註待正式修訂，§5／§9 等架構章節留待 Route B 整合設計完成後再修）、`STATUS.md`、`NEXT_ACTION.md`。直接受影響的既有 Issue：#8（F-7 Provider Adapter）、#9（F-8 Agent Loop/EventSink/context）需重新檢視是否仍以 Go 自建。後續需要：(1) 在未安裝 Git Bash 的 Windows VM/runner 補測 Gate 0；(2) 設計並實作 Taylor RPC client 的 `bash` command allowlist/lint 防線；(3) 將 Route B 的正式整合工作拆解為 GitHub Issues。

**狀態**：確認

---

### 2026-07-14 — Alpha 1 v1.2 收斂安全、TUI、完成報告與 benchmark 邊界

**決策**：安全定位改為事故防護與 `AUTO`／`CONFIRM`；Alpha 1 採 Go 1.25.x + Bubble Tea v2 薄型 TUI；CompletionReport 只記客觀事實；benchmark runner 移回 Alpha 4；`docs/spec.md` 合併重複契約為 v1.2 單一來源。

**原因**：原安全模型嘗試精細分類任意 PowerShell，複雜度高但無法提供相稱保證；完成證據狀態機只能檢查模型填寫的字串；benchmark runner 與產品提案的 Phase 邊界衝突。薄型 TUI 則改善互動透明度，但必須與 agent core 解耦。

**影響範圍**：`docs/spec.md`、Go 工具鏈 ADR、Alpha 1 Issue 範圍、安全與 completion 介面、CLI／TUI 架構及測試計畫。

**狀態**：確認

---

### 2026-07-13 — 建立公開 GitHub repository 與 Maze 工作流

**決策**：以 `bext1998/brunel` 作為公開 repository；使用 GitHub Issues 與 spec-to-issues，採 `priority: P1`、`type: bug` 結構化標籤，預設指派 `bext1998`，並允許建立缺少的標籤。Coding Agent 使用 Codex 與 Claude Code。

**原因**：讓 Alpha 1 的需求、驗收條件與實作進度可追蹤，並提供跨 Coding Agent 的一致專案定位。

**影響範圍**：GitHub repository 設定、`MAZE_PROJECT.md`、`AGENTS.md`、狀態與後續 Issue 工作流。

**狀態**：確認

---

<!-- 新決策按時間順序追加於最上方。規格內既有決策以 docs/spec.md 修訂記錄為準，不在此重複宣告。 -->

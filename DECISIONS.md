# Brunel — 決策紀錄

> 格式：每條決策包含「時間、決策內容、原因、影響範圍」；最新決策置頂。

## 決策紀錄

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

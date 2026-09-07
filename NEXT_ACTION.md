# Brunel — 下一步行動

> 最後同步：2026-09-08

## 下一個 Session 目標

ADR-002 的架構轉向已拆解完成（#29／#30 拆出、#8／#9 改標題與範圍），`docs/spec.md` 修訂至 v1.3.1（INV-6 best-effort、OQ-10）。#5（F-4，`internal/filetools`）與 #7（F-6，`internal/safety`）的核心實作已分別經 PR #26／PR #27 合併至 `main`。下一步回到 Alpha 1 實作前線，優先推進現已無阻塞的 #4，並行處理 #30、#8。

## 優先行動

1. 實作 #4（F-3：8 工具固定 schema 與安全入口接線）：現已無阻塞（#5／#7 核心已合併）。把 `internal/filetools`（#5）與 `internal/safety`（#7）接上 `workspace.Workspace` 與實際 8 工具呼叫路徑；每次 I/O 前經 `safety.Gate`；§5.4 工具 schema snapshot。完成後解除 #29、#14、部分 #2。
2. 實作 #30（INV-9）：無阻塞、範圍小。`internal/pirpc` 原始碼禁止 `{"type":"bash"}` literal 的 lint／AST 檢查，接進 `.github/workflows/ci.yml`，補 `TC-PIRPC-001`。
3. 實作 #8（F-7）：無阻塞（#12 已完成）。透傳 `--provider`／`--model` 給 Pi RPC 啟動參數、Credential Manager 憑證注入 Pi 子行程、轉譯 Pi 回報的 provider 層錯誤；不自行 SSE／probe／retry。
4. 收尾 #5／#7 的 Issue：核心已合併，wiring 由 #4 承接；確認 #5／#7 的 GitHub Issue 是否隨 #4 一併關閉或先行關閉。
5. 相依鏈：#4 完成後解除 #29；#9（Pi RPC 橋接）待 #4／#8／#29／#30 全部完成才排入；#2（含 `safety.Approver` 的 TUI／純文字實作）待 #9；#11 待 #9；#14 待 #4／#9；#22 待多項。
6. 規劃 #2 前將 Go module 基線由 1.22 同步至 1.25.x 並引入 Bubble Tea v2；此項需另行實作授權，與 Route B 無關（Host 層仍是 Go）。

## 阻塞與待決策

- 無 Alpha 1 硬阻塞；`docs/spec.md` §5／§9 的 Route B 修訂已於 v1.3 完成。
- Gate 0（物理上無 Git Bash 的環境）補測：依使用者裁決不另建 Issue，維持 ADR-002 現況——Git for Windows 為已文件化安裝依賴，spec OQ-9 視為接受風險、不驗證。
- OQ-8（Pi 版本釘選與升級前 Gate 重跑政策）尚未建 Issue；升級 Pi 版本前需重跑對應 Gate 等價測試。
- OQ-10（檔案寫入的 sub-millisecond rename 競態）：Alpha 1 已裁決接受為 best-effort（見 DECISIONS.md 2026-09-08）；Alpha 3「單一 writer」時重評，屆時若 Brunel 內部出現併發 writer 需加 path-keyed 序列化。
- spec §16 其餘 Open Questions 依各自裁決前行為處理。

## 參考

- `docs/spec.md` §4～§6、§8～§16
- `docs/adr/ADR-002-pi-agent-runtime.md`
- `MAZE_PROJECT.md`

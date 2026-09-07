# Brunel — 下一步行動

> 最後同步：2026-09-08

## 下一個 Session 目標

ADR-002 的架構轉向已拆解完成：#8／#9 依 v1.3 §4 矩陣改標題與範圍，並拆出 #29（`taylor-tools.ts` extension 與 `brunel.exe --taylor-tool` 派工）與 #30（INV-9 `bash` command 禁令與 CI lint）。`docs/spec.md` §5／§9 的 Route B 修訂已在 `7e9e01e`（v1.3）完成。下一步回到 Alpha 1 實作前線，優先推進沒有開放阻塞的 #7、#8、#30（#5 已在 PR #26 review 中）。

## 優先行動

1. 實作 #7（F-6 AUTO／CONFIRM 事故防護與 Approver）：無阻塞，且是 #2／#4／#9 的共同前置，優先解除。依 spec §5.2、§6、§9～§14。
2. 收尾 #5（F-4 stale-read）：核心實作已在 `internal/filetools` 完成、PR #26 review 中（全檔 SHA-256、`expected_hash` 前置條件、精確 hunk 套用、暫存檔＋原子替換）；跟進 review 意見併入。實際接上 `workspace.Workspace` 與 §5.4 工具 schema 屬 #4。
3. 實作 #30（INV-9）：無阻塞、範圍小。`internal/pirpc` 原始碼禁止 `{"type":"bash"}` literal 的 lint／AST 檢查，接進 `.github/workflows/ci.yml`，補 `TC-PIRPC-001`。
4. 實作 #8（F-7）：無阻塞（#12 已完成）。透傳 `--provider`／`--model` 給 Pi RPC 啟動參數、Credential Manager 憑證注入 Pi 子行程、轉譯 Pi 回報的 provider 層錯誤；不自行 SSE／probe／retry。
5. 相依鏈：#7 完成後解除 #4（F-3 8 工具）；#4 完成後解除 #29。#9（Pi RPC 橋接）待 #4／#8／#29／#30 全部完成才排入；#2 待 #7／#9，#11 待 #9，#14 待 #4／#9，#22 待多項。
6. 規劃 #2 前將 Go module 基線由 1.22 同步至 1.25.x 並引入 Bubble Tea v2；此項需另行實作授權，與 Route B 無關（Host 層仍是 Go）。

## 阻塞與待決策

- 無 Alpha 1 硬阻塞；`docs/spec.md` §5／§9 的 Route B 修訂已於 v1.3 完成。
- Gate 0（物理上無 Git Bash 的環境）補測：依使用者裁決不另建 Issue，維持 ADR-002 現況——Git for Windows 為已文件化安裝依賴，spec OQ-9 視為接受風險、不驗證。
- OQ-8（Pi 版本釘選與升級前 Gate 重跑政策）尚未建 Issue；升級 Pi 版本前需重跑對應 Gate 等價測試。spec §16 其餘 Open Questions 依各自裁決前行為處理。

## 參考

- `docs/spec.md` §4～§6、§8～§16
- `docs/adr/ADR-002-pi-agent-runtime.md`
- `MAZE_PROJECT.md`

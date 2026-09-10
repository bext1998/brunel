# Brunel — 下一步行動

> 最後同步：2026-09-11

## 下一個 Session 目標

#4（F-3，`internal/tools`）、#29（`taylor-tools.ts` ＋ `cmd/brunel --taylor-tool`）核心分別經 PR #33／PR #37 合併至 `main`；#30（INV-9 AST 檢查）經 PR #36 完成並 **CLOSED**。#5／#7／#8 核心早先已合併。`cmd/brunel`（repo 第一個可執行檔，僅 `--taylor-tool` 路徑）已在 `main`。可執行前線收斂為 **#9**（Pi RPC 橋接，相依核心均已落地）與獨立的 **#31**。

## 優先行動

1. 實作 **#9**（F-8：Pi RPC 子行程 ↔ `Agent`／`EventSink` 橋接）：相依 #4／#8／#29 核心與 #30 皆已落地，現可排入。啟動並管理 `pi --mode rpc --no-builtin-tools --no-extensions -e taylor-tools.ts --no-session --provider <p> --model <m>` 子行程；把 RPC event（`message_update`、`tool_execution_start`／`end`、`agent_end`…）轉譯為既有 `agent.Event`／`EventKind`；注入 AGENTS.md；轉譯後 event append 至 Brunel 自己的 `events.jsonl`。橋接前須決定 `workspace_diff` 語意 vs §8 CompletionReport 的 `Diff`（git-only 且只看未暫存變更會漏新建／staged 檔）——見 [#9 review 交接註解](https://github.com/bext1998/brunel/issues/9#issuecomment-5620455153)。此路徑打通後即可正式判定 AC-2／AC-6／AC-7／AC-9～AC-11／AC-14。
2. **#31**（§6.2 classifier 漏判／過度確認強化，PR #27 事後審查）：無硬阻塞、獨立於 #9，建議 Alpha 1 發布前完成。
3. F-3 收尾（不阻塞前線）：處理 [#4 review 註解](https://github.com/bext1998/brunel/issues/4#issuecomment-5620453317) 的 10 項 minor／nit（`workspace_diff` timeout 混碼＋stderr 汙染＋漏 staged／untracked、`search_text` 無上限、`max_depth:0` 文件化、巢狀 null coerce、缺失路徑錯誤碼、3 個測試缺口、`STATUS.md` 用詞）。另 #29 的 `taylor-tools.ts` 仍待真實 TS 編譯／Pi runtime 驗證、`typebox` 依賴重複、per-call re-bind 無 session root identity（INV-5）——皆屬 #9。
4. INV-9 CI 防線的 follow-up（PR #36 review nit）：`go test -run '^TestTCPIRPC001$'` 在守衛測試被刪／改名時退出 0，靜默失去防線；可另讓 CI 斷言該測試存在。
5. 收尾 #4／#5／#7／#8／#29 的 Issue：核心均已合併，剩餘工作由 #9／#2／#31 承接；確認各 Issue 是否隨 #9 一併關閉或先行關閉。
6. 相依鏈：#2（含 `safety.Approver` 的 TUI／純文字實作）待 #9；#11 待 #9；#14 待 #9；#22 待多項。
7. 規劃 #2 前將 Go module 基線由 1.22 同步至 1.25.x 並引入 Bubble Tea v2；此項需另行實作授權，與 Route B 無關（Host 層仍是 Go）。

## 阻塞與待決策

- 無 Alpha 1 硬阻塞；`docs/spec.md` §5／§9 的 Route B 修訂已於 v1.3 完成。
- Gate 0（物理上無 Git Bash 的環境）補測：依使用者裁決不另建 Issue，維持 ADR-002 現況——Git for Windows 為已文件化安裝依賴，spec OQ-9 視為接受風險、不驗證。
- OQ-8（Pi 版本釘選與升級前 Gate 重跑政策）尚未建 Issue；升級 Pi 版本前需重跑對應 Gate 等價測試。
- OQ-10（檔案寫入的 sub-millisecond rename 競態）：Alpha 1 已裁決接受為 best-effort（見 DECISIONS.md 2026-09-08）；Alpha 3「單一 writer」時重評，屆時若 Brunel 內部出現併發 writer 需加 path-keyed 序列化。
- 公開錯誤不含 secret 的最終責任邊界：Pi 自行探索、Brunel 從未持有的 provider key 若被 Pi 回顯於錯誤訊息，`internal/pirpc` 只能做啟發式遮罩（`internal/redact` 已能在 Brunel 持有實際值時精確替換）。責任歸屬需在 spec 或 #9 定義。
- spec §16 其餘 Open Questions 依各自裁決前行為處理。

## 參考

- `docs/spec.md` §4～§6、§8～§16
- `docs/adr/ADR-002-pi-agent-runtime.md`
- `MAZE_PROJECT.md`

# Brunel — 下一步行動

> 最後同步：2026-09-10

## 下一個 Session 目標

#4（F-3，`internal/tools`）核心已經 PR #33 合併至 `main`（含兩個隔離 Sonnet subagent 審查後的 major 修正）；PR #34（CI 矩陣標籤）、PR #32（`CLAUDE.md`）亦已合併。#5／#7／#8 核心早先已合併。可執行前線為 **#29**（現已解除，#4 核心已合併）與 **#30**（無阻塞、範圍小），兩者皆 block #9；建議並行推進。

## 優先行動

1. 實作 **#29**（`taylor-tools.ts` extension ＋ `brunel.exe --taylor-tool` 子行程派工）：現已解除阻塞。以 `pi.registerTool()` 註冊 8 個 Taylor tools；Go 端新增 `--taylor-tool <name>` 進入點——讀 stdin JSON → `internal/tools`（PR #33 已就緒）→ 輸出結構化結果與終態；參數／結果 JSON 契約與 8 工具 schema snapshot（PR #33 已起頭 `internal/tools/testdata/params_schema.json`，此處擴為含 prompt 描述的完整契約）；沿用 Issue #24 Gate 3/4 的子行程模式。`list_files` 的 `max_depth`（含 `0`=不限深度）語意須寫進工具說明文字。見 [#29 review 交接註解](https://github.com/bext1998/brunel/issues/29#issuecomment-5620454751)。
2. 實作 **#30**（INV-9）：無阻塞、範圍小、可與 #29 並行。`internal/pirpc` 原始碼禁止 `{"type":"bash"}` literal 的 lint／AST 檢查，接進 `.github/workflows/ci.yml`，補正式 `TC-PIRPC-001` 正反例（PR #28 目前只有過渡性 literal tripwire，不作為 INV-9 驗收證據）。
3. #29／#30 皆完成後排入 **#9**（Pi RPC 子行程 ↔ `Agent`／`EventSink` 橋接，#4／#8 已就緒）。橋接前須處理 `workspace_diff` 語意 vs §8 CompletionReport 的 `Diff`（git-only 且只看未暫存變更會漏新建／staged 檔）——見 [#9 review 交接註解](https://github.com/bext1998/brunel/issues/9#issuecomment-5620455153)。
4. F-3 收尾（不阻塞前線）：處理 [#4 review 註解](https://github.com/bext1998/brunel/issues/4#issuecomment-5620453317) 的 10 項 minor／nit（`workspace_diff` timeout 混碼＋stderr 汙染＋漏 staged／untracked、`search_text` 無上限、`max_depth:0` 文件化、巢狀 null coerce、缺失路徑錯誤碼、3 個測試缺口、`STATUS.md` 用詞）。
5. 收尾 #4／#5／#7／#8 的 Issue：核心已合併，#5／#7 已由 #4 接上工具呼叫路徑、#8 橋接由 #9 承接；確認各 Issue 是否隨 #29／#9 一併關閉或先行關閉。
6. #31（§6.2 classifier 漏判／過度確認強化，PR #27 事後審查）：無硬阻塞，建議 Alpha 1 發布前完成；獨立於 #29／#30 前線。
7. 相依鏈：#9 待 #29／#30；#2（含 `safety.Approver` 的 TUI／純文字實作）待 #9；#11 待 #9；#14 待 #9；#22 待多項。
8. 規劃 #2 前將 Go module 基線由 1.22 同步至 1.25.x 並引入 Bubble Tea v2；此項需另行實作授權，與 Route B 無關（Host 層仍是 Go）。

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

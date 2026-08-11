# ADR-002 採用 Pi 作為 Model-facing Agent Runtime（Route B）

## 狀態

Accepted

## 日期

2026-08-12

## 脈絡

ADR-001 訂下兩項硬需求：(a) 乾淨 Windows x64 環境下載單一 `brunel.exe` 即可執行，不需預裝任何 runtime；(b) 導入 Windows Job Object 以實作程序樹終止與資源上限。需求 (a) 是選擇 Go native、放棄採用既有 Agent Runtime（如 [earendil-works/pi](https://github.com/earendil-works/pi)）的主要理由。

[Issue #24](https://github.com/bext1998/brunel/issues/24) 針對「Pi 能否在 Windows 上作為 Taylor Agent Runtime，同時不破壞 Brunel 既有 Workspace Authority、PowerShell execution、Job Object 與 Completion Evidence 設計」做了 5 個 Gate 的 Spike 驗證（分支 `agent/pi-spike-issue-24`，可拋棄原型，未合併 main），完整報告見 [`spikes/pi-compatibility/REPORT.md`](../../spikes/pi-compatibility/REPORT.md)（該分支）。Gate 結果：

| Gate | 結果 | 摘要 |
|---|---|---|
| Gate 0 Windows Runtime Requirement | Partial | RPC `get_state` round-trip 不依賴 bash；`bash` RPC command 本身依賴可用 shell。測試機已裝 Git Bash，無法完整驗證「物理上完全沒有 bash」情境 |
| Gate 1 Model Tool Authority | Pass | 自訂 `taylor_*` 工具為唯一副作用路徑，模型對停用/不存在工具僅得到協定層 unavailable |
| Gate 2 RPC Side Channel | Pass（有但書） | `--no-builtin-tools` 不會停用 host-level RPC `{"type":"bash"}` command；Model Authority 需由 Taylor RPC client 自行禁止發送此 command 才成立 |
| Gate 3 Windows Process Authority | Pass | Go Host 的 Job Object 與模擬的 Pi 行程樹完全獨立；Pi abort/crash 不影響 Job Object 清理 |
| Gate 4 Dual Runtime Complexity | Pass | 完成 Pi RPC → TypeScript extension → Go Host → PowerShell 的 read/modify/test/diff/completion 全流程；跨邊界 glue code 約 26 行 |

在此次決策評估中，使用者（產品負責人）進一步決定放棄需求 (a)：「零依賴、單檔 exe」在現實中本來就難以完全做到（Brunel 本身已要求使用者另外安裝 `pwsh` 7，並非 Windows 內建），繼續以此為硬約束並不合理。

## 決策

放棄 ADR-001 硬需求 (a)。採用 Route B：Pi 作為 model-facing Agent Runtime（Provider abstraction、Agent Loop、Tool-call lifecycle），透過 RPC 模式（`pi --mode rpc`）被 Go Host 呼叫；Go 繼續持有 Workspace boundary、Safety 決策模型、PowerShell 7 執行器與 Windows Job Object 這幾項 Host-only authority，經由自訂 Taylor tools（`taylor_read`／`taylor_write`／`taylor_run`）暴露給 Pi 內的模型。

使用者需另外安裝 Node.js/npm（或內嵌的 Pi 執行環境）與 Git for Windows（提供 Git Bash），視為與 `pwsh` 7 同等級的已知部署依賴，不再視為需要規避的阻礙。

## 理由

1. 需求 (a) 已由使用者明確放棄；ADR-001 選擇 Go native 而非 Node/Pi 的兩個理由之一因此不再成立。
2. 需求 (b)（Job Object）不受影響：Gate 3/4 證實 Job Object 職權可留在獨立的 Go Host 行程，與 Pi 的行程樹完全脫鉤，不需要 Pi 存活或配合才能清理。
3. Gate 1/2/3/4 的實測證據支持「Pi 可作為 model-facing shell，只要 Taylor 自建的 RPC client 遵守不主動發送 `bash` RPC command」這個條件式 Model Authority。
4. Gate 4 顯示跨語言邊界的 glue code 量體很小（~26 行），不是「龐大 Go↔TypeScript bridge」等級的隱藏成本。
5. 對單一維護者而言，繼續自行追蹤多家模型 API 協定的長期負擔，在放棄需求 (a) 後不再有相抵的零依賴利益，維持 Route A 的邊際效益轉為負值。

## 承擔的代價

1. 使用者需另外安裝 Node.js/npm 與 Git for Windows；`brunel.exe` 不再是單檔零依賴發布。README／安裝文件需明確揭露此依賴，比照現有 `pwsh` 7 要求的揭露方式。
2. Taylor RPC client（Go）需自行實作並維護 `{"type":"bash"}` 送出的 allowlist/lint 防線；此防線不是 Pi 提供的保證，是 Brunel 自己要建立與測試的 invariant（見 Gate 2 建議事項）。
3. Gate 0 的「物理上無 Git Bash」情境仍未完整驗證；此風險被接受為與 Git Bash 依賴同等級，不再視為阻擋因素，但跨邊界的 cancellation 語意、streaming、retry/backoff 等仍是未驗證的整合風險（見 Gate 4 報告），需在正式整合時補測。
4. ADR-001 原本承擔的「SSE streaming 解析、agent loop、tool schema 產生器需自行實作」的代價，隨這次決策大部分轉移給 Pi；但需重新驗證 Pi 的 provider 涵蓋範圍是否仍滿足 Brunel「只綁 OpenRouter」或後續 provider 需求（本 Spike 未涵蓋）。

## 推翻條件

- 若正式整合後發現 Taylor RPC client 的 `bash` command allowlist 防線在實務上無法可靠維持（例如 Pi 版本更新後出現新的 host-level side channel），需重新評估 Model Authority 是否仍能成立。
- 若跨語言邊界的 cancellation／streaming／retry 整合成本，在正式串接後遠高於 Gate 4 的粗估（~26 行 glue code），且已成為龐大 Go↔TypeScript bridge，則回頭比較 Route C（更深度整合／自建）的相對成本。
- 若使用者重新要求恢復零依賴單檔發布（需求 (a) 復活），本決策需重新評估。

## 被否決的選項與原因

- **維持 Route A（繼續自建 Agent Loop / Provider 抽象層）**：在需求 (a) 已放棄的前提下，繼續自行維護多供應商協定的長期負擔，不再有零依賴的對等收益，是本次決策要解決的核心問題本身，故否決。
- **Route C（Fork Pi 或更深度整合）**：Issue #24 明確排除本次評估此選項，除非 Gate 1/2/3 證實 SDK/RPC/Extension 三種公開介面確實無法滿足需求；本次 Gate 1/2/3 均 Pass，不成立 Fork 的前提，故不評估。

## 參考

- [Issue #24](https://github.com/bext1998/brunel/issues/24)
- `agent/pi-spike-issue-24` 分支（未合併，可拋棄原型）：`spikes/pi-compatibility/REPORT.md`、`PLAN.md`
- [ADR-001](ADR-001-runtime-language.md)（被本文件部分取代）

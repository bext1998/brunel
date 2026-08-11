# Pi Compatibility Spike — Report

Issue：[#24 Pi Compatibility Spike](https://github.com/bext1998/brunel/issues/24)

執行 branch：`agent/pi-spike-issue-24`

執行 worktree：`D:\AgentCoding\Brunel-pi-spike`

Pi：`@earendil-works/pi-coding-agent@0.84.1`

## Gate verdicts

| Gate | Verdict | 結論 |
|---|---|---|
| Gate 0 — Windows Runtime Requirement | **Partial** | Pi RPC 在 baseline 與 clean PATH/隔離設定下都能 `get_state` round-trip；但本機仍安裝 Git Bash，無法在不改系統的前提下完成「實體不存在 Git Bash」的完整環境測試。不存在 shell 的隔離 surrogate 顯示只有 `bash` command 階段失敗，RPC 啟動仍成功。 |
| Gate 1 — Model Tool Authority | **Pass** | openai-codex OAuth 實際模型回合只執行 `taylor_read`、`taylor_write`、`taylor_run`；模型對 `bash`、`read`、`write`、`nonexistent_tool` 回報 unavailable，沒有任何非 Taylor tool execution 或副作用。 |
| Gate 2 — RPC Side Channel | **Pass** | 直接送 `{"type":"bash","command":"echo gate2"}` 在有與無 `--no-builtin-tools` 時都成功，回傳 `data.output = "gate2\n"`。 |
| Gate 3 — Windows Process Authority | **Pass** | 新增的 Pi abort/crash 隔離測試與 CLI round-trip 通過；Go Job Object 清理不依賴 Pi surrogate。 |
| Gate 4 — Dual Runtime Complexity | **Pass** | 真實模型完成 read → modify → test → diff → completion；`taylor_run` 經 `taylorrun.exe` 進入既有 `internal/exec`。 |

## Phase 0 — Recon

版本與 credential 快照見 [`environment.md`](environment.md)。

- Node `v24.16.0`、npm `11.18.0`、Git `2.53.0.windows.2`、PowerShell `7.6.4`。
- `OPENROUTER_API_KEY` 存在；其他四個計畫列出的環境變數 absent。
- `pi auth check --provider openai-codex --model gpt-5.6-luna --json --no-refresh` 回傳 `status=ready`、`authType=oauth`。
- OpenRouter 實際呼叫曾回傳 `404 No endpoints available matching your guardrail restrictions and data policy`，記錄於 [`gate1-model-authority.jsonl`](gate1-model-authority.jsonl)。這不是 credential 缺失；後續改用已 ready 的 openai-codex OAuth 完成 Gate 1/4。
- 未將任何 credential 值寫入 repository。

## Gate 0 — Windows Runtime Requirement

### 實測結果

1. 有 Git Bash baseline：[`gate0-baseline.jsonl`](gate0-baseline.jsonl) 收到：

   ```json
   {"id":"state","type":"response","command":"get_state","success":true}
   ```

2. clean 子程序環境：[`gate0-no-git-bash-v2.jsonl`](gate0-no-git-bash-v2.jsonl) 記錄 42 個 PATH entries 移除 1 個、剩餘可見 `bash.exe` entries 為 0，並使用隔離 `PI_CODING_AGENT_DIR`；`get_state` 仍成功。

3. 重要環境限制：Pi 官方實作除了 PATH 外，還直接檢查 `ProgramFiles\Git\bin\bash.exe`。本機 Git Bash 存在，而且 Windows 子程序的 `ProgramFiles` 不能由本測試可靠覆寫，因此上述 clean-PATH run 仍可能使用已安裝的 known-location Git Bash。這不是完整的 physical “no Git Bash” 測試。

4. 為確認 failure scope，使用隔離暫存 settings 強制指定不存在的 `C:\__brunel_missing__\bash.exe`，結果見 [`gate0-shell-unavailable-surrogate.jsonl`](gate0-shell-unavailable-surrogate.jsonl)：

   - `get_state`：`success=true`。
   - `bash`：`success=false`，錯誤為 `Custom shell path not found`。

因此可證實 Pi process/RPC 啟動不依賴 bash；bash shell 是執行 bash host command 時的需求。要把 Gate 0 從 Partial 補成完整 Pass，需要一台未安裝 Git Bash 的 Windows VM/runner，或使用者明確允許可回復的系統級隔離措施。未修改本機 PATH、Git Bash 安裝或系統設定。

## Gate 1 — Model Tool Authority

工具 extension 位於 [`taylor-tools.ts`](taylor-tools.ts)：`taylor_read` line 23、`taylor_write` line 40、`taylor_run` line 58，註冊於 lines 107–109。

完整 openai-codex RPC 證據在 [`gate1-model-authority-openai-codex-v2.jsonl`](gate1-model-authority-openai-codex-v2.jsonl)。模型實際造成的副作用與 tool execution：

- `taylor_read`：讀取 baseline，最後再次讀取修改後內容。
- `taylor_write`：將 `gate1 modified by Taylor` 加入 scratch file。
- `taylor_run`：透過 PowerShell 7 讀取檔案。
- 沒有 `bash`、`read`、`write`、`edit`、`grep`、`find` 或 `nonexistent_tool` 的 `tool_execution_start/end`。
- 模型在 forbidden prompt 中明確回報四個名稱都是 `not attempted — unavailable`，並說明只有 exposed Taylor tools 可被呼叫；沒有靜默轉發。

這證實測試模型能在 custom tools-only registry 下完成副作用閉環。限制是模型正確拒絕了未暴露工具，因此沒有產生一個可觀察的「任意未知 tool call → 協定錯誤 response」事件；該 negative protocol branch 不是 Pi 文件所提供的 RPC client input。這不影響本次 Pass 的實際 side-effect authority 結果，但後續正式 client 應補一個專門的 negative tool-call harness。

## Gate 2 — RPC Side Channel

兩組啟動參數的證據：

- [`gate2-no-builtin-tools.jsonl`](gate2-no-builtin-tools.jsonl)：`bash_execution_update` 回傳 `gate2`，response `success=true`、`exitCode=0`。
- [`gate2-with-builtin-tools.jsonl`](gate2-with-builtin-tools.jsonl)：結果相同。

結論：`--no-builtin-tools` 不會停用 host-level RPC `bash`。因此「Taylor 自己撰寫的 RPC client 只要不主動發送 `{"type":"bash"}`，Model Authority 即可成立」是**條件式成立**；它不是 Pi flag 自動保證的 invariant。RPC client 必須自己遵守不發送此 command。

建議的 architecture invariant（本 Spike 未接入 CI）：

1. 將正式 Go RPC client 的可發送 command 建立 allowlist，不包含 `bash`。
2. 對 client source 做 CI lint/AST check；最小版可檢查 JSON literal 中 `type` 值不得為 `bash`。Gate 2 fixture 與測試程式需放在明確排除的 test-only path，避免把測試本身當成 production violation。
3. 對 client 收到的 response 保留 `command`、`success`、`error` 欄位，禁止 unknown command 被 fallback 到 bash。

## Gate 3 — Windows Process Authority

新增最小 CLI：[`cmd/taylorrun/main_windows.go`](cmd/taylorrun/main_windows.go)。它只呼叫公開的 `internal/exec.NewRunner()`/`Runner.Run()`，將 command 直接交給既有 PowerShell 7 + Job Object 執行器；沒有複製或重寫 Job Object 邏輯。

新增 Spike 專屬測試：[`taylorrun_windows_test.go`](cmd/taylorrun/taylorrun_windows_test.go)。

- `TestPiAbortDoesNotOwnGoJob`：取消 Go context 後，Job Object descendant 停止，Pi surrogate heartbeat 持續。
- `TestPiCrashDoesNotOwnGoJob`：kill Pi surrogate 後，Go Job Object descendant heartbeat 持續；之後由 test context 清理 descendant。

Verbose 結果見 [`gate3-go-test.log`](gate3-go-test.log)：兩個新增測試均 PASS，整體 `ok`。

既有覆蓋直接引用 [`internal/exec/runner_windows_test.go`](../../internal/exec/runner_windows_test.go)：

- normal exit：`TestPSRunner_HappyPath` line 129。
- timeout / descendant process tree：`TestPSRunner_Timeout_KillsProcessTree` line 242。
- process-wide concurrent runs：`TestPSRunner_ConcurrentRunsAcrossRunners` line 180。
- job assignment before resume：`TestPSRunner_JobBoundBeforeResume` line 299。
- bounded pipe/termination cleanup：`TestWaitForExitAndOutput` line 43。

CLI `taylorrun.exe -command "Write-Output gate3-cli"` 實際回傳 JSON：`exitCode=0`、`stdout=gate3-cli`。Binary 是 build artifact，已由 [`spikes/pi-compatibility/.gitignore`](.gitignore) 排除，不作為 source deliverable。

## Gate 4 — Dual Runtime Complexity

`taylor_run` 的 Go-host branch 在 [`taylor-tools.ts`](taylor-tools.ts) lines 65–90：

```text
Pi RPC prompt → Node taylor_run → execFile(taylorrun.exe, argv)
→ Go internal/exec → pwsh child + independent Job Object
→ JSON stdout → Node tool result → Pi RPC events
```

實測證據在 [`gate4-dual-runtime.jsonl`](gate4-dual-runtime.jsonl)。模型完成：

1. `taylor_read` 讀取 `gate4 baseline`。
2. `taylor_write` 新增 `gate4 modified through Go host`。
3. `taylor_run` 執行檢查，回傳 `gate4-test-pass`。
4. `taylor_run` 經 Go host 執行 `git diff --no-index`，取得預期 diff。
5. `taylor_read` 再次讀取並回報 completion。

量化證據：

- JSONL 共 596 records。
- 5 次 `tool_execution_start`：`taylor_read` 2、`taylor_write` 1、`taylor_run` 2。
- RPC response 6 次：1 次 `get_state`、5 次 prompt acceptance。
- Pi event 中有 499 次 `message_update`，另有 20 次 `message_start` 與 20 次 `message_end`；Pi streaming event 可被 Node client 觀察，但本 Spike 的 Go CLI 是同步 JSON stdout，不做 partial output streaming。
- `taylor-tools.ts` 共 102 行；Go CLI 82 行；兩個 Go process-authority tests 共 201 行。新增的 Go-host glue branch 約 26 行，包含 argv 組裝、JSON parse、`errorCode`/`error` 映射，沒有手工 retry。
- 本次實際跨邊界的 payload 只有一種 host result JSON shape；Node 以 `payload.error` 判斷失敗，成功時把 `stdout`/`stderr` 轉成 Pi tool result。

未在本 Gate 量測的項目：Node → Go 的 cancellation 語意、Go stdout partial streaming、retry/backoff、session restart。`execFile` 有傳 `AbortSignal`，但 `taylorrun.exe` CLI 目前以 `context.Background()` 執行，因此 cancellation propagation 尚未是完整契約；這是後續正式整合的風險，不把它誇大成已驗證能力。

## Authority invariants

### Model Authority

**Conditional pass / not automatic.** Gate 1 實測中，所有成功副作用均來自 `taylor_*`。Gate 2 同時證明 `--no-builtin-tools` 仍保留一條可直接執行的 RPC `bash` side channel，所以 invariant 必須落在 Taylor RPC client 的 allowlist/lint，而不是 Pi 啟動旗標本身。Formal client 尚未存在，建議的 static guard 尚未接 CI。

### Runtime Authority

**Pass。** Gate 3 的兩個 sibling-process tests 證明 Go Host 的 Job Object 與 Pi surrogate process tree 分離；Pi abort/crash 不會替 Go Host 決定 descendant cleanup。Gate 4 也用同一個既有 `internal/exec` 實際完成 Go-host run。

## Decision output

目前證據不是「全部 Gate 完整通過」：Gate 0 的 physical no-Git-Bash condition 受本機環境限制，只完成 Partial。依 Issue 的保守判斷表，現在不能直接把 Route B 宣布為第一候選；暫時應回到 **Route A / pending validation**，直到在沒有 Git Bash 的 Windows VM/runner 補齊 Gate 0。

同時，Gate 1/2/3/4 的證據支持以下較窄結論：Pi 可作 Taylor Runtime 的 model-facing shell，但必須由自有 RPC client 明確禁止 host-level `bash`；Go Host 可維持獨立的 Windows process authority。若 Gate 0 補測成功，Route B 才有完整證據進入候選評估。

## Reproduction commands

在 `D:\AgentCoding\Brunel-pi-spike` 執行：

```powershell
go test -v ./spikes/pi-compatibility/cmd/taylorrun
go build -o .\spikes\pi-compatibility\cmd\taylorrun\taylorrun.exe ./spikes/pi-compatibility/cmd/taylorrun
```

Pi RPC probe 使用 [`rpc-probe.mjs`](rpc-probe.mjs)；Gate 1/4 的 command sequences 分別在 [`gate1-commands.json`](gate1-commands.json) 與 [`gate4-commands.json`](gate4-commands.json)。

## Scope and deviations

- 沒有 fork Pi。
- 沒有修改 `main`、共用 `D:\AgentCoding\Brunel` checkout、`DECISIONS.md`、任何 ADR 或 `docs/spec.md`。
- 沒有 merge 或開 PR。
- 沒有修改系統 PATH、Git Bash 安裝或 Windows Credential Manager；npm global install 是 Phase 0 計劃內的環境準備。
- 實際偏差只有 Gate 0：發現 Pi 的 known-location lookup 使「只清 PATH」不足以模擬無 Git Bash，因此補做了不存在 `shellPath` surrogate，並在本報告保留 Partial，不冒充完整 Pass。

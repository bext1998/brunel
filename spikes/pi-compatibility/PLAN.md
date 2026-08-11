# Pi Compatibility Spike — Implementation Plan (Issue #24)

來源：https://github.com/bext1998/brunel/issues/24
角色：本文件由 claude-brunel（指揮者）撰寫，交付 codex-brunel 實作。

## 0. 邊界重申（來自 Issue Non-goals，實作時不得逾越）

- 不 Fork Pi。
- 不推翻 ADR-001（Go native / CGO_ENABLED=0 / Windows x64 / PowerShell 7 / Job Object）。
- 產出程式碼視為可拋棄原型，**不進 `main`**，存放於獨立 branch + `spikes/pi-compatibility/` 資料夾。
- 範圍僅限 Brunel / Taylor Agent Runtime。
- **禁止捏造結果**：任一 Gate 若因缺少 LLM API credential 或環境限制而無法實測，必須明確標記為 `Blocked`，寫出缺什麼、如何補；不得用推測結果冒充實測結果。

## 1. 隔離策略

同一台機器上，`D:\AgentCoding\Brunel` 這個 worktree 目前被 claude-brunel（本指揮者）與 codex-brunel 共用同一個 checkout。**codex-brunel 必須先建立獨立 git worktree，不得在共用 worktree 內切換 branch**，避免與指揮者的 git 狀態互相干擾：

```powershell
git worktree add ../Brunel-pi-spike -b agent/pi-spike-issue-24
cd ../Brunel-pi-spike
mkdir spikes/pi-compatibility
```

所有程式碼、腳本、記錄一律寫入 `spikes/pi-compatibility/`。此 branch 全程不 merge、不開 PR（除非之後使用者另外指示）。

## 2. 已知環境資訊（指揮者已確認，可直接沿用，不需重查）

- Node v24.16.0、npm 11.18.0、bunx 1.3.11、git 2.53.0.windows.2、pwsh 7.6.4 均已安裝於本機。
- **正確套件名稱**：`@earendil-works/pi-coding-agent`（issue 內 `pi` 一詞易誤植為 `@earendil-works/pi`，該名稱在 npm registry 上 404）。安裝：`npm install -g @earendil-works/pi-coding-agent` 或改用 `npx @earendil-works/pi-coding-agent`。
- 官方文件（`packages/coding-agent/docs/windows.md`）明載：Pi 在 Windows 上依序檢查 `~/.pi/agent/settings.json` 自訂路徑 → Git Bash（`C:\Program Files\Git\bin\bash.exe`）→ PATH 上的 `bash.exe`（Cygwin/MSYS2/WSL）。原文即為「Pi requires a bash shell on Windows」——這是 Gate 0 的重要文件線索，仍須實測驗證是否僅影響 `bash` 內建工具、還是連 RPC 啟動本身都會失敗。
- 已確認存在的 CLI flags：`--no-builtin-tools`/`-nbt`、`--no-extensions`、`--no-skills`、`--no-prompt-templates`、`--no-context-files`/`-nc`、`--no-session`、`--mode rpc`、`--provider`、`--model`、`-e <path>`（載入單一 extension）。
- RPC `bash` command（`{"type":"bash","command":"..."}`）為 host-level command，文件記載於 `packages/coding-agent/docs/rpc.md`，獨立於 `pi.registerTool()` 註冊的工具 registry。
- 自訂工具透過 extension 模組的 `pi.registerTool({ name, label, description, parameters, execute })` 註冊，測試用 `pi -e ./taylor-tools.ts`。
- Brunel 現有 `internal/exec`（`jobobject_windows.go`、`runner_windows.go`、`runner_windows_test.go` 等，PR #20 / Issue #6 已合併）已是可運作的 PowerShell 7 + Windows Job Object 執行器，涵蓋逾時、取消、程序樹終止。Gate 3 應**重用**此套件而非重寫。

## 3. Phase 0 — Recon（進場先做，5–10 分鐘）

- 在新 worktree 內記錄：`node -v`、`npm -v`、`git --version`、`pwsh -v` 版本快照存入 `spikes/pi-compatibility/environment.md`。
- 安裝 `@earendil-works/pi-coding-agent`，記錄實際安裝版本（`pi --version`）。
- 檢查是否有可用的模型 provider credential（不要印出 secret 本身，只回報「有/無」與來源）：
  - 環境變數：`ANTHROPIC_API_KEY`、`OPENAI_API_KEY`、`OPENROUTER_API_KEY`、`GEMINI_API_KEY`
  - `~/.pi/agent/settings.json` 內是否已設定 provider
  - Windows Credential Manager（`cmdkey /list`，比照 spec 既有慣例）
- 若完全沒有可用 credential：Gate 1 與 Gate 4 需要真實模型呼叫的部分將被標記 `Blocked`，但 Gate 0（部分）、Gate 2、Gate 3 不受影響，仍應完整執行。

## 4. Phase 1 — Gate 0：Windows Runtime Requirement

目標：確認沒有 Git Bash 時，Pi 是否仍可啟動並完成一次 RPC round-trip。

1. Baseline（有 Git Bash）：
   ```
   pi --mode rpc --no-builtin-tools --no-extensions --no-skills --no-prompt-templates --no-context-files --no-session
   ```
   啟動後送 `{"type":"get_state"}`，確認能收到 `{"type":"response","command":"get_state",...}`。這一步**不需要**真實模型呼叫，純粹驗證 process 啟動與協定 round-trip，credential 缺失不影響本步驟。
2. 無 Git Bash 情境：**不要修改系統 PATH**。改用子程序層級的乾淨環境變數（例如 PowerShell `Start-Process` 或 Go/Node 子程序時明確組一份不含 `Git\\bin`、不含任何 `bash.exe` 所在目錄的 `PATH`）啟動同一條指令，觀察啟動是否成功、失敗訊息為何。
3. 若步驟 2 啟動失敗：記錄失敗訊息、失敗發生在啟動階段還是特定工具呼叫階段。若啟動仍成功但只有呼叫 `bash` 內建工具時才失敗，這代表 Gate 0 的限制範圍比文件字面描述更窄，需寫清楚。
4. 產出：Gate 0 verdict（Pass / Fail / Partial，附證據）+ 若判定需要 bundle Git for Windows，記為 Route B 的 deployment cost。

## 5. Phase 2 — Gate 1：Model Tool Authority（需要 provider credential，若 Blocked 可跳過並標記）

1. 撰寫 `spikes/pi-compatibility/taylor-tools.ts`，用 `pi.registerTool()` 註冊 `taylor_read` / `taylor_write` / `taylor_run` 三個最小可用工具（`taylor_run` 這階段可先回傳假結果或呼叫 `child_process`，Gate 3/4 會換成真正的 Go Host 呼叫）。
2. 啟動：
   ```
   pi --mode rpc --no-builtin-tools --no-extensions -e ./taylor-tools.ts --no-session --provider <p> --model <m>
   ```
3. 依序送 prompt，要求模型：讀檔、改檔、執行命令、**明確嘗試**呼叫內建工具名稱（`bash`、`read`、`write` 等）、呼叫一個不存在的工具名稱。
4. 完整記錄每次的 RPC event/response（存成 `.jsonl` log）。
5. **PASS 條件**：模型能造成的所有副作用，唯一路徑是 `taylor_*` 工具；呼叫已停用/不存在工具名稱時，得到的是協定層錯誤而非靜默成功或被 Pi 私自轉發到內建工具。
6. 若 Phase 0 判定無 credential：本 Phase 標記 `Blocked — 缺少 LLM provider credential`，記錄需要哪一種 key 才能補測，不要用其他 Gate 的結果去推論本 Gate。

## 6. Phase 3 — Gate 2：RPC Side Channel（不需要真實模型呼叫，務必完整執行）

1. Pi 以 RPC 模式啟動（可沿用 Gate 0 baseline 啟動指令），在**尚未有任何模型回合**的狀態下，直接對 stdin 送出：
   ```json
   {"type":"bash","command":"echo gate2"}
   ```
   觀察是否得到 `bash` command 的 `response`（含 `data.output`）。
2. 分別在有 / 無 `--no-builtin-tools` 的啟動參數組合下重覆步驟 1，確認 `--no-builtin-tools` 是否影響此 host-level RPC command 的可用性。
3. 寫成結論：「Taylor 自己撰寫的 RPC client 只要不主動發送 `{"type":"bash"}`，Model Authority 是否即可視為成立」——這是人工判斷題，需要基於步驟 1/2 的實測證據論證，不是猜測。
4. 草擬 architecture invariant 的具體落地方式：例如在 Go 端 RPC client 程式碼上加一條可執行的 lint/CI 規則（如對 client 原始碼做字面 grep `"type":\\s*"bash"` 或等效 AST 檢查，禁止出現），作為**提案**寫入 REPORT.md，不需要真的接進本 repo 的 CI（這是 Spike 產出，不進 main）。

## 7. Phase 4 — Gate 3：Windows Process Authority（不需要 Pi/LLM，可獨立完整執行）

1. 在 `spikes/pi-compatibility/cmd/taylorrun/` 建立一個最小 Go CLI，直接呼叫既有 `internal/exec` 套件執行一條命令字串——用來代表「Taylor custom run tool → Go Host → PowerShell 7 → 獨立 Job Object」這條路徑上 Go Host 端的行為。**優先重用 `internal/exec` 既有 API，不要重寫 Job Object 邏輯。**
2. 覆蓋以下情境（`internal/exec/runner_windows_test.go` 已涵蓋的部分直接引用既有測試證據，不必重跑；只需新增下方兩個既有測試沒有的情境）：
   - 既有覆蓋（沿用即可）：normal exit、timeout、abort/cancel、child process、grandchild process、background process。
   - 需新增的 Spike 專屬情境：
     a. **模擬 Pi abort**：Go 端 context 被取消，但代表 Pi 的 Node 行程完全不受影響地繼續存活，驗證 Job Object 清理不依賴 Pi。
     b. **模擬 Pi crash**：直接 kill 掉代表 Pi 的父 Node 行程，驗證 Go Host 這邊獨立的 Job Object 與其管理的子程序樹完全不受影響（因為它本來就不是 Pi 行程樹的子系）。
3. **PASS 條件**：Task 終態後沒有任何應被終止的 descendant process 殘留，且整個清理路徑不需要連 Pi Runtime 一起關閉。
4. 產出：測試結果摘要 + 指向重用的既有測試檔案位置，避免重複造輪子的紀錄要寫清楚（哪些是新增、哪些是引用既有）。

## 8. Phase 5 — Gate 4：Dual Runtime Complexity（需要 provider credential；若 Gate 1 是 Blocked，本 Phase 一併標記 Blocked，不要獨立硬做）

1. 讓 Phase 2 的 `taylor-tools.ts` 的 `taylor_run` 改為真正呼叫 Phase 4 的 `taylorrun.exe`（subprocess 呼叫即可，不需要額外設計協定）。
2. 透過 Pi RPC 完成一個最小真實 coding task：read → modify → test → diff → completion（例如在 spike 資料夾內建一個 3 行的 scratch 檔案，要求模型加一行、跑一個檢查指令、回報差異）。
3. 記錄跨越邊界時觀察到的複雜度：IPC 呼叫序列化格式、error 從 Go → Node tool result → RPC event 的轉譯方式、streaming events、cancellation 傳遞、tool result transport、session lifecycle、實際除錯難度（例如卡住時要看幾層 log 才找得到根因）。
4. 產出具體量化證據（非主觀形容詞），例如：新增了幾種訊息型別的轉譯、glue code 大約幾行、是否需要手工做 retry/error-mapping。這些數字直接支撐「是否需要龐大 Go↔TypeScript bridge」的判斷。

## 9. Phase 6 — Report & Decision

於 `spikes/pi-compatibility/REPORT.md` 產出：

- 每個 Gate 的 verdict：`Pass` / `Fail` / `Blocked`（Blocked 需寫清楚缺什麼、如何補測），並附證據指向（log 檔路徑、程式碼位置）。
- 兩條 Authority Invariant（Model Authority 必須成立 / Runtime Authority 明確不要求）分別是否被證實、被反證，或因 Blocked 而無法判定。
- 對照 Issue 內「決策輸出」表格，指出目前證據落在哪一列（全部 Gate 通過 → Route B 列第一候選 / Gate 0 或 3 未過 → 回 Route A / Gate 2 顯示 Model Authority 無法成立 → 回 Route A 或重新評估容器化）。若因 Blocked 導致無法判定，明講「證據不足以下決策，需要 <什麼> 才能補齊」，不要硬選一列。
- **不要**動 `DECISIONS.md` 或任何 ADR 內容——是否寫入 ADR 屬於使用者決策，不在本 Spike 實作範圍。
- Commit 所有內容到 `agent/pi-spike-issue-24`，不 merge、不開 PR（除非另有指示）。

## 10. 完成後回報格式（給指揮者 QA 用）

工作結束（或全部卡在 Blocked）時，回報：
- 每個 Gate 的 verdict 一行摘要
- REPORT.md 路徑與 branch 名稱
- 明確列出哪些部分是 Blocked、原因、需要什麼才能補測
- 是否有任何越界行為需要指揮者注意（例如不小心動到 main、或系統層級設定被修改）

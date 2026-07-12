# Brunel Alpha 1 Specification

**版本**：v1.1
**狀態**：Review
**日期**：2026-07-13
**適用對象**：實作工程師、AI 代理（Claude Code / Codex）、規格審查者
**技術環境**：Go 1.24.x、零 CGO、靜態編譯、Windows x64、PowerShell 7（`pwsh`）
**前置文件**：`Brunel_產品提案.md`（產品層決策）、`ADR-001`（執行環境選型）
**專案根目錄**：`D:\AgentCoding\Brunel`（獨立 repository，Apache License 2.0）

---

## 0. Assumptions（假設彙整）

> 本節列出所有因資訊不足而自行假設的內容。實作前請與使用者確認。

| # | 假設內容 | 影響範圍 | 確認狀態 |
|---|---|---|---|
| A-1 | Go 版本採 1.24.x（最新穩定分支），`GOOS=windows GOARCH=amd64 CGO_ENABLED=0` | 建置腳本、CI | 待確認 |
| A-2 | 批准互動採純文字 prompt（`y` / `n` / `a`=本 session 同類全允許），不做 TUI | §7.5 ApprovalGate | 待確認 |
| A-3 | Alpha 1 的 `brunel.exe` 不做程式碼簽章，SmartScreen 警告在 README 揭露 | 發布流程 | 待確認 |
| A-4 | Alpha 1 無預設成本上限，僅記錄與顯示；benchmark 模式強制要求上限 | §7.6 Provider、§11 Benchmark | 待確認 |
| A-5 | 3 個 smoke benchmark tasks 直接沿用提案 §13 的三類任務（小型 bug 修復、小功能修改、失敗測試診斷），任務語料自行建立於 `bench/tasks/` | §11 | 待確認 |
| A-6 | 文字搜尋自行以 Go 實作（走訪 + 正規表示式），不外部依賴 ripgrep 執行檔 | §7.3 `search_text` | 待確認 |
| A-7 | `workspace_diff` 在 Git repo 中以 `git diff` 子程序取得；不引入 go-git | §7.3 `workspace_diff` | 待確認 |
| A-8 | 對話式 REPL 的輸出為串流純文字，無語法高亮與 markdown 渲染 | §5.1 | 待確認 |

---

## 1. 文件目的與背景

### 1.1 為什麼要做這個

Brunel 驗證一個可證偽的產品命題：**當模型的自主推理與工具使用能力足夠強時，coding harness 不需要用固定工作流、強制規劃、龐大系統提示與角色限制替模型做決定。**框架應退回到工具、邊界、透明度與完成證據四件事上。

這個命題目前只是信念。Alpha 1 的存在意義是把它變成**有量尺、可被推翻的實驗**——因此 benchmark smoke runner 是 Alpha 1 的發布門檻，而非後續里程碑。

### 1.2 本文件的角色

本規格書是 Brunel Alpha 1 的主契約文件，定義 agent loop、工具 schema、安全分類、session schema、provider adapter 與驗收測試。

`[FROZEN]` 標記的介面、資料結構與模組邊界不得由 AI 代理自行修改；需修改者必須提出 spec revision 並取得使用者裁決。

### 1.3 與 Taylor / Watt 的關係 [FROZEN]

Brunel 屬於 Taylor 產品線的**品牌血緣**（同 Woolf / Perkins 模式），**工程上完全獨立**：

| 項目 | 規定 |
|---|---|
| 共用程式碼 | **無**。Brunel 不 import Taylor 或 Watt 的任何套件 |
| 共用資料格式 | **無**（`CompletionReport` 例外，見 §7.8——它是 Brunel 自有格式，Taylor 未來單向讀取） |
| 對 Taylor 的依賴 | **無**。Brunel 可在完全沒有 Taylor 的環境獨立運作 |
| 對 Watt 的依賴 | **無**。Brunel **不得**偵測、呼叫或整合 Watt |

---

## 2. 目標（Goals）

- **G-1**：在乾淨 Windows x64 環境下，使用者下載單一 `brunel.exe` 即可執行，無需預裝任何 runtime。
- **G-2**：模型可在固定 workspace 內完成「理解 → 修改 → 驗證」閉環，全程使用 8 個固定內建工具。
- **G-3**：工作區內的一般操作不打斷使用者；危險操作被確定性攔截，且同類授權在單一 session 內不重複詢問。
- **G-4**：Harness 不強制模型規劃，但完成時要求可檢查的證據，而非模型的文字宣告。
- **G-5**：所有被裁剪的上下文皆保留於本機 append-only event log，可按需重載，不靜默遺失。
- **G-6**：可執行 3 個 smoke benchmark tasks，在拋棄式 workspace 副本中無人值守跑完，輸出成功率／token／耗時／工具成功率。

---

## 3. 非目標（Non-Goals）

- **NG-1**：不實作 MCP、通用 plugin system、Taylor connector 或任何外掛機制。
- **NG-2**：不實作技能（skills）系統。技能目錄結構在 §6.2 預留，但 Alpha 1 不得填入載入邏輯。
- **NG-3**：不實作 subagent 派遣。
- **NG-4**：不實作 LSP、瀏覽器、專用 web 搜尋或網頁讀取工具。
- **NG-5**：不實作完整 Git 抽象、worktree、PR 或 GitHub 整合。不做自動 commit / push / 發布 / 部署。
- **NG-6**：不實作全螢幕 TUI。
- **NG-7**：不支援 macOS、Linux、Windows PowerShell 5.1、CMD、Git Bash、WSL。
- **NG-8**：不實作 ChatGPT OAuth adapter（法遵可行性未驗證，見 `docs/open-questions.md`）。
- **NG-9**：不收集任何遙測。不提供任何預設上傳路徑。
- **NG-10**：**不實作驗證子系統。** Brunel 記錄模型執行了哪些驗證命令與結果，但不提供 pipeline 定義、驗證命令的內建知識、快取或歷史統計。（見 §8.4）
- **NG-11**：不加密 session；憑證與 session 分離，僅 API key 進 Windows Credential Manager。

---

## 4. 使用情境

### 4.1 主要使用情境

**情境 1：互動式修 bug**

```
使用者：進階個人開發者
前置狀態：位於某 Go 專案目錄，已設定 OpenRouter API key 與模型
動作：執行 `brunel`，輸入「TestParseConfig 在 Windows 上失敗，找出原因並修好」
預期結果：
  - 模型自主搜尋、讀檔、下 patch、跑 `go test`
  - 全程未被權限詢問打斷（皆為 workspace 內 AUTO 操作）
  - 完成時輸出：diff、執行過的驗證命令與退出碼、剩餘風險
  - Harness 檢查完成證據存在，否則要求模型補齊
```

**情境 2：單次任務 + 完成報告**

```
使用者：進階個人開發者
前置狀態：CI 或腳本環境，無 TTY
動作：`brunel "把 logger 的 level 從 env var 讀取" --report out.json`
預期結果：
  - 無 TTY 時不卡住等待批准
  - 若過程中需要 SESSION_GRANT 以上權限而無法取得 → 以 E_APPROVAL_REQUIRED_NO_TTY 及非零退出碼結束
  - 成功時寫出 CompletionReport JSON
```

**情境 3：唯讀分析**

```
使用者：進階個人開發者
前置狀態：不信任的第三方 repo
動作：`brunel --mode readonly "這個專案的認證流程怎麼跑的"`
預期結果：
  - 所有寫入工具與 run_powershell 一律 DENY，不詢問
  - 僅 list_files / search_text / read_file / workspace_diff 可用
```

**情境 4：命名 session 恢復**

```
使用者：進階個人開發者
前置狀態：昨天以 `--name refactor-auth` 執行過，中途離開
動作：`brunel --resume refactor-auth`
預期結果：
  - 載入結構化摘要、目標、決策、diff、驗證結果、未完成事項
  - 完整歷史不自動載入，可按需重載
  - 不恢復任何舊 shell 程序；不恢復任何危險操作授權
```

### 4.2 例外情境 / 邊界情況

| 情境 | 預期處理方式 |
|---|---|
| 檔案在 `read_file` 後被外部修改，模型才下 `apply_patch` | 回傳 `E_STALE_READ`（含 `expected_hash` / `actual_hash` / `path`）。**不得靜默重讀後覆寫。** |
| 模型給出的路徑經 junction / symlink 逃逸出 workspace root | 回傳 `E_PATH_ESCAPE`。路徑解析必須在 `syscall` 層取得最終真實路徑後比對 |
| 無 TTY 且需要批准 | 回傳 `E_APPROVAL_REQUIRED_NO_TTY`，非零退出碼，不阻塞 |
| `run_powershell` 逾時 | 以 Job Object 終止整棵程序樹，回傳 `E_TOOL_TIMEOUT` 及已捕獲的 stdout/stderr |
| Provider 連續 3 次失敗 | 回傳 `E_PROVIDER_EXHAUSTED`。若有其他相容模型，詢問使用者是否切換；**不靜默切換** |
| 選定模型未通過 tool-call probe | 回傳 `E_MODEL_NOT_TOOL_CAPABLE`，拒絕進入 agent loop。**不得以純文字模擬工具呼叫** |
| `apply_patch` 的 context 行對不上 | 回傳 `E_PATCH_CONFLICT`，附衝突區塊。不做模糊比對 |
| `create_file` 目標已存在 | 回傳 `E_FILE_EXISTS`。**不覆寫** |
| benchmark 模式中觸發 CONFIRM 以上操作 | 該案例**直接判失敗**，不等待批准，不 fallback |

---

## 5. 功能規格

### 5.1 Must Have（Alpha 1 必須完成）

- **F-1｜CLI 進入點**
  - `brunel`：於目前目錄啟動互動式 REPL（串流純文字輸出）。
  - `brunel "<任務>"`：單次任務模式。
  - `--mode workspace|readonly|benchmark`（預設 `workspace`）。
  - `--name <n>`、`--resume <name|id>`、`--report <path>`、`--model <id>`。
  - 有 TTY → 就地處理批准；無 TTY → 缺授權即以 `E_APPROVAL_REQUIRED_NO_TTY` 結束。

- **F-2｜Workspace 綁定**
  - Session 開始時固定 workspace root，session 中不可變更。
  - 所有路徑經 `filepath.EvalSymlinks` + 真實路徑比對，攔截 junction / symlink 逃逸。

- **F-3｜8 個固定內建工具**（schema 見 §7.3，`[FROZEN]`）
  `list_files` / `search_text` / `read_file` / `apply_patch` / `create_file` / `write_file` / `run_powershell` / `workspace_diff`

- **F-4｜Stale-read 防護**
  - `read_file` 回傳 `content_hash`（SHA-256，全檔）。
  - `apply_patch` / `write_file` 必須帶入 `expected_hash`，不符即 `E_STALE_READ`。

- **F-5｜受控 PowerShell 7 執行**
  - 子程序一律綁入 Windows **Job Object**（`CreateJobObject` / `AssignProcessToJobObject`）。
  - 支援逾時、取消、**整棵程序樹終止**（`JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE`）。
  - Job Object 同時設定程序數與記憶體上限（benchmark 模式為硬性；一般模式為寬鬆保護值）。

- **F-6｜安全分類與批准閘門**
  - 四級風險：`AUTO` / `SESSION_GRANT` / `CONFIRM` / `DENY`（見 §7.5）。
  - **所有工具呼叫必經單一 ApprovalGate**（架構不變式，見 §6.4）。
  - 同一 session 內已授權的同類操作不重複詢問。

- **F-7｜OpenRouter Provider Adapter**
  - 模型清單只顯示 OpenRouter metadata 標示支援結構化 tool call 的模型。
  - 首次使用某模型執行一次無副作用 tool-call probe，通過後於本機短期快取。
  - 重試策略：總共最多 3 次；間隔約 1s / 3s（加抖動）；遵守 `Retry-After` 但單次等待上限 30s；認證／額度／模型不存在／格式錯誤不重試。

- **F-8｜Agent Loop 與上下文管理**
  - 不強制規劃、不強制 plan/implement/review/verify 階段。
  - 確定性保留規則 + 模型摘要（見 §8.2）。
  - 所有裁剪內容保留於 append-only `events.jsonl`，可按需重載。

- **F-9｜Session 保存與恢復**
  - 預設暫存；可命名（啟動時或進行中）。未命名 session 正常退出後刪除。
  - 異常中止的未命名 session 短期保留（預設 72 小時）供救援，逾期清理。
  - Session 名稱不唯一，內部使用不可變 ID（ULID）。
  - 落盤前盡力遮罩已知 secret 模式；**不宣稱能偵測所有敏感資訊**（README 明確揭露）。

- **F-10｜`AGENTS.md` 讀取**
  - 啟動讀 workspace root；操作子目錄檔案前按需讀取更接近的 `AGENTS.md`，較近者優先。
  - **`AGENTS.md` 只能約束工作方式，不能授予權限或關閉安全紅線**（`[FROZEN]` 不變式）。

- **F-11｜設定分層**
  - 優先序：CLI flags > 專案 `<workspace>\.brunel\config.json` > 全域 `%USERPROFILE%\.brunel\config.json` > Brunel 預設。
  - 憑證只存 Windows Credential Manager（DPAPI），**專案設定不得存憑證，不得降低硬性安全政策**。

- **F-12｜完成判定**
  - 模型先宣告本次驗收條件，Harness 檢查對應證據是否存在（見 §8.3）。
  - 缺證據 → 要求模型補齊或明確揭露未執行原因，不得直接宣告完成。

- **F-13｜CompletionReport 輸出**（schema `[FROZEN]`，見 §7.8）
  - `--report <path>` 寫出結構化 JSON：需求對應、diff、執行過的驗證命令與退出碼、未執行項與原因、剩餘風險、成本。

- **F-14｜Smoke Benchmark Runner**（3 個任務）
  - 每案例使用專用、可拋棄的 workspace 副本。
  - 硬性時間 / token / 費用 / 程序數上限（透過 Job Object 與 provider 計數）。
  - CONFIRM 以上操作直接判該案例失敗，不等待批准，不 fallback 模型。
  - 輸出：任務成功率、測試通過率、首次完成率、返工次數、token、費用、耗時、工具成功率、安全攔截次數。

### 5.2 Should Have（不擋 Alpha 1 驗收）

- **F-15｜Context Ledger**：`brunel context` 顯示目前 context 的 token 占用按來源分佈（system prompt / AGENTS.md / 使用者指令 / 工具輸出 / diff）。這是「薄」這個主張的唯一量尺，強烈建議做進 Alpha 1。
- **F-16｜Prompt 透明化**：`brunel prompt --show` 印出實際送出的 agent prompt 與其 token 占用，並允許使用者自訂一般 agent prompt（安全核心與工具真實契約固定不可改）。
- **F-17｜非 Git 專案的 session 快照 diff**（Git 專案走 `git diff`）。

### 5.3 Could Have（Alpha 1 不實作）

- **F-18**：技能按需載入（Alpha 2）。
- **F-19**：Subagent 派遣（Alpha 3）。
- **F-20**：完整 benchmark runner（10–15 任務、比較組報表）（Alpha 4）。

### 5.4 Won't Have（明確排除）

- **F-21**：MCP / plugin system / Taylor connector — 對 Brunel 的實際事件與控制需求尚未有經驗證據，提前設計必然錯。
- **F-22**：LSP — 僅在 benchmark 證明文字搜尋與編譯器診斷不足時才重新評估。
- **F-23**：**任何形式的驗證子系統**（pipeline 定義、驗證命令內建知識、結果快取、歷史統計、token 節省估算）。見 §8.4。

---

## 6. 系統架構

### 6.1 模組總覽

```
cmd/brunel                      CLI 進入點（cobra）
  │
  ├── internal/config           分層設定 + Credential Manager
  ├── internal/session          Session 生命週期、events.jsonl、summary、resume
  │
  └── internal/agent            Agent Loop（核心）
        ├── internal/context    Context 組裝、確定性保留規則、裁剪、摘要觸發
        ├── internal/provider   Provider Port + OpenRouter Adapter、retry、probe
        ├── internal/completion 完成判定、CompletionReport 產出
        │
        └── internal/tools      Tool Registry（8 個固定工具）
              │
              └── internal/safety      ApprovalGate（唯一風險裁決點）
                    ├── internal/workspace  路徑守衛、hash、快照、diff
                    └── internal/exec       PowerShell Runner + Job Object

internal/bench                  Smoke Benchmark Runner（獨立於 agent，驅動 agent）
```

### 6.2 模組職責

| 模組 | 職責 | **不負責** |
|---|---|---|
| `agent` | Agent loop、與模型往返、工具派發 | 不決定任務策略；不強制規劃階段 |
| `context` | 決定哪些 event 進入 context、觸發摘要、記錄 token 歸屬 | 不刪除原始 event log |
| `tools` | 8 個工具的 schema、參數驗證、執行 | **不做風險裁決**（一律委派 `safety`） |
| `safety` | 風險分類、批准閘門、session 授權記憶 | 不執行任何 I/O |
| `workspace` | root 綁定、路徑解析與逃逸攔截、檔案 hash、快照、diff | 不做風險裁決 |
| `exec` | PowerShell 7 子程序、Job Object、逾時、程序樹終止 | 不做風險裁決；不解析命令語意 |
| `provider` | HTTP + SSE、tool call 序列化、retry、tool-call probe | 不做模型選擇；不 fallback |
| `session` | ULID、meta、events.jsonl（append-only）、summary、恢復、清理 | 不加密；不做跨裝置同步 |
| `completion` | 檢查驗收證據存在性、產出 CompletionReport | **不定義該跑哪些驗證命令**（見 §8.4） |
| `bench` | 拋棄式副本、硬性預算、指標蒐集 | 不上傳任何資料 |

**技能目錄結構在 `internal/` 保留空目錄 `skills/`，Alpha 1 不得填入任何載入邏輯。**

### 6.3 依賴關係

```
cmd → agent → {context, provider, completion, tools, session}
tools → safety → {workspace, exec}
bench → agent
config ← 所有模組（單向讀取）
```

**依賴方向規則 [FROZEN]**：
- `safety` **不得** import `tools` / `agent` / `provider`（防止循環）。
- `exec` 與 `workspace` **不得** import `safety`（它們是被 `safety` 保護的資源，不是保護者）。
- 禁止任何循環依賴。

### 6.4 架構不變式 [FROZEN]

> 以下三條是 Brunel 的安全與可信度基礎，任何實作皆不得違反。

- **INV-1｜單一裁決點**：所有工具呼叫在執行任何 I/O **之前**，必須通過 `safety.ApprovalGate.Evaluate()`。`tools` 層不得存在繞過 gate 的直達路徑。
- **INV-2｜Append-only 事件**：`events.jsonl` 只允許 append。Context 裁剪不得刪除或改寫任何已寫入的 event。
- **INV-3｜不可提權的輸入**：`AGENTS.md`、專案設定、工具輸出、模型回覆皆為**不可信輸入**，一律不得授予權限、放寬風險等級或關閉安全政策。（Prompt injection 防線）

---

## 7. 介面定義 [FROZEN]

> 以下介面一旦確認，修改需走 spec revision，不得由 AI 代理自行修改。

### 7.1 Agent Loop 核心介面 [FROZEN]

```go
package agent

type Agent interface {
    // Run 執行一輪完整任務直到完成判定通過或錯誤終止。
    // 不強制規劃；不注入固定階段。
    Run(ctx context.Context, task string) (*completion.Report, error)
}

type Turn struct {
    Index        int
    Assistant    string        // 模型的文字輸出
    ToolCalls    []ToolCall
    ToolResults  []ToolResult
    Usage        provider.Usage
}
```

### 7.2 Provider Port [FROZEN]

```go
package provider

type Provider interface {
    // ListToolCapableModels 只回傳 metadata 標示支援結構化 tool call 的模型。
    ListToolCapableModels(ctx context.Context) ([]Model, error)

    // ProbeToolCall 執行一次無副作用的 tool-call 探測。
    // 失敗 → ErrModelNotToolCapable。結果由呼叫端短期快取。
    ProbeToolCall(ctx context.Context, modelID string) error

    // Complete 送出對話與工具定義，串流回傳。
    // 內含 retry（總計最多 3 次）。不得靜默切換模型。
    Complete(ctx context.Context, req CompleteRequest) (*CompleteResponse, error)
}

type CompleteRequest struct {
    ModelID  string          // 必填
    Messages []Message       // 必填
    Tools    []ToolSchema    // 必填：8 個固定工具
    Stream   bool
}

type CompleteResponse struct {
    Text      string
    ToolCalls []ToolCall
    Usage     Usage          // 以 provider 回報為準
    Attempts  int            // 實際嘗試次數（1–3）
}

type Usage struct {
    PromptTokens     int
    CompletionTokens int
    CostUSD          *float64 // provider 未回報時為 nil，不自行估算
}
```

**重試規則 [FROZEN]**：總計最多 3 次（含首次）。第 1 次失敗後等待 ~1s，第 2 次後 ~3s，皆加隨機抖動。若回應含 `Retry-After` 則優先遵守，但單次等待硬性上限 30 秒。`401` / `402` / `404` / 明確格式錯誤**不重試**。3 次皆失敗 → `E_PROVIDER_EXHAUSTED`；有其他相容模型時**詢問**使用者是否切換。**benchmark 模式不 fallback。**

### 7.3 工具 Schema [FROZEN]

> 8 個固定工具。Alpha 1 不得新增、刪除或改名。

```go
package tools

type Tool interface {
    Name() string
    Schema() ToolSchema           // 送給模型的 JSON Schema
    Risk(args json.RawMessage) safety.Risk  // 靜態風險預判
    Execute(ctx context.Context, args json.RawMessage) (Result, error)
}
```

| # | 名稱 | 參數 | 回傳 | 預設風險 |
|---|---|---|---|---|
| 1 | `list_files` | `path` (相對), `glob?`, `max_depth?` | 檔案清單 + 大小 | AUTO |
| 2 | `search_text` | `pattern` (regex), `path?`, `glob?`, `max_results?` | 命中行（含檔名、行號） | AUTO |
| 3 | `read_file` | `path`, `start_line?`, `end_line?` | 帶行號內容 + `content_hash` (SHA-256 全檔) | AUTO |
| 4 | `apply_patch` | `path`, `expected_hash`, `hunks[]` | 新 `content_hash` | AUTO（workspace 內） |
| 5 | `create_file` | `path`, `content` | `content_hash` | AUTO（**已存在則 `E_FILE_EXISTS`，絕不覆寫**） |
| 6 | `write_file` | `path`, `expected_hash`, `content` | 新 `content_hash` | AUTO；**檔案 > 500 行或單輪累計覆寫 > 3 檔 → CONFIRM** |
| 7 | `run_powershell` | `command`, `timeout_sec?`, `cwd?` | `stdout`, `stderr`, `exit_code`, `truncated` | **動態分類**（見 §7.5） |
| 8 | `workspace_diff` | `path?` | unified diff | AUTO |

**工具層規則 [FROZEN]**：
- 結構化檔案工具（3–6）**優先於** `run_powershell` 寫檔。Agent prompt 明確聲明此偏好。
- `read_file` 未先執行過的路徑，`apply_patch` / `write_file` 必須拒絕（`E_STALE_READ`，`expected_hash` 為空亦視為不符）。
- 所有工具的 `path` 一律先經 `workspace.Resolve()`，逃逸即 `E_PATH_ESCAPE`。

### 7.4 錯誤碼定義 [FROZEN]

| 錯誤碼 | 觸發條件 | 上層處理建議 |
|---|---|---|
| `E_STALE_READ` | `expected_hash` ≠ 實際 hash，或未曾讀取該檔 | **回報模型並附 `expected_hash` / `actual_hash` / `path`。禁止靜默重讀後覆寫。** 模型應重新 `read_file` 後重下 patch |
| `E_PATH_ESCAPE` | 解析後的真實路徑落在 workspace root 之外 | 回報模型，不重試 |
| `E_FILE_EXISTS` | `create_file` 目標已存在 | 回報模型，改用 `apply_patch` 或 `write_file` |
| `E_PATCH_CONFLICT` | `apply_patch` 的 context 行對不上 | 附衝突區塊回報模型。**不做模糊比對** |
| `E_APPROVAL_DENIED` | 使用者拒絕批准 | 回報模型，模型應改變作法或詢問使用者 |
| `E_APPROVAL_REQUIRED_NO_TTY` | 無 TTY 且需 SESSION_GRANT 以上 | **立即以非零退出碼結束整個 run**，不阻塞 |
| `E_TOOL_TIMEOUT` | `run_powershell` 逾時 | Job Object 終止整棵程序樹，回傳已捕獲輸出 |
| `E_PROVIDER_AUTH` | 401 / 402 | 不重試。提示使用者檢查 Credential Manager |
| `E_PROVIDER_EXHAUSTED` | 3 次嘗試皆失敗 | 清楚報錯。有相容模型時**詢問**切換；benchmark 模式直接失敗 |
| `E_MODEL_NOT_TOOL_CAPABLE` | tool-call probe 未通過 | 拒絕進入 agent loop。**不得以純文字模擬工具呼叫** |
| `E_WORKSPACE_UNBOUND` | workspace root 未綁定或已失效 | 終止 session |
| `E_BUDGET_EXCEEDED` | benchmark 硬性上限（時間／token／費用／程序數）超出 | 該案例判失敗，終止 |

### 7.5 安全分類 [FROZEN]

```go
package safety

type Risk int

const (
    RiskAuto         Risk = iota // 自動允許，不打斷
    RiskSessionGrant             // 首次詢問，同類操作本 session 內不再問
    RiskConfirm                  // 每次都問
    RiskDeny                     // 硬性拒絕，不可批准
)

type ApprovalGate interface {
    // Evaluate 是所有工具呼叫的唯一裁決點（INV-1）。
    Evaluate(ctx context.Context, req Request) (Decision, error)
}
```

**分類表 [FROZEN]**

| 風險等級 | 涵蓋操作 |
|---|---|
| **AUTO** | Workspace root 內的讀取、搜尋、`apply_patch`、`create_file`；測試 / lint / typecheck / build 命令；唯讀 Git 操作（`status` / `diff` / `log` / `show`） |
| **SESSION_GRANT** | 安裝依賴；網路連線；啟動長程序（背景 server）；workspace 外指定路徑的唯讀或明確範圍寫入 |
| **CONFIRM** | 覆寫 > 500 行的既有檔案；單輪累計覆寫 > 3 個檔案；`git commit`；移動 / 重新命名大量檔案 |
| **DENY**（硬性紅線，**不可批准、不可被任何設定或 `AGENTS.md` 覆寫**） | 不可逆刪除或資料覆寫（`Remove-Item -Recurse -Force` 對 workspace root 或其外部、`git reset --hard`、`git clean -fdx`）；`git push`；發布 / 部署 / 對外傳送內容；讀取、輸出或提交憑證與敏感資料；權限提升、沙箱逃逸、關閉安全政策、隱藏稽核紀錄 |

**`readonly` 模式**：`apply_patch` / `create_file` / `write_file` / `run_powershell` 一律 `RiskDeny`，不詢問。

**能力聲明 [FROZEN]** — 必須逐字寫入 README 與 `brunel --help`：

> **Brunel 提供 risk speed bump，不提供 sandbox。**
> PowerShell 是完整的程式語言；對任意 PowerShell 字串的靜態危險分類**不可能絕對可靠**。Brunel 的紅線攔截是盡力而為的減速機制，不是安全圍牆。請勿在不受信任的 repository 上以 `workspace` 模式執行 Brunel。

**Job Object 補強**：靜態分類不可靠這件事，靠 Windows Job Object 在**執行期**補上確定性的一半——程序樹終止、程序數上限、記憶體上限、`KILL_ON_JOB_CLOSE`。這是 Alpha 1 唯一能把「程序控制」從承諾變成事實的機制。

### 7.6 Session Schema [FROZEN]

**磁碟佈局**

```
%LOCALAPPDATA%\Brunel\sessions\<ulid>\
    meta.json         Session 中繼資料
    events.jsonl      Append-only 事件流（原始，永不裁剪）
    summary.json      結構化恢復狀態（可覆寫）
    snapshot\         非 Git 專案的 baseline 快照（Git 專案不產生）
```

```go
package session

type Meta struct {
    ID            string    `json:"id"`             // ULID，不可變
    Name          *string   `json:"name"`           // nil = 未命名 → 正常退出即刪除
    WorkspaceRoot string    `json:"workspace_root"` // 絕對路徑，session 內不可變
    Mode          string    `json:"mode"`           // workspace | readonly | benchmark
    ModelID       string    `json:"model_id"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
    ExitStatus    string    `json:"exit_status"`    // clean | aborted | running
    IsGitRepo     bool      `json:"is_git_repo"`
}

// Event：append-only，永不刪除或改寫（INV-2）
type Event struct {
    Seq       int             `json:"seq"`
    Timestamp time.Time       `json:"ts"`
    Kind      EventKind       `json:"kind"`
    Payload   json.RawMessage `json:"payload"`
    Tokens    int             `json:"tokens"`     // 供 Context Ledger 歸屬
    Source    string          `json:"source"`     // system | agents_md | user | tool_output | diff
}

type EventKind string
const (
    EvUserInstruction EventKind = "user_instruction"
    EvAssistantText   EventKind = "assistant_text"
    EvToolCall        EventKind = "tool_call"
    EvToolResult      EventKind = "tool_result"
    EvApproval        EventKind = "approval"
    EvSummary         EventKind = "summary"
    EvError           EventKind = "error"
)

// Summary：恢復時優先載入，完整歷史按需重載
type Summary struct {
    Goal            string            `json:"goal"`
    Decisions       []string          `json:"decisions"`        // 使用者做過的決策
    ModifiedFiles   []string          `json:"modified_files"`
    LatestDiff      string            `json:"latest_diff"`
    Verifications   []Verification    `json:"verifications"`
    OpenErrors      []string          `json:"open_errors"`
    Pending         []string          `json:"pending"`          // 未完成事項
    LastEventSeq    int               `json:"last_event_seq"`
}
```

**Session 規則 [FROZEN]**
- 未命名 session：`ExitStatus == "clean"` → 立即刪除整個目錄。
- 未命名 session：`ExitStatus == "aborted"` → 保留 72 小時供救援，逾期由下次啟動時清理。
- 恢復時**不重建任何 shell 程序**；**不恢復任何危險操作授權**（SESSION_GRANT 記憶不落盤）。
- 落盤前對已知 secret 模式（API key 前綴、`Authorization:` header、`.env` 內容）盡力遮罩。**README 必須揭露此為 best-effort。**
- Session **不加密**。API key / token 只存 Windows Credential Manager，永不寫入 session 目錄。

### 7.7 Context 保留規則 [FROZEN]

**永遠保留（不得裁剪）**
- 使用者指令
- 安全政策與能力聲明
- 當前目標
- 使用者做過的決策
- 已修改檔案清單
- 最新 diff
- 驗證結果
- 未解錯誤
- 未完成事項

**可自動裁剪（原始 event 仍留在 `events.jsonl`，可按需重載）**
- 重複內容
- 過時的搜尋結果
- 已被後續工具輸出取代的舊輸出
- 冗長的成功 log

**摘要觸發**：context token 達模型上限的 70% 時，由模型對「可裁剪區」產生摘要，摘要作為 `EvSummary` event 寫入。**摘要不得取代永遠保留區**。

### 7.8 CompletionReport Schema [FROZEN]

> 這是 Brunel 自有的完成證據格式，也是未來 Taylor 單向讀取的唯一介面。此 schema 在 Alpha 1 即凍結。

```go
package completion

type Report struct {
    SchemaVersion string        `json:"schema_version"` // "1.0"
    SessionID     string        `json:"session_id"`
    Task          string        `json:"task"`
    Status        string        `json:"status"`         // completed | incomplete | failed

    // 模型宣告的驗收條件，逐項對應
    AcceptanceCriteria []Criterion `json:"acceptance_criteria"`

    ModifiedFiles []string      `json:"modified_files"`
    Diff          string        `json:"diff"`

    // Brunel 只記錄「模型實際跑了什麼」，不定義該跑什麼（NG-10）
    Verifications []Verification `json:"verifications"`
    SkippedVerifications []Skipped `json:"skipped_verifications"`

    ToolFailures  []ToolFailure `json:"tool_failures"`   // 未回報的工具失敗必須為空
    PendingApprovals []string   `json:"pending_approvals"` // 必須為空才可 completed
    RemainingRisks []string     `json:"remaining_risks"`

    Cost          CostSummary   `json:"cost"`
}

type Criterion struct {
    Statement string `json:"statement"` // 模型宣告的驗收條件
    Met       bool   `json:"met"`
    Evidence  string `json:"evidence"`  // 指向 diff / verification / tool result
}

type Verification struct {
    Command  string `json:"command"`
    ExitCode int    `json:"exit_code"`
    Summary  string `json:"summary"`   // 截斷後的輸出摘要
}

type Skipped struct {
    Command string `json:"command"`
    Reason  string `json:"reason"`     // 必填：未執行的原因必須揭露
}

type CostSummary struct {
    PromptTokens     int      `json:"prompt_tokens"`
    CompletionTokens int      `json:"completion_tokens"`
    CostUSD          *float64 `json:"cost_usd"` // provider 未回報時 nil
    DurationSec      float64  `json:"duration_sec"`
    Turns            int      `json:"turns"`
}
```

---

## 8. 運作原理

### 8.1 Agent Loop 主流程

```
Step 1  綁定 workspace root（EvalSymlinks → 真實路徑）
Step 2  載入設定（CLI > 專案 > 全域 > 預設），從 Credential Manager 取 API key
Step 3  解析模型：ListToolCapableModels → ProbeToolCall（快取命中則跳過）
        未通過 → E_MODEL_NOT_TOOL_CAPABLE，終止
Step 4  讀取 workspace root 的 AGENTS.md（作為不可提權的工作方式約束）
Step 5  組裝 context（§7.7 保留規則）→ Provider.Complete
Step 6  模型回傳 text + tool_calls
          for each tool_call:
            a. tools 層驗證參數
            b. workspace.Resolve(path) → 逃逸即 E_PATH_ESCAPE
            c. safety.ApprovalGate.Evaluate()  ← INV-1 唯一裁決點
                 AUTO          → 直接執行
                 SESSION_GRANT → 已授權則放行；否則詢問（無 TTY → E_APPROVAL_REQUIRED_NO_TTY）
                 CONFIRM       → 每次詢問
                 DENY          → 拒絕，回報模型
            d. 執行 → 寫入 events.jsonl（append-only, INV-2）
Step 7  context token > 70% → 觸發摘要（僅對可裁剪區）
Step 8  模型宣告完成 → completion 檢查證據（§8.3）
          證據不足 → 把缺口回報模型，回到 Step 5
          證據齊備 → 產出 Report，結束
```

**沒有 plan 階段。沒有 review 階段。沒有角色提示。** 模型可直接執行安全且可逆的必要工作。

### 8.2 關鍵設計決策

| 決策 | 選擇 | 理由 |
|---|---|---|
| 執行環境 | Go 1.24 / 零 CGO / 靜態編譯 | Job Object 為硬需求；Node 無法在維持單檔 exe 的前提下滿足（ADR-001） |
| 程序控制 | Windows Job Object | 唯一能把程序樹終止與資源上限從承諾變成事實的機制；同時是 benchmark 硬性預算的實作基礎 |
| 安全模型 | 4 級風險 + 單一 gate | 單一裁決點（INV-1）使安全性可被審查與測試；分散判斷必然出漏洞 |
| 安全能力聲明 | speed bump，非 sandbox | PowerShell 是完整語言，靜態分類不可能可靠。誠實揭露優於虛假保證 |
| 完成判定 | 證據存在性檢查 | 不信任模型的文字宣告；但也不硬編該跑哪些命令（那會變成另一個產品，見 §8.4） |
| Benchmark | 進 Alpha 1 發布門檻 | 「薄 harness 更好」是本產品的核心命題。沒有量尺就無法證偽，前三個里程碑都會在造尺子之前先造房子 |
| Provider | 只綁 OpenRouter | 多供應商抽象會退化成最低共同能力。保留小型 adapter 邊界即可 |
| Tokenize | 吃 provider 回報的 usage | 使用者可自選任意模型，本地精確 tokenize 本來就不可能。不自行估算成本 |

### 8.3 完成判定狀態機

```
[Working] --模型宣告完成--> [Claiming]
[Claiming] --證據齊備--> [Completed]
[Claiming] --證據不足--> [Working]（把缺口回報模型）
[Working] --致命錯誤--> [Failed]
```

**證據齊備的定義（全部滿足）**
1. 每一條模型宣告的 `AcceptanceCriteria` 都有對應 `Evidence`（指向 diff / verification / tool result）。
2. `ToolFailures` 中沒有未回報的失敗。
3. 所有 `ModifiedFiles` 皆可產生 diff。
4. `Verifications` 非空 **或** `SkippedVerifications` 中每一項都有非空的 `Reason`。
5. `PendingApprovals` 為空。
6. 最終回覆包含成果、驗證與剩餘風險。

### 8.4 「不做驗證子系統」的界線 [FROZEN]

> 這是最容易被 AI 代理擅自擴張的地方。以下對照表為硬性契約。

| Brunel **做** | Brunel **不做** |
|---|---|
| 記錄模型實際執行了哪些驗證命令 | 內建「Go 專案該跑 `go test`」這類知識 |
| 記錄退出碼與輸出摘要 | pipeline / task 定義檔（如 `*.toml`） |
| 記錄哪些驗證該做而沒做，以及原因 | 驗證結果快取 |
| 檢查「證據存在」 | 跨 session 的歷史統計 |
| — | token 節省估算 |
| — | 偵測或呼叫任何外部 CI 工具 |

**Brunel 不硬編任何專案類型應執行哪些驗證命令。** 該跑什麼由模型決定，Brunel 只要求可檢查的證據。

---

## 9. 替代方案

### 方案 A：單一 ApprovalGate + 硬編 4 級風險（推薦）

**做法**：風險分類邏輯以 Go 程式碼實作於 `internal/safety`，所有工具呼叫必經單一 `Evaluate()`。分類表寫死在程式碼中，使用者只能在 `AUTO` ↔ `SESSION_GRANT` 之間微調，`DENY` 紅線完全不可設定。

**優點**：
- 單一裁決點可被完整單元測試覆蓋，安全性可審查。
- 紅線無法被 `AGENTS.md`、專案設定或模型輸出繞過（滿足 INV-3）。
- 實作簡單，Alpha 1 可如期完成。

**缺點 / 限制**：
- 新增風險規則需改程式碼並重新發布。
- 使用者無法為特殊 workflow 客製分類。

### 方案 B：宣告式 Policy Engine（如 Rego / CEL）

**做法**：把風險分類外部化成宣告式規則檔，`safety` 只做規則求值。

**優點**：
- 規則可熱更新，使用者可客製。
- 規則與程式碼分離，易於審計。

**缺點 / 限制**：
- **與 INV-3 直接衝突**：規則檔一旦可被專案層覆寫，`AGENTS.md` 或惡意 repo 就可能提權。要防這件事又得再加一層「哪些規則可被覆寫」的元規則，複雜度爆炸。
- 引入額外依賴（CEL / OPA），違背「依賴數量是負債」原則。
- Alpha 1 沒有足夠的真實規則樣本來設計正確的規則語言——提前設計必然錯。

### 推薦方向

選擇**方案 A**。理由：Alpha 1 的目的是驗證「薄 harness」命題，不是驗證「可設定的安全引擎」。方案 B 的彈性在目前沒有任何需求證據支撐，而它引入的提權風險（INV-3）是實打實的。等 benchmark 累積出真實的規則需求後，再評估是否外部化。

---

## 10. 編程範例

### 10.1 工具執行的正確路徑

```go
// ✅ 正確：所有 I/O 前必經 ApprovalGate（INV-1）
func (r *Registry) Dispatch(ctx context.Context, call ToolCall) (Result, error) {
    tool, ok := r.tools[call.Name]
    if !ok {
        return Result{}, ErrUnknownTool
    }

    // 1. 路徑解析與逃逸攔截
    if p, ok := call.PathArg(); ok {
        if _, err := r.ws.Resolve(p); err != nil {
            return Result{}, err // E_PATH_ESCAPE
        }
    }

    // 2. 唯一裁決點
    dec, err := r.gate.Evaluate(ctx, safety.Request{
        Tool: call.Name,
        Args: call.Args,
        Risk: tool.Risk(call.Args),
    })
    if err != nil || !dec.Allowed {
        return Result{}, err // E_APPROVAL_DENIED / E_APPROVAL_REQUIRED_NO_TTY
    }

    // 3. 才執行
    return tool.Execute(ctx, call.Args)
}
```

```go
// ❌ 錯誤：工具自行判斷風險後直接執行，繞過 gate（違反 INV-1）
func (t *WriteFileTool) Execute(ctx context.Context, args json.RawMessage) (Result, error) {
    if t.isSafe(args) {          // ← 分散的風險判斷
        return t.doWrite(args)   // ← 繞過 ApprovalGate
    }
    ...
}
```

### 10.2 Stale-read 的正確處理

```go
// ✅ 正確：hash 不符即報錯，把決定權交回模型
func (t *ApplyPatchTool) Execute(ctx context.Context, args json.RawMessage) (Result, error) {
    var a ApplyPatchArgs
    json.Unmarshal(args, &a)

    actual, err := t.ws.HashFile(a.Path)
    if err != nil {
        return Result{}, err
    }
    if actual != a.ExpectedHash {
        return Result{}, &StaleReadError{
            Path:         a.Path,
            ExpectedHash: a.ExpectedHash,
            ActualHash:   actual,
        } // E_STALE_READ
    }
    return t.applyHunks(a)
}
```

```go
// ❌ 錯誤：靜默重讀後覆寫，吃掉使用者的變更
if actual != a.ExpectedHash {
    content, _ := os.ReadFile(a.Path)  // ← 靜默重讀
    return t.applyHunksTo(content, a)  // ← 覆蓋使用者變更
}
```

### 10.3 Job Object 綁定

```go
// ✅ 正確：子程序在啟動時即綁入 Job Object，KILL_ON_JOB_CLOSE 確保程序樹不外洩
func (r *PSRunner) Run(ctx context.Context, cmd string, timeout time.Duration) (Output, error) {
    job, err := windows.CreateJobObject(nil, nil)
    if err != nil {
        return Output{}, err
    }
    defer windows.CloseHandle(job) // KILL_ON_JOB_CLOSE → 整棵程序樹隨之終止

    limits := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{
        BasicLimitInformation: windows.JOBOBJECT_BASIC_LIMIT_INFORMATION{
            LimitFlags: windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE |
                        windows.JOB_OBJECT_LIMIT_ACTIVE_PROCESS |
                        windows.JOB_OBJECT_LIMIT_PROCESS_MEMORY,
            ActiveProcessLimit: r.maxProcesses,
        },
        ProcessMemoryLimit: r.maxMemoryBytes,
    }
    windows.SetInformationJobObject(job, windows.JobObjectExtendedLimitInformation,
        uintptr(unsafe.Pointer(&limits)), uint32(unsafe.Sizeof(limits)))

    c := exec.CommandContext(ctx, "pwsh", "-NoProfile", "-NonInteractive", "-Command", cmd)
    c.SysProcAttr = &syscall.SysProcAttr{CreationFlags: windows.CREATE_SUSPENDED}
    if err := c.Start(); err != nil {
        return Output{}, err
    }
    // 先綁 Job，再 Resume（避免子程序在綁定前 fork 出逃逸程序）
    windows.AssignProcessToJobObject(job, processHandle(c))
    resumeThread(c)
    ...
}
```

> **順序至關重要**：必須 `CREATE_SUSPENDED` → `AssignProcessToJobObject` → `ResumeThread`。先啟動再綁定會留下逃逸視窗。

---

## 11. 測試策略

### 11.1 測試層次

| 層次 | 範圍 | 工具 | 覆蓋目標 |
|---|---|---|---|
| Unit | 單一函式 | `testing` + `testify` | 風險分類表、路徑解析、hash 比對、patch 套用、retry 退避計算 |
| Integration | 跨模組 | `testing` | ApprovalGate ↔ tools 契約、Provider retry 行為（mock HTTP）、Session 存讀與恢復 |
| E2E | 完整 CLI | `testing` + 暫存 workspace | 4 個主要使用情境（§4.1） |
| Smoke Benchmark | 真實模型 | `internal/bench` | 3 個真實任務，拋棄式副本，硬性預算 |

### 11.2 必須通過的關鍵測試案例

```go
// 安全紅線
func TestGate_DenyIsNotOverridableByAgentsMD(t *testing.T)   // INV-3
func TestGate_DenyIsNotOverridableByProjectConfig(t *testing.T)
func TestGate_SessionGrantNotRepeated(t *testing.T)
func TestGate_NoTTY_ReturnsErrAndExits(t *testing.T)
func TestRegistry_AllToolsPassThroughGate(t *testing.T)      // INV-1，反射檢查無旁路

// 資料保護
func TestApplyPatch_StaleRead_DoesNotOverwrite(t *testing.T) // 核心防護
func TestWriteFile_StaleRead_DoesNotOverwrite(t *testing.T)
func TestCreateFile_ExistingFile_DoesNotOverwrite(t *testing.T)
func TestWorkspace_JunctionEscape_Blocked(t *testing.T)
func TestWorkspace_SymlinkEscape_Blocked(t *testing.T)
func TestWorkspace_AbsolutePathEscape_Blocked(t *testing.T)

// 程序控制
func TestPSRunner_Timeout_KillsProcessTree(t *testing.T)     // 起孫程序，驗證全滅
func TestPSRunner_JobBoundBeforeResume(t *testing.T)         // 逃逸視窗

// Provider
func TestProvider_MaxThreeAttempts(t *testing.T)
func TestProvider_RetryAfter_CappedAt30s(t *testing.T)
func TestProvider_AuthError_NoRetry(t *testing.T)
func TestProvider_NeverSilentlySwitchesModel(t *testing.T)
func TestProvider_NonToolCapableModel_Rejected(t *testing.T)

// Session
func TestSession_UnnamedCleanExit_Deleted(t *testing.T)
func TestSession_UnnamedAbort_RetainedThenCleaned(t *testing.T)
func TestSession_Resume_DoesNotRestoreGrants(t *testing.T)   // 授權不落盤
func TestEventLog_AppendOnly(t *testing.T)                   // INV-2

// 完成判定
func TestCompletion_ClaimWithoutEvidence_Rejected(t *testing.T)
func TestCompletion_SkippedVerification_RequiresReason(t *testing.T)
```

### 11.3 不測試的範圍

- OpenRouter 的實際 API 呼叫 → Unit / Integration 一律 mock；只有 smoke benchmark 打真實 API。
- 模型的推理品質 → 由 benchmark 指標間接衡量，不寫斷言。
- 完整 PowerShell 語言的危險分類正確性 → **明確不宣稱，不測試**（見 §7.5 能力聲明）。

---

## 12. 風險與限制

| 風險 | 嚴重度 | 緩解方式 |
|---|---|---|
| PowerShell 靜態危險分類不可能絕對可靠 | **高** | 誠實聲明為 speed bump 而非 sandbox；以 Job Object 在執行期補上確定性控制；README 警告勿在不受信任 repo 上用 `workspace` 模式 |
| Prompt injection 經 repo / `AGENTS.md` / 命令輸出進入模型 | **高** | INV-3：所有這些來源皆為不可信輸入，一律不得提權。單一 gate 使此規則可被測試 |
| OpenRouter metadata 不保證路由後端能正確完成工具迴圈 | 中 | 首次使用執行無副作用 tool-call probe，通過才進 loop |
| 模型摘要遺漏細節 | 中 | 原始 event log append-only、永不裁剪、可按需重載（INV-2） |
| 未加密 session 含程式碼與操作紀錄 | 中 | README 明確揭露；提供 `brunel session rm`；憑證與 session 分離 |
| Windows 單檔打包、程序樹終止、Credential Manager、路徑逃逸 | 中 | 全部列入 §11.2 必過測試；發布前在乾淨 VM 驗證 |
| **「薄」不代表更好** | **高** | Smoke benchmark 進 Alpha 1 發布門檻。若指標顯示品質下降，產品命題本身必須被修正而非護航 |
| Job Object 在某些 Windows 版本（巢狀 Job）行為差異 | 低 | Windows 8+ 支援巢狀 Job；最低支援版本鎖 Windows 10 21H2 |

---

## 13. 驗收標準（Alpha 1 發布門檻）

| # | 驗收項目 | 測量方式 | 通過標準 |
|---|---|---|---|
| AC-1 | 乾淨環境可執行 | 全新 Windows 11 VM，無 Node / Go，下載 `brunel.exe` 執行 | 成功啟動，無 runtime 缺失錯誤 |
| AC-2 | 憑證與模型選擇 | 設定 OpenRouter key，列出模型並選一個 | 只顯示 tool-capable 模型；probe 通過後可進 loop |
| AC-3 | `AGENTS.md` 讀取 | root 與子目錄各放一份，操作子目錄檔案 | 較近者優先；不因 `AGENTS.md` 內容改變任何權限 |
| AC-4 | 8 個工具閉環 | 在真實 Go 專案跑搜尋 → 讀檔 → patch → `go test` → diff | 全部成功，diff 正確 |
| AC-5 | Stale-read 防護 | `read_file` 後由外部改檔，再 `apply_patch` | 回 `E_STALE_READ`，**檔案未被覆寫** |
| AC-6 | 路徑逃逸攔截 | 以 junction / symlink / 絕對路徑指向 workspace 外 | 三種皆回 `E_PATH_ESCAPE` |
| AC-7 | 不打斷使用者 | 完成 AC-4 的完整流程 | 0 次權限詢問（皆為 AUTO） |
| AC-8 | 危險操作攔截 | 嘗試 `git push` / `Remove-Item -Recurse -Force` | 皆 `DENY`，不可批准 |
| AC-9 | Session 授權不重複 | 同 session 內兩次網路連線 | 第 2 次不再詢問 |
| AC-10 | 程序樹終止 | `run_powershell` 起一個會 fork 孫程序的命令並逾時 | 整棵程序樹全滅，回 `E_TOOL_TIMEOUT` |
| AC-11 | Provider retry | Mock 連續 3 次 500 | 嘗試恰好 3 次，退避 ~1s / ~3s，回 `E_PROVIDER_EXHAUSTED` 並詢問切換 |
| AC-12 | Session 保存與恢復 | `--name x` 中途離開 → `--resume x` | 摘要 / 目標 / 決策 / diff / 未完成事項齊備；授權未恢復；無舊 shell 程序 |
| AC-13 | 未命名 session 刪除 | 正常退出 | Session 目錄不存在 |
| AC-14 | 三類真實任務 | 小型 bug 修復、小功能修改、失敗測試診斷 | 3/3 成功完成並產出 CompletionReport |
| AC-15 | **Smoke benchmark** | `brunel bench --smoke` | 3 個任務在拋棄式副本中無人值守跑完；輸出成功率 / token / 耗時 / 工具成功率；CONFIRM 以上操作判失敗；硬性預算生效 |
| AC-16 | 完成判定 | 讓模型宣告完成但不跑任何驗證 | Harness 拒絕，要求補齊或揭露原因 |
| AC-17 | 安全測試全綠 | §11.2 全部案例 | 100% 通過 |
| AC-18 | 能力聲明可見 | `brunel --help` 與 README | 逐字包含 §7.5 的 speed bump 聲明 |

---

## 14. Phase 邊界聲明 [FROZEN]

> **Alpha 1 不得實作以下功能**，即使技術上可行、即使模型認為「順手就做了」：

| 預留給 | 不得提前實作 |
|---|---|
| Alpha 2 | 技能索引、技能全文載入、技能 token 成本顯示。`internal/skills/` 保留空目錄，**不得填入邏輯** |
| Alpha 3 | Subagent 派遣、writer 鎖、派遣預算 |
| Alpha 4 | 完整 benchmark runner（10–15 任務）、比較組報表、thin/thick profile 切換 |
| 未定 | MCP、plugin system、Taylor connector、LSP、web 工具、ChatGPT OAuth |
| **永不** | 驗證子系統（見 §8.4 對照表） |

架構中已埋入的擴充點（`Provider` port、`Tool` interface、`EventKind` 列舉）**不得提前填入未來 Phase 的實作邏輯**。

---

## 15. 規格補強契約

> 本節將既有需求整理為可實作、可拆 Issue、可驗收的工程契約。若本節與前文衝突，以標有 `[FROZEN]` 的既有條文為準；其餘衝突必須先走規格修訂，不得由實作者自行選擇。

### 15.1 Contract（輸入、輸出與失敗契約）

| 契約 ID | 輸入要求 | 成功輸出 | 失敗行為 |
|---|---|---|---|
| CT-1 CLI 任務 | 互動模式需有 TTY；單次模式的 task 去除前後空白後不得為空 | 串流文字；指定 `--report` 時另產生符合 §7.8 的 JSON | 空 task 回 `E_INVALID_ARGUMENT`；不得啟動 provider 請求 |
| CT-2 Workspace | 啟動目錄必須存在、可解析為絕對真實路徑且可讀；session 期間固定 | 所有檔案工具只操作該 root 內的最終真實路徑 | 無法解析回 `E_WORKSPACE_INVALID`；逃逸回 `E_PATH_ESCAPE`，不得執行 I/O |
| CT-3 Tool call | 名稱必須屬於 §7.3 的 8 個工具；參數必須通過對應 schema | 結構化 `ToolResult`，並 append 至 `events.jsonl` | 未知工具回 `E_UNKNOWN_TOOL`；缺漏或錯型參數回 `E_INVALID_ARGUMENT`；不得自行補值 |
| CT-4 寫入工具 | `apply_patch` / `write_file` 必須帶最近一次完整讀取所得 `expected_hash` | 原子性完成寫入並回傳新 hash；修改可由 `workspace_diff` 觀察 | hash 不符回 `E_STALE_READ`；patch context 不符回 `E_PATCH_CONFLICT`；任何失敗不得留下部分檔案 |
| CT-5 Provider | model ID 必須存在於 tool-capable 清單且 probe 通過；API key 僅取自 Credential Manager | 結構化文字、tool calls 與 usage | 不可重試錯誤立即失敗；可重試錯誤依 §5.1 F-7；不得靜默切換模型或退回文字模擬工具 |
| CT-6 Session resume | `--resume` 必須能唯一解析至一個可讀 session ID | 恢復 §4.1 情境 4 所列狀態，但不恢復授權或程序 | 無匹配回 `E_SESSION_NOT_FOUND`；名稱多重匹配回 `E_SESSION_AMBIGUOUS` 並列出 ID／時間，不自動挑選 |
| CT-7 Completion report | `--report` 路徑須位於 workspace 內，父目錄存在且可寫 | UTF-8 JSON，schema version `1.0`，寫入採暫存檔後原子替換 | 序列化或寫入失敗時回非零退出碼且不得留下宣稱 completed 的部分 JSON；既有檔案策略見 OQ-5 |
| CT-8 Benchmark | task fixture、模型與所有硬性預算均已明確提供 | 每案例結果及聚合指標；案例 workspace 可拋棄 | 缺少任何預算回 `E_BENCH_CONFIG_INVALID`；CONFIRM 以上直接判案例失敗；不得互動等待 |

所有公開錯誤必須至少包含穩定錯誤碼與可行的短訊息；可安全揭露時再加入 `path`、`expected_hash`、`actual_hash` 等欄位。錯誤不得包含 API key、Authorization header 或未遮罩的已知 secret。

### 15.2 Invariants（系統不變式）

| ID | 不變式 | 違反症狀 | 必要偵測／回歸測試 |
|---|---|---|---|
| INV-1 | 所有工具 I/O 前必經唯一 ApprovalGate（§6.4） | 未產生 gate 決策 event 即出現檔案、程序或網路副作用 | `TestRegistry_AllToolsPassThroughGate`；以拒絕 gate 驗證 8 個工具均無副作用 |
| INV-2 | `events.jsonl` 只 append，不改寫既有完整 event | 舊 event 位元組、順序或 ID 在摘要／resume 後改變 | `TestEventLog_AppendOnly`；摘要前後比較既有 prefix bytes |
| INV-3 | 不可信輸入不得授權或降低風險 | `AGENTS.md`、專案 config、模型或工具輸出使 DENY／CONFIRM 降級 | `TestGate_DenyIsNotOverridableByAgentsMD` 與 project config／model output 對應測試 |
| INV-4 | Workspace root 在 session 中不可變 | resume 或子目錄操作後，解析 root／volume identity 改變 | 建立同名路徑與 junction 後操作，所有解析仍綁定啟動時 root identity |
| INV-5 | 寫入不得覆蓋未知的新版本 | stale hash、patch conflict 或取消後，目標內容被改動 | stale-read、patch conflict、取消與磁碟錯誤測試均比較完整檔案 hash |
| INV-6 | Session grants 不得跨 session 或 resume 保存 | 新 session／resume 未詢問便放行 SESSION_GRANT | `TestSession_Resume_DoesNotRestoreGrants` 及跨 session 隔離測試 |
| INV-7 | 完成狀態必須有可追溯證據 | `completed` report 存在未滿足 criterion、pending approval 或不可定位 evidence | §8.3 狀態機測試逐項移除證據，均不得進入 Completed |
| INV-8 | Benchmark 不得污染 fixture 或來源 workspace | 案例後基準目錄 hash 改變，或案例間互相看到修改 | 每案例前後 fixture hash；兩案例並行／循序執行隔離測試 |

### 15.3 Edge Cases（邊界與失敗體驗）

| ID | 條件 | 預期結果 |
|---|---|---|
| EC-1 | 單次模式 task 為空字串或全空白 | `E_INVALID_ARGUMENT`，不呼叫 provider，不建立可恢復 session |
| EC-2 | 從不存在、無法讀取或無法解析的工作目錄啟動 | `E_WORKSPACE_INVALID`，訊息指出失敗階段，不 fallback 到父目錄或使用者目錄 |
| EC-3 | 路徑含中文、空白、emoji、保留字、尾端句點／空白或僅大小寫不同 | 使用 Windows 最終路徑語意一致處理；不支援者回穩定錯誤，不得誤判為 workspace 外或操作另一檔案 |
| EC-4 | Ctrl+C 發生於 provider 等待、tool 執行或 report 寫入 | 傳遞取消、終止 Job Object 程序樹、寫入中止 event；不得留下部分檔案或有效 completed report |
| EC-5 | `events.jsonl` 尾端為不完整 JSON line | 保留所有完整 event；隔離尾端殘片並回復警告；不得重寫既有完整內容。是否自動截除見 OQ-6 |
| EC-6 | 多個 session 使用相同名稱 | 名稱可建立；`--resume <name>` 回 `E_SESSION_AMBIGUOUS`，使用者須改用不可變 ID |
| EC-7 | 重複執行 create／patch／resume／cleanup | `create_file` 不覆寫；同 patch 不可重複套用；cleanup 對已不存在目標視為成功；resume 不重播副作用 |
| EC-8 | 磁碟滿、權限撤銷或防毒軟體鎖檔 | 寫入失敗並保留原檔；event/report 不得假裝成功；可安全重試時由模型重新決定 |
| EC-9 | `--report` 指向既有檔案 | Alpha 1 不得默認覆寫；正式行為待 OQ-5 裁決，裁決前回 `E_FILE_EXISTS` |
| EC-10 | 非 Windows 平台或 Windows PowerShell 5.1／CMD 啟動 | 在任何工作前回 `E_UNSUPPORTED_PLATFORM` 或 `E_PWSH_REQUIRED`，不嘗試降級執行 |
| EC-11 | Provider 回傳重複 tool-call ID、未知 finish reason 或 malformed SSE | 回 `E_PROVIDER_PROTOCOL`；記錄可遮罩診斷；不得重播可能已有副作用的 tool call |
| EC-12 | Benchmark case 中斷或超出任一預算 | 該案例標記 failed，終止程序樹並保留診斷；其他案例可繼續，來源 fixture 不變 |

### 15.4 Acceptance Criteria（需求追蹤與可驗收性）

§13 的 AC-1～AC-18 均為正式 Alpha 1 發布門檻。除既有「測量方式」與「通過標準」外，新增以下共通契約：

1. 每次驗收必須保存機器可讀結果（測試輸出、report JSON、hash 或退出碼）及執行環境資訊；只有人工文件檢查可用審查紀錄代替。
2. AC-1、AC-2、AC-4～AC-17 可自動化；AC-3 與 AC-18 為「自動化測試 + 人工抽查」。任何未自動化項目必須在發布報告記錄原因與人工證據。
3. AC-4 的「diff 正確」定義為：只包含任務要求的檔案與內容，且 `ModifiedFiles` 與 diff 路徑集合完全相同。
4. AC-9 必須先取得一次 SESSION_GRANT，並證明相同 session 的等價操作第二次為 0 次詢問；不同參數是否同類由 §7.5 的正式分類鍵決定。
5. AC-11 不得依賴真實 sleep；以可注入 clock 驗證退避區間、抖動邊界及三次上限。
6. AC-14／AC-15 只有在 OQ-4 的硬性預算已裁決並記入 fixture manifest 後，才具備可重現的通過判定；裁決前狀態為「Blocked by decision」，不得宣稱發布門檻已通過。

需求追蹤矩陣：

| Task ID | 需求 | 狀態 | 目標與範圍 | 非範圍 | 依賴 | 主要風險 | 優先級依據 | 驗收／測試 |
|---|---|---|---|---|---|---|---|---|
| TASK-A1-F01 | F-1 CLI | 正式 | REPL、單次模式、flags、TTY 行為 | TUI、其他 shell | config、session | 無 TTY 阻塞 | 所有入口前置 | AC-1、AC-13、AC-16；TC-CLI-* |
| TASK-A1-F02 | F-2 Workspace | 正式 | root 綁定與最終路徑守衛 | sandbox | workspace、Windows API | junction 逃逸 | 安全邊界 | AC-5、AC-6；TC-WS-* |
| TASK-A1-F03 | F-3 Tools | 正式 `[FROZEN]` | 8 個 schema 與執行器 | plugin、額外工具 | safety、workspace、exec | gate 旁路 | 核心閉環 | AC-4、AC-7；TC-TOOL-* |
| TASK-A1-F04 | F-4 Stale read | 正式 | hash 前置條件與衝突錯誤 | 自動 merge | workspace | 覆蓋外部修改 | 資料保護 | AC-5；TC-FILE-* |
| TASK-A1-F05 | F-5 PowerShell | 正式 | pwsh 7、Job Object、取消與資源限額 | 命令語意 sandbox | Windows API | 程序樹逃逸 | 確定性控制 | AC-10；TC-EXEC-* |
| TASK-A1-F06 | F-6 Safety | 正式 `[FROZEN]` | 四級分類與唯一 gate | 宣告式 policy engine | tools、TTY | 誤分類／提權 | 高風險核心 | AC-7～AC-9、AC-17；TC-GATE-* |
| TASK-A1-F07 | F-7 Provider | 正式 | OpenRouter、probe、retry | 多 provider、自動 fallback | HTTP/SSE、Credential Manager | 協定差異、重播 | agent 可運作性 | AC-2、AC-11；TC-PROV-* |
| TASK-A1-F08 | F-8 Agent/context | 正式 | 自主 loop、保留規則、摘要 | 強制 plan／review 階段 | provider、session | 摘要遺漏 | 核心命題 | AC-4、AC-16；TC-AGENT-* |
| TASK-A1-F09 | F-9 Session | 正式 | 保存、命名、恢復、清理、遮罩 | 加密、跨裝置同步 | filesystem、ULID | 敏感內容、損毀 | 可恢復性 | AC-12、AC-13；TC-SESSION-* |
| TASK-A1-F10 | F-10 AGENTS.md | 正式 | root／就近規範載入 | 權限授予 | context、workspace | prompt injection | 工作規則一致性 | AC-3；TC-INSTR-* |
| TASK-A1-F11 | F-11 Config | 正式 | 四層優先序與憑證分離 | 專案憑證、降級紅線 | Credential Manager | 惡意 config | 安全啟動 | AC-2、AC-17；TC-CONFIG-* |
| TASK-A1-F12 | F-12 Completion | 正式 | 證據存在性狀態機 | 驗證知識／pipeline | diff、tool events | 假完成 | 可信交付 | AC-16；TC-COMP-* |
| TASK-A1-F13 | F-13 Report | 正式 `[FROZEN]` | §7.8 JSON 輸出 | Taylor 寫入或雙向整合 | completion、filesystem | schema 漂移 | 公開證據介面 | AC-14～AC-16；TC-REPORT-* |
| TASK-A1-F14 | F-14 Smoke bench | 正式 | 3 任務、隔離副本、預算、指標 | 10–15 任務與比較組 | agent、Job Object、provider usage | 不可重現、污染 | 核心命題量尺 | AC-14、AC-15；TC-BENCH-* |
| TASK-A1-F15 | F-15 Context ledger | 候選 | token 來源分布 | 歷史統計 | context、provider usage | provider usage 粒度不足 | Should，不擋發布 | 候選 AC：來源合計等於送出 context token |
| TASK-A1-F16 | F-16 Prompt 透明化 | 候選 | 顯示實際 prompt 與 token | 修改安全核心／工具契約 | context | secret 洩漏 | Should，不擋發布 | 候選 AC：顯示內容與實際送出內容 hash 一致且已遮罩 secret |
| TASK-A1-F17 | F-17 非 Git diff | 候選 | session 快照差異 | 完整 VCS 抽象 | workspace、session | 大 workspace 成本 | Should，不擋發布 | 候選 AC：新增／修改／刪除與基準快照一致 |

候選 F-15～F-17 不得因出現在矩陣中而升級為 Alpha 1 發布需求；F-18 以後維持 §5.3～§5.4 與 §14 的排除狀態。

### 15.5 Test Plan（測試計畫與禁止行為）

在 §11 的層次之外，測試案例採以下穩定 ID 前綴：

| 前綴 | 範圍 | 最低要求 |
|---|---|---|
| TC-CLI | 參數、TTY、退出碼、平台檢查 | 空輸入、無 TTY、取消、非支援平台 |
| TC-WS／TC-FILE | 路徑、hash、patch、原子寫入 | junction、symlink、非 ASCII、case alias、stale read、磁碟錯誤 |
| TC-GATE | 分類、授權記憶、不可提權 | 8 工具全路徑、每個風險級、跨 session 隔離 |
| TC-EXEC | Job Object、timeout、cancel、限額 | 子／孫程序全滅、先綁後 resume、輸出截斷 |
| TC-PROV | SSE、tool call、probe、retry | 4xx／5xx／429、malformed stream、重複 ID、取消 |
| TC-SESSION | append、resume、cleanup、損毀恢復 | 同名、尾端殘片、72 小時邊界、grant 不落盤 |
| TC-COMP／TC-REPORT | 狀態機、schema、原子輸出 | 每個 completed 前置條件的單獨反例、JSON schema 相容性 |
| TC-BENCH | 隔離、預算、指標 | fixture hash、不互動、各預算越界、案例失敗後續跑 |

下列測試保護 `[FROZEN]` 契約，任何修改都需 spec revision：`TC-GATE-001`（唯一 gate）、`TC-GATE-002`（DENY 不可覆寫）、`TC-SESSION-001`（append-only）、`TC-PROV-001`（結構化 tool call，不得文字模擬）、`TC-REPORT-001`（schema 1.0）、`TC-COMP-001`（證據狀態機）、`TC-PHASE-001`（Alpha 1 禁止模組／依賴）。

發布驗證順序為：靜態／schema 檢查 → unit → integration → Windows E2E → 乾淨 VM 可攜性 → smoke benchmark。前一層失敗時不得以後一層成功抵銷。

**禁止的測試行為**：

- 不得以 `|| true`、忽略退出碼、跳過失敗案例或只記 log 的方式製造通過。
- 不得只斷言錯誤訊息文字；至少斷言穩定錯誤碼、無副作用與必要結構欄位。
- 不得讓 unit／integration 測試呼叫真實 OpenRouter；只有明確標記且有預算的 smoke benchmark 可用真實 API。
- 不得共用使用者的真實 session、Credential Manager 項目或來源 workspace；測試必須使用隔離暫存資源並清理。
- 不得在安全測試中以 mock 掉 ApprovalGate 的方式宣稱 INV-1 已驗證；至少一組 integration 測試必須走真實 gate。

### 15.6 FROZEN（凍結決策與變更同步）

既有 `[FROZEN]` 內容維持不變，並統一受以下變更程序約束：

| 凍結範圍 | 變更時必須同步更新 |
|---|---|
| §6.3 依賴方向、§6.4 不變式 | 架構圖、module imports 測試、TC-GATE／TC-PHASE、風險表 |
| §7.1～§7.7 介面與 schema | Go 型別、序列化 fixture、工具 schema snapshot、錯誤契約、相容性說明 |
| §7.8 CompletionReport 1.0 | schema version、golden JSON、讀取端相容性測試、README 範例、修訂記錄 |
| §8.4 驗證界線 | NG-10／NG-23、completion 測試、CLI help／README 能力聲明 |
| §14 Phase 邊界 | 目錄／依賴禁止測試、roadmap、Non-Goals、修訂記錄 |

變更程序：提出 revision → 列出相容性與遷移影響 → 更新上述同步面 → 使用者明確裁決 → 提升規格版本。未完成任一步驟，不得合併與凍結契約衝突的實作。

### 15.7 Drift Risk（規格漂移風險）

| ID | 漂移面 | 早期訊號 | 防漂移控制 |
|---|---|---|---|
| DR-1 | Tool schema | schema、Go 型別與 prompt 描述欄位不同 | snapshot／golden tests；唯一 schema 來源；TC-REPORT-001 |
| DR-2 | Safety classification | 新命令樣式被分散判斷，或設定可降低紅線 | 所有分類集中於 safety；TC-GATE-001／002；code review checklist |
| DR-3 | CLI 與文件 | flags、預設值、退出碼與 README 不一致 | CLI help golden test；每次 flag 變更同步 §5.1 與 README |
| DR-4 | Session format | writer 與 resume reader 對 event kind／欄位理解不同 | event schema version、round-trip fixture、舊版 fixture 相容測試 |
| DR-5 | Completion evidence | report 欄位存在但 evidence 不可定位 | criterion-to-event/diff reference 驗證；TC-COMP-001 |
| DR-6 | Benchmark | fixture、模型、預算或成功判定未版本化 | manifest 含 hash、model ID、預算與 evaluator version；結果攜帶 manifest hash |
| DR-7 | Phase creep | `internal/skills`、subagent、MCP 等預留點出現邏輯 | TC-PHASE-001 掃描禁止依賴／入口；發布審查 §14 |
| DR-8 | Windows 行為 | 開發機通過但乾淨 VM、不同磁碟或非 ASCII 使用者目錄失敗 | 乾淨 VM matrix；不同 volume／長路徑／Unicode E2E |

### 15.8 Open Questions（待確認，不得自行定案）

| ID | 待裁決事項 | 影響 | 裁決前行為 |
|---|---|---|---|
| OQ-1 | A-1～A-8 是否成立 | build、CLI、成本、benchmark、工具實作 | 保持假設狀態；不得標記相關決策 `[FROZEN]` |
| OQ-2 | Windows 最低支援映像是否固定為 Windows 10 21H2，乾淨 VM 驗收是否只用 Windows 11 | AC-1、Job Object 相容性 | 兩者均列測試候選；發布聲明不得超出已驗證版本 |
| OQ-3 | `SESSION_GRANT` 的「同類操作」分類鍵 | AC-9、安全體驗 | 只允許完全相同的正式分類鍵命中，不做模糊歸類 |
| OQ-4 | benchmark 的時間、token、費用、程序數、記憶體硬上限 | AC-14、AC-15 是否可判定 | benchmark 可開發但發布門檻維持 blocked，不得使用無上限預設 |
| OQ-5 | `--report` 遇到既有檔案要拒絕，或新增明確 overwrite flag | CT-7、EC-9 | 回 `E_FILE_EXISTS`，不覆寫 |
| OQ-6 | `events.jsonl` 尾端殘片是否可自動截除 | session recovery、append-only 解釋 | 不改寫原檔；隔離殘片並以完整 event 恢復，回警告 |
| OQ-7 | F-15～F-17 是否納入 Alpha 1 實作承諾 | scope、排程 | 維持候選，不阻塞 Alpha 1 |
| OQ-8 | `docs/open-questions.md` 尚不存在；是否建立並作為決策單一來源 | 文件治理 | 本節為暫時來源，不建立新檔 |

---

## 16. 未解問題（預定同步至 `docs/open-questions.md`）

| # | 問題 | 阻塞哪個 Phase |
|---|---|---|
| Q-1 | ChatGPT 訂閱 OAuth 是否存在適合公開第三方 CLI 的正式授權方式（服務條款層面） | 實驗項目，不阻塞任何 Phase |
| Q-2 | 是否需要 LSP — 待 benchmark 證明文字搜尋 + 編譯器診斷不足 | Alpha 4 之後 |
| Q-3 | 是否需要專用 web 工具 | Alpha 4 之後 |
| Q-4 | `brunel.exe` 是否需程式碼簽章（SmartScreen） | 正式版發布 |

---

## 17. 修訂記錄

| 版本 | 日期 | 修改內容 | 作者 |
|---|---|---|---|
| v1.1 | 2026-07-13 | 依 `maze-spec-hardening` 補入 Contract、Invariants、Edge Cases、Acceptance Criteria、Test Plan、FROZEN、Drift Risk、Open Questions 八個工程契約區塊；新增 TASK-A1-F01～F17 追蹤矩陣、測試 ID／禁止行為及可觀察失敗契約。未將候選需求或待確認假設升級為正式決策。 | Codex |
| v1.0 | 2026-07-13 | 初版建立。已納入四項裁決：benchmark smoke tasks 進 Alpha 1 發布門檻、安全能力聲明降級為 risk speed bump、ADR-001 採用 Go + Job Object、ChatGPT OAuth 移出里程碑。Watt 解耦寫入 §1.3 與 §8.4。 | Maze + Claude |

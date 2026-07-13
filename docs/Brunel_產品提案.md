# Brunel：面向前沿模型的薄型 Coding Harness CLI 提案

## 1. 提案摘要

Brunel 是一個從零建構、Windows-first、可公開安裝的 coding harness CLI。它的核心假設是：當前沿模型的自主推理與工具使用能力已足夠強時，coding harness 不需要以大量固定工作流、強制規劃、角色提示或技能規則替模型做決定；框架應退回到一個精簡但可靠的位置，提供完整的工作工具、安全與執行邊界、透明的上下文管理，以及可驗證的完成機制。

Brunel 不等於「沒有框架」或「完全信任模型」。再強的模型仍可能誤判或失手，因此安全紅線、工作區界線、危險操作批准、程序控制、錯誤處理與變更證據仍由 harness 確定性地管理。除此之外，模型可以自由決定如何探索、修改、驗證，以及何時載入技能或派遣 subagent。

Brunel 源自 Taylor 專案的產品脈絡，但會以獨立 repository、獨立 CLI、獨立版本與 Apache License 2.0 發布。待 Brunel 成熟後，再評估作為 Taylor coding harness connector／外掛的整合方式；外掛系統不屬於目前 MVP。

## 2. 問題與機會

現有 coding harness 常透過固定階段、強制計畫、龐大系統提示、技能自動觸發與代理角色限制來提高可靠性。這些設計對較弱模型可能有幫助，但用在前沿模型上也可能帶來額外成本：

- 占用上下文與 token，增加費用。
- 干擾模型原本可自行完成的推理與工具選擇。
- 增加等待、確認與重複工作。
- 讓框架本身變得難以理解、維護與量化。
- 難以判斷成果來自模型能力，還是框架堆疊。

Brunel 首要驗證「薄 harness 本身」是否成立。次要方向是讓旗艦模型擔任技術顧問，負責規劃、派工與驗收，由成本較低的模型執行明確子任務，以降低費用並提高效率；此能力不阻塞第一個 alpha。

## 3. 產品原則

### 3.1 模型負責思考，Harness 負責邊界

模型負責任務策略、工具選擇、工作順序、技能載入與未來的 subagent 派遣。Harness 負責工具協定、權限、安全、程序生命週期、上下文預算、錯誤回報與完成證據。

### 3.2 不強制規劃

Brunel 不要求每個任務先建立 plan，也不強制固定的 plan／implement／review／verify 階段。模型可以直接執行安全且可逆的必要工作；只有遇到高風險、範圍模糊或真正會改變產品方向的分支時，才需要說明或詢問。

### 3.3 少打斷使用者

工作區內的一般讀寫與開發命令預設自動允許。可預見的同類權限應合併成一次、範圍明確的 session 授權，避免反覆詢問「可不可以」。不可逆刪除、外部發布、部署、敏感憑證、權限提升與安全策略繞過仍受硬性管理。

### 3.4 工具完整，但數量精簡

模型不能被放進一個什麼都沒有的辦公室。Brunel 提供完成「理解、修改、驗證」閉環所需的固定內建工具，但不在首版引入 MCP、通用 plugin system、瀏覽器、專用網路搜尋或 LSP。

### 3.5 透明、可追溯、可恢復

框架可以主動裁剪上下文並觸發摘要，但所有被移出的內容都必須保留於本機紀錄，可按需重新載入。框架不得在不可追溯的情況下靜默遺失重要資訊。

## 4. 目標使用者與平台

- 主要使用者：熟悉 coding agent 的進階個人開發者。
- 使用情境：單機、單使用者、本地 coding 工作。
- 首要平台：Windows x64。
- 正式 shell：PowerShell 7（`pwsh`）。
- 實作基線：TypeScript／Node.js 22。
- 發布形式：GitHub pre-release，提供獨立 `brunel.exe` 與 Windows x64 壓縮包。
- 預定工作目錄：`D:\AgentCoding\Brunel`，獨立 repository。

## 5. Alpha 1 功能範圍

### 5.1 互動方式

- `brunel`：在目前目錄啟動互動式 REPL。
- `brunel "任務內容"`：執行單次任務。
- `workspace`：預設模式，可在固定 workspace root 內修改與執行命令。
- `readonly`：只允許讀取、搜尋與分析。
- `benchmark`：後續用於拋棄式副本的無人值守測試。
- 首版不做全螢幕 TUI。

有 TTY 時，單次命令模式可就地處理批准；沒有 TTY 時不得卡住等待，缺少授權便以明確錯誤及非零退出碼結束。

### 5.2 固定內建工具

- 專案導覽與檔案搜尋。
- 文字搜尋。
- 分段檔案讀取，附行號與內容雜湊。
- `apply_patch` 式精確修改。
- 僅在不存在時建立新檔案。
- 受限制的完整寫檔；覆寫大型既有檔案或大量檔案需提高風險等級。
- 受控 PowerShell 7 執行、逾時、取消及程序樹終止。
- Workspace diff：Git repository 使用 Git 資訊；非 Git 專案依 session 快照顯示 Brunel 修改。

模型可執行任意 PowerShell 字串，但 Harness 只能承諾風險攔截，不能宣稱已對完整 PowerShell 語言提供絕對安全分類。結構化檔案工具應優先於 shell 寫檔。

### 5.3 明確排除

Alpha 1 不包含：

- MCP。
- 通用第三方 plugin system。
- Taylor connector 或 Taylor 程式碼修改。
- 專用 web 搜尋、瀏覽器或網頁讀取工具。
- LSP；未來只有 benchmark 證明文字搜尋與編譯器診斷不足時才重新評估。
- 完整 Git 抽象、worktree、PR 或 GitHub 整合。
- 自動 commit、push、發布或部署。
- macOS、Linux、Windows PowerShell 5.1、CMD、Git Bash 或 WSL 的正式支援。
- 跨裝置 session 同步。

## 6. 安全與批准模型

Brunel 採「少量硬性紅線＋使用者可調政策」。安全核心不可由模型、`AGENTS.md`、技能或專案設定覆寫。

### 6.1 自動允許

- Workspace root 內的一般讀取、搜尋、patch、建檔與開發命令。
- 測試、lint、typecheck、build 及唯讀 Git 操作。

### 6.2 可作 Session 範圍授權

- 安裝依賴。
- 網路連線。
- 啟動長程序。
- 指定工作區外路徑的唯讀或明確範圍寫入。

同一 session 內，已授權的同類操作不重複詢問。模型已預見多項權限需求時可集中申請，但 Brunel 不強制先做完整計畫。

### 6.3 高風險或硬性紅線

- 不可逆刪除或資料覆寫。
- Git push、發布、部署或對外傳送內容。
- 讀取、輸出或提交憑證與敏感資料。
- 權限提升、沙箱逃逸、安全政策關閉或稽核紀錄隱藏。

工作區在 session 開始時固定。需檢查絕對路徑、junction／symlink 逃逸，以及檔案在讀取後是否被使用者或其他程序修改。Brunel 不得靜默覆蓋使用者既有變更。

## 7. 模型與 Provider

### 7.1 正式基線

- Alpha 1 正式支援 OpenRouter API key。
- Provider 層保留小型 adapter 邊界，但不追求多供應商最低共同能力。
- 不寫死、不推薦任何模型名稱，讓使用者自由選擇。
- 模型清單只顯示 OpenRouter metadata 標示支援結構化工具呼叫的模型。
- 首次使用模型時執行一次無副作用 tool-call probe，通過後在本機短期快取相容性結果。
- 不允許以純文字模擬工具呼叫混入正式 agent loop。

### 7.2 ChatGPT OAuth 實驗

ChatGPT 訂閱 OAuth 可作為後續 experimental MVP，但不得成為正式核心依賴。若可行，應以可拔除 adapter 實作，主動啟用、明確標示非穩定能力，token 存入 Windows Credential Manager；失效時不得破壞 OpenRouter 基線。是否存在適合公開第三方 CLI 的正式授權方式仍需另行驗證。

### 7.3 重試與切換

- 最多共嘗試 3 次，包含第一次請求。
- 第一次失敗後約等待 1 秒；第二次失敗後約等待 3 秒，加入隨機抖動。
- 若 provider 回傳 `Retry-After` 則優先遵守，但單次等待上限 30 秒。
- 認證、額度、模型不存在或明確格式錯誤不做無效重試。
- 三次皆失敗後報錯；若有其他相容模型可用，再詢問使用者是否切換。
- 不靜默切換模型；benchmark 模式不 fallback。

## 8. 上下文與 Session

### 8.1 上下文管理

採「確定性保留規則＋模型摘要」：

框架保留使用者指令、安全政策、當前目標、使用者決策、已修改檔案、最新 diff、驗證結果、未解錯誤與未完成事項。重複內容、過時搜尋、已被取代的工具輸出及冗長成功 log 可自動裁剪；原始 event log 仍留在本機，可按需重載。

### 8.2 Session 保存規則

- Session 預設為暫存。
- 使用者可在啟動時或進行中命名；只有已命名 session 持久保存。
- 未命名 session 正常退出後自動刪除。
- 異常中止的未命名 session 短期保留，供下次啟動救援，逾期自動清理。
- Session 名稱不必唯一，內部使用不可變 ID。
- 恢復時先載入結構化摘要、目標、決策、diff、驗證與未完成事項；完整歷史按需載入。
- 不恢復舊 shell 程序，也不永久保存危險操作授權。

Session 預定存於 `%LOCALAPPDATA%\Brunel\sessions`。Alpha 1 不自行加密 session；憑證與 session 分離，API key／token 只存 Windows Credential Manager。落盤前盡力遮罩已知 secret，但不宣稱能偵測所有敏感資訊。

## 9. 專案指令、設定與技能

### 9.1 `AGENTS.md`

Brunel 直接使用既有 `AGENTS.md`，不另創專案指令格式。啟動時讀取 workspace root 規則，操作子目錄檔案前按需讀取更接近的 `AGENTS.md`。較近規則優先，但 `AGENTS.md` 只能約束工作方式，不能授予權限或關閉安全紅線。

### 9.2 設定分層

- 全域：`%USERPROFILE%\.brunel\config.json`
- 專案：`<workspace>\.brunel\config.json`
- 臨時覆寫：CLI flags
- 憑證：Windows Credential Manager

優先序為 CLI flags、專案設定、全域設定、Brunel 預設。專案設定不能存憑證，也不能降低硬性安全政策。

安全核心與工具真實契約固定；一般 agent prompt 可檢視及自訂，並可顯示提示來源與 token 占用。Brunel 不強制規劃。

### 9.3 技能

技能是後續里程碑。Brunel 只讀自己的技能目錄，不掃描 `.codex`、`.agents`、`.claude` 或其他 harness：

- 全域：`%USERPROFILE%\.brunel\skills\<name>\SKILL.md`
- 專案：`<workspace>\.brunel\skills\<name>\SKILL.md`

專案技能優先於同名全域技能。Session 開始只向模型提供名稱、用途、觸發條件與預估載入成本的精簡索引；選定後才載入全文及必要 references。使用者明確指定時必須載入。技能不能擴張工具權限。

## 10. Subagent 後續方向

Subagent 不阻塞 Alpha 1。成熟後採以下邊界：

- 主模型決定派遣、任務、模型與驗收標準。
- Harness 控制總預算、並行數、逾時與最大派遣深度。
- Subagent 只取得必要檔案、工具、規則與技能。
- 預設不允許 subagent 再派遣 subagent。
- 共享工作區同時只能有一個 writer，其他 subagent 唯讀分析。
- 主模型必須驗收 subagent 結果，不把其自我宣告直接當成完成。
- Subagent 不得繞過主 session 的安全與批准政策。

這將作為日後「大模型擔任技術顧問，帶領便宜模型執行」成本效益實驗的基礎。

## 11. 完成判定

模型先宣告本次驗收條件，Harness 檢查是否存在對應證據；不只相信模型的文字完成宣告。完成前至少確認：

- 使用者需求已逐項對應。
- 沒有未回報的工具或程序失敗。
- 所有修改可產生 diff。
- 已執行相關測試、lint、typecheck 或 build；未執行項目與原因已揭露。
- 沒有待處理的危險操作批准。
- 最終回覆包含成果、驗證與剩餘風險。

Harness 不硬編每種專案應執行哪些驗證命令，但會要求可檢查的證據。

## 12. Benchmark 與成功衡量

Brunel 不收集使用者遙測。所有 session、成本與評估資料預設只留在本機，不提供預設上傳；未來若分享結果，必須由使用者明確匯出。

後續提供純本機 benchmark runner：

- 日常只跑約 3 個快速 smoke tasks。
- 發布或重大架構變更時，無人值守執行 10–15 個完整任務。
- 每個案例使用專用、可拋棄的 workspace 副本。
- 危險操作直接使該案例失敗，不等待批准。
- 限制時間、token、費用、程序數與磁碟用量。
- MVP 階段採目錄與程序層限制，不宣稱完整 OS sandbox。

比較組：

1. 單一模型＋Brunel 薄 harness。
2. 同一模型＋較厚的工作流與技能規則。
3. 主模型規劃／驗收＋便宜模型執行（次要實驗）。

衡量指標包含任務成功率、測試通過率、首次完成率、返工次數、token、費用、耗時、技能上下文占用、工具成功率、人工介入、安全攔截與誤攔截。日常 session 預設只追蹤成本，使用者可自行設定上限；benchmark 必須設定硬性成本與時間上限。

## 13. Alpha 1 發布門檻

Brunel 達到以下最低限度即可發布 alpha：

- 在乾淨 Windows x64 環境啟動 `brunel.exe`，不要求使用者先建立 Node.js 開發環境。
- 可設定 OpenRouter 憑證並選擇經 tool-call probe 驗證的模型。
- 可讀取適用的 `AGENTS.md`。
- 可搜尋、讀取、patch、建檔、執行 PowerShell 7 並顯示 workspace diff。
- 可偵測檔案在讀取後遭外部修改，避免靜默覆蓋。
- 工作區內一般操作不中斷；危險操作會攔截，且同類 session 授權不重複詢問。
- Provider 最多共嘗試三次，失敗後清楚報錯並可詢問切換模型。
- 已命名 session 可保存並以摘要恢復；未命名 session 正常退出後刪除。
- 至少完成三類真實任務：小型 bug 修復、小功能修改、失敗測試診斷。
- 安全紅線、錯誤處理及資料保護的必要測試通過。

技能、subagent、benchmark runner、ChatGPT OAuth、Taylor connector、外掛、LSP 與專用 web 工具均不阻塞 Alpha 1。

## 14. 建議里程碑

1. **Alpha 1：單一主模型 coding 閉環**  
   完成 OpenRouter provider、agent loop、固定內建工具、workspace／readonly 模式、安全批准、上下文管理、命名 session 與 Windows binary。
2. **Alpha 2：技能按需載入**  
   加入 Brunel 專屬技能索引、全文載入、來源與 token 成本透明化。
3. **Alpha 3：Subagent**  
   加入主模型派工、單一 writer、預算與驗收機制。
4. **Alpha 4：本機 Benchmark**  
   加入快速組、完整組、拋棄式副本、硬性預算與比較報告。
5. **實驗項目**  
   評估 ChatGPT OAuth adapter；以實際證據決定是否加入 LSP 或專用 web 工具。
6. **成熟後整合**  
   根據 Brunel 的真實事件與控制需求，設計 Taylor connector 與外掛機制。

## 15. 核心風險

- PowerShell 是完整程式語言，靜態危險分類不可能絕對可靠。
- OpenRouter metadata 不保證所有路由後端都能正確完成工具迴圈，因此需要實際 probe。
- 模型摘要可能遺漏細節，必須保留可重載的原始事件。
- Prompt injection 可經 repository、`AGENTS.md`、技能、命令輸出或未來網路內容進入模型；這些來源都不能授權工具。
- 未加密 session 可能包含程式碼與操作紀錄，需清楚揭露並提供刪除機制。
- Windows 單檔打包、程序樹終止、Credential Manager 與路徑逃逸需在真實環境驗證。
- 「薄」不代表自然更好；必須以 benchmark 證明品質沒有因較少框架介入而下降。

## 16. 結論

Brunel 的定位不是另一個用更多規則包住模型的 coding agent，而是一套給強模型使用的精簡工作台：工具足夠、安全邊界清楚、上下文透明、使用者少被打斷，並以可驗證成果而非固定工作流判定完成。

下一步應把本提案轉成 Alpha 1 工程規格，定義 agent loop、工具 schema、安全分類、session schema、provider adapter 與驗收測試，再於 `D:\AgentCoding\Brunel` 建立獨立專案。建立 repository、執行 Git 初始化或開始實作前，仍應取得使用者明確授權。

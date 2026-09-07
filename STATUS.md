# Brunel — 當前狀態

> 最後同步：2026-09-07
> Branch：maze/2026-09-07-23b374
> Working tree：乾淨

## 架構轉向

- [ADR-002](docs/adr/ADR-002-pi-agent-runtime.md)（2026-08-12）：放棄零依賴單檔 exe 需求（ADR-001 硬需求 (a)、spec.md 原 G-1），改採 Pi 作為 model-facing Agent Runtime（Route B）；ADR-001 部分 Superseded，Job Object／Workspace／Safety／PowerShell 執行器與 Go 1.25.x／Bubble Tea v2 基線仍維持 Go 實作。依據 [#24 Pi Compatibility Spike](https://github.com/bext1998/brunel/issues/24) 的 Gate 1/3/4 Pass、Gate 2 Pass（有但書）、Gate 0 Partial；Spike 分支 `agent/pi-spike-issue-24` 已 push 至遠端（不合併）。
- `docs/spec.md` 已同步修訂至 **v1.3**：8 個工具全留 Go（經 `taylor-tools.ts` extension 暴露給 Pi）、Session 以 Brunel 自己的 `events.jsonl` 為準（Pi session 停用）、Provider 開放多家（不再限定 OpenRouter，交由 Pi 生態決定）。新增 INV-9（`internal/pirpc` 禁止送出 `bash` RPC command）、OQ-8／OQ-9。
- 直接受影響：#8（F-7 Provider Adapter）已依 v1.3 §4 矩陣新描述完成核心實作（`internal/pirpc`，見下方進行中 Issues）；#9（F-8 Agent Loop/EventSink/context）仍待依新描述重新拆解為實作任務。Route B 正式整合工作尚未完全拆解為 Issues：(1) 未安裝 Git Bash 的 Windows VM/runner 補測 Gate 0（OQ-9）、(2) `taylor-tools.ts` 尚未實作（`internal/pirpc` 已由 #8 起頭）、(3) #9 拆解為正式整合 Issue，取代其原本範圍。

## 進行中 Issues

- [#1 Alpha 1：薄型 coding harness 實作追蹤](https://github.com/bext1998/brunel/issues/1) 已依 v1.2 對齊；未完成子項為 #2、#4、#5、#7、#8、#9、#11、#14、#22。**#9 範圍受 ADR-002 影響，待重新檢視。**
- [#5 F-4：實作 stale-read hash 防護與原子寫入](https://github.com/bext1998/brunel/issues/5) 已在 [PR #26](https://github.com/bext1998/brunel/pull/26)（`internal/filetools`）完成實作，等待 review／合併中，暫緩處理。
- [#7 F-6：實作 AUTO／CONFIRM 事故防護與 Approver](https://github.com/bext1998/brunel/issues/7) 已在 [PR #27](https://github.com/bext1998/brunel/pull/27)（`internal/safety`）完成實作，等待 review／合併中，暫緩處理。
- [#8 F-7：實作 Provider Adapter（Pi delegated，ADR-002）](https://github.com/bext1998/brunel/issues/8) 已在 `internal/pirpc`（新套件）完成實作並等待 review：`BuildArgs` 把使用者選擇的 provider／model 透傳為 `pi --mode rpc` 的凍結啟動參數；`InjectCredentials` 依 Pi 既有慣例（`OPENROUTER_API_KEY` 等，經 Issue #24 Spike 分支 `recon-credentials.ps1` 確認）以環境變數注入憑證，不寫入專案檔；`TranslateProviderError` 把認證／額度／模型不存在／協定錯誤轉譯為 Brunel 錯誤碼，不自行重試；已加入 TC-PIRPC-001（INV-9 `bash` RPC command 禁止清單原始碼掃描）保護未來變更。README 已補上 provider 覆蓋範圍揭露。不含 Pi 子行程啟動／生命週期管理與 RPC event 轉譯（屬 #9）。
- [#22 F-13：建立 Alpha 1 三類 E2E fixtures](https://github.com/bext1998/brunel/issues/22) 已新增；#2、#7、#9、#14 已分別同步薄型 TUI、事故防護、EventSink 與客觀 CompletionReport 範圍。

## 阻塞 Issues

- 無規格決策阻塞 Alpha 1 實作。
- #13（完成證據狀態機）與 #15（Smoke Benchmark Runner）已依 v1.2 以 `not planned` 關閉。
- `docs/spec.md` §5（架構與公開介面）、§9（Contract）等章節的 Route B 正式修訂，待整合設計完成後才能排入 #8/#9 後續工作。

## 等待 Review

- [PR #26](https://github.com/bext1998/brunel/pull/26)（#5 F-4 stale-read hash 防護與原子寫入）：`internal/filetools` 實作，分支 `maze/2026-09-05-c207b9`。
- [PR #27](https://github.com/bext1998/brunel/pull/27)（#7 F-6 AUTO／CONFIRM 事故防護與 Approver）：`internal/safety` 實作，分支 `maze/2026-09-07-b352e0`。
- #8（F-7 Provider Adapter，ADR-002）：`internal/pirpc` 實作，分支 `maze/2026-09-07-23b374`，PR 待建立。

## 等待 Merge

- 無。

## 已合併待關閉

- 無。

## 最近完成

- [ADR-002](docs/adr/ADR-002-pi-agent-runtime.md) 確認：完成 Issue #24 Pi Compatibility Spike（5 Gate，證據存於已 push 的 `agent/pi-spike-issue-24` 分支）並據此裁決放棄零依賴需求、轉向 Route B。
- PR #23（v1.2 規格與相關文件對齊）已合併至 `main`。
- 完成 v1.2 GitHub Issue 同步：更新 #1、#2、#4、#5、#7、#8、#9、#11、#14，關閉 #13／#15，新增 #22；候選 F15～F17 未建立。
- 完成並取得使用者裁決的 Alpha 1 v1.2 規格：安全收斂為事故防護、加入薄型 TUI、簡化完成報告並將 benchmark runner 移回 Alpha 4。
- #6（F-5 PowerShell Job Object 執行器）已透過 PR #20 合併至 `main` 並關閉；經三輪 review 修正 pipe read handle 重複關閉、`TerminateJobObject` 錯誤處理與有界等待（含 pipe drain）、handle 繼承 mutex 範圍不足。
- #3（F-2 Workspace）已透過 PR #19 合併至 `main` 並關閉（root 真實路徑／identity 綁定、junction／絕對路徑／symlink 逃逸攔截與 TC-WS 測試全數通過）。
- #10／#12 已分別透過 PR #16／#17 合併至 `main` 並關閉。
- 完成 Alpha 1 v1.1 規格補強；已由 v1.2 取代。
- 建立 Git、GitHub 與 Maze 專案治理基礎。
- 依 `docs/spec.md` 需求追蹤矩陣建立 #1～#15、結構化標籤與原生父子關係。

## 未追蹤本機工作

- PR #21（GitHub Actions CI workflow，`.github/workflows/ci.yml`）已合併至 `main`，無對應 Issue；每次 push／PR 自動跑 `go build`／`go vet`／`go test`／零 CGO Windows build。

## 已知驗證限制

- v1.2 AC-7 的 stale-read 防護（#5）已完成核心邏輯與單元測試（PR #26），尚未經 #4 接上實際 workspace／8 工具 schema 做 E2E／Integration 驗證。
- AC-9～AC-11（AUTO 體驗、CONFIRM 分類、readonly／無 TTY）對應的 #7 已完成 `internal/safety` 核心決策邏輯與單元測試（PR #27），正式判定需待 #4（工具接線）、#2（TUI／純文字 Approver 實作）完成後的整合測試。
- AC-4（Provider 與憑證）對應的 #8 已完成 `internal/pirpc` 的 provider／model 透傳、憑證環境變數注入與錯誤轉譯核心邏輯與單元測試，但未經真實 Pi RPC 子行程驗證（子行程啟動／管理屬 #9）；`OPENROUTER_API_KEY` 等憑證環境變數命名依 Issue #24 Spike 分支的偵測結果，未對照 Pi 正式文件逐一確認。
- `internal/exec` 的 Timeout／MaxProcesses／MaxMemoryBytes／MaxOutputBytes 一律由呼叫端明確提供，套件本身不內建預設值；Alpha 1 不再需要 benchmark 硬性預算。
- repository 的 `go.mod` 目前仍是 Go 1.22；本次依約不修改程式碼或依賴，後續實作 TUI 前需另行同步至 Go 1.25.x。

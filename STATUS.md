# Brunel — 當前狀態

> 最後同步：2026-07-14
> Branch：maze/2026-07-14-12aa7b
> Working tree：乾淨；v1.2 規格與相關文件對齊修訂已提交於 PR #23

## 進行中 Issues

- [#1 Alpha 1：薄型 coding harness 實作追蹤](https://github.com/bext1998/brunel/issues/1) 已依 v1.2 對齊；未完成子項為 #2、#4、#5、#7、#8、#9、#11、#14、#22。
- [#22 F-13：建立 Alpha 1 三類 E2E fixtures](https://github.com/bext1998/brunel/issues/22) 已新增；#2、#7、#9、#14 已分別同步薄型 TUI、事故防護、EventSink 與客觀 CompletionReport 範圍。

## 阻塞 Issues

- 無規格決策阻塞 Alpha 1 實作。
- #13（完成證據狀態機）與 #15（Smoke Benchmark Runner）已依 v1.2 以 `not planned` 關閉。

## 等待 Review

- PR #23（v1.2 規格與相關文件對齊）已建立 draft，等待 review。

## 等待 Merge

- 無。

## 已合併待關閉

- 無。

## 最近完成

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

- v1.2 AC-7 的 stale-read 防護屬 #5（F-4），尚未實作。
- `internal/exec` 的 Timeout／MaxProcesses／MaxMemoryBytes／MaxOutputBytes 一律由呼叫端明確提供，套件本身不內建預設值；Alpha 1 不再需要 benchmark 硬性預算。
- repository 的 `go.mod` 目前仍是 Go 1.22；本次依約不修改程式碼或依賴，後續實作 TUI 前需另行同步至 Go 1.25.x。

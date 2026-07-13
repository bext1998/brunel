# Brunel — 當前狀態

> 最後同步：2026-07-14
> Branch：main
> Working tree：乾淨，無進行中的本機變更

## 進行中 Issues

- [#1 Alpha 1：薄型 coding harness 實作追蹤](https://github.com/bext1998/brunel/issues/1)，含 #2、#4、#5、#7、#8、#9、#11、#13、#14、#15 等尚未完成的原生子項。
- 對應 `TASK-A1-F01`／`F-3`／`F-4`／`F-6`～`F-8`／`F-10`／`F-12`～`F-14`；均標記為 `priority: P1` 並指派給 `bext1998`。

## 阻塞 Issues

- `docs/spec.md` §15.8 列出的 Open Questions 尚待使用者裁決。
- #15（Smoke Benchmark Runner）受 OQ-4 的硬性預算裁決阻塞。

## 等待 Review

- 無。

## 等待 Merge

- 無。

## 已合併待關閉

- 無。

## 最近完成

- #6（F-5 PowerShell Job Object 執行器）已透過 PR #20 合併至 `main` 並關閉；經三輪 review 修正 pipe read handle 重複關閉、`TerminateJobObject` 錯誤處理與有界等待（含 pipe drain）、handle 繼承 mutex 範圍不足。
- #3（F-2 Workspace）已透過 PR #19 合併至 `main` 並關閉（root 真實路徑／identity 綁定、junction／絕對路徑／symlink 逃逸攔截與 TC-WS 測試全數通過）。
- #10／#12 已分別透過 PR #16／#17 合併至 `main` 並關閉。
- 完成 Alpha 1 v1.1 規格，狀態為 Review。
- 建立 Git、GitHub 與 Maze 專案治理基礎。
- 依 `docs/spec.md` 需求追蹤矩陣建立 #1～#15、結構化標籤與原生父子關係。

## 未追蹤本機工作

- PR #21（GitHub Actions CI workflow，`.github/workflows/ci.yml`）已合併至 `main`，無對應 Issue；每次 push／PR 自動跑 `go build`／`go vet`／`go test`／零 CGO Windows build。

## 已知驗證限制

- AC-5 的 stale-read 防護屬 #5（F-4），尚未實作（先前文件誤標為 #4，#4 實為 F-3 工具實作，已一併更正）。
- `internal/exec` 的 Timeout／MaxProcesses／MaxMemoryBytes／MaxOutputBytes 一律由呼叫端明確提供，套件本身不內建預設值；benchmark 模式的實際硬性數字仍受 OQ-4 未裁決阻塞。

# Brunel — 當前狀態

> 最後同步：2026-07-14
> Branch：maze/2026-07-14-ec533b
> Working tree：Issue #6（F-5 PowerShell Job Object 執行器）實作於功能分支，待 review／merge，尚未開 PR

## 進行中 Issues

- [#1 Alpha 1：薄型 coding harness 實作追蹤](https://github.com/bext1998/brunel/issues/1)，含 #2～#15 的 14 個原生子項。
- #2～#15 對應 `TASK-A1-F01`～`TASK-A1-F14`；均標記為 `priority: P1` 並指派給 `bext1998`。

## 阻塞 Issues

- `docs/spec.md` §15.8 列出的 Open Questions 尚待使用者裁決。
- #15（Smoke Benchmark Runner）受 OQ-4 的硬性預算裁決阻塞。

## 等待 Review

- 無（Issue #6 尚在本機開發中，未開 PR）。

## 等待 Merge

- 無。

## 已合併待關閉

- 無。

## 最近完成

- #3（F-2 Workspace）已透過 PR #19 合併至 `main`（root 真實路徑／identity 綁定、junction／絕對路徑／symlink 逃逸攔截與 TC-WS 測試全數通過）。
- #10／#12 已分別透過 PR #16／#17 合併至 `main`。
- 完成 Alpha 1 v1.1 規格，狀態為 Review。
- 建立 Git、GitHub 與 Maze 專案治理基礎。
- 依 `docs/spec.md` 需求追蹤矩陣建立 #1～#15、結構化標籤與原生父子關係。

## 分支內工作

- `internal/exec/`（Issue #6 F-5 實作：Job Object 綁定、`CREATE_SUSPENDED → AssignProcessToJobObject → ResumeThread` 順序、timeout／取消分流、輸出擷取上限；本機完整測試、`go vet` 與 Windows 零 CGO build 通過，尚未開 PR）。

## 已知驗證限制

- AC-5 的 stale-read 防護屬 #4，尚未實作。
- `internal/exec` 的 Timeout／MaxProcesses／MaxMemoryBytes／MaxOutputBytes 一律由呼叫端明確提供，套件本身不內建預設值；benchmark 模式的實際硬性數字仍受 OQ-4 未裁決阻塞。

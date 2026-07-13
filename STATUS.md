# Brunel — 當前狀態

> 最後同步：2026-07-13
> Branch：agent/config-issue-12
> Working tree：保留既有 STATUS／NEXT_ACTION 修改與未追蹤規格文件

## 進行中 Issues

- [#1 Alpha 1：薄型 coding harness 實作追蹤](https://github.com/bext1998/brunel/issues/1)，含 #2～#15 的 14 個原生子項。
- #2～#15 對應 `TASK-A1-F01`～`TASK-A1-F14`；均標記為 `priority: P1` 並指派給 `bext1998`。

## 阻塞 Issues

- #2（CLI）等待 #10 與 #12 的 PR review／合併後才能開始實作。
- 規格引用的 `Brunel_產品提案.md` 與 `ADR-001` 雖出現在未追蹤工作區，尚未納入 repository。
- `docs/spec.md` §15.8 列出的 Open Questions 尚待使用者裁決。
- #15（Smoke Benchmark Runner）受 OQ-4 的硬性預算裁決阻塞。

## 等待 Review

- [#16 Implement Session persistence and resume](https://github.com/bext1998/brunel/pull/16)：draft，關聯 #10；本機測試、vet 與 Windows 零 CGO build 通過，GitHub 無適用 checks。
- [#17 Implement layered config and credential separation](https://github.com/bext1998/brunel/pull/17)：draft，關聯 #12；本機測試、vet 與 Windows 零 CGO build 通過，GitHub 無適用 checks。

## 等待 Merge

- 無。

## 已合併待關閉

- 無。

## 最近完成

- 完成 Alpha 1 v1.1 規格，狀態為 Review。
- 建立 Git、GitHub 與 Maze 專案治理基礎。
- 依 `docs/spec.md` 需求追蹤矩陣建立 #1～#15、結構化標籤與原生父子關係。

## 未追蹤本機工作

- `docs/Brunel_產品提案.md`、`docs/adr/`、`docs/open-questions.md`（既有未追蹤項目）。

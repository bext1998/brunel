# Brunel — Coding Agent 指令

> 本文件供 Codex、Claude Code 等 coding agent 在每個 session 開始時閱讀。

## 專案概述

Brunel 是一個面向 Windows x64 的薄型 coding harness 實驗，重點是工具、安全邊界、透明度與完成證據。

技術棧：Go 1.24.x、零 CGO、PowerShell 7、Windows x64、本機檔案與 Windows Credential Manager。

## 工作原則

1. 先閱讀 `MAZE_PROJECT.md` 取得規格與關鍵文件的實際路徑。
2. 只實作任務要求的功能，不添加額外功能或任務外重構。
3. 修改前閱讀相關實作、型別、測試、文件、呼叫者與資料流。
4. 每次 session 結束前同步 `STATUS.md` 與 `NEXT_ACTION.md`。
5. Git commit 或 push 前遵循 `maze-github-safe-ops` 的 pre-commit 與 pre-push 檢查清單。
6. 後續變更使用功能分支與 Pull Request；不得直接推送 `main`。

## 當前狀態與下一步

- 閱讀 `STATUS.md` 了解當前開發狀態。
- 閱讀 `NEXT_ACTION.md` 了解下一個 session 的目標。

## 重要文件

| 文件 | 用途 |
|---|---|
| `docs/spec.md` | Alpha 1 功能規格、凍結契約與驗收標準 |
| `MAZE_PROJECT.md` | 專案定位、實際文件路徑與 GitHub 工作流 |
| `PROJECT_BRIEF.md` | 專案目的、技術棧與重要限制 |
| `STATUS.md` | 當前狀態與阻塞 |
| `NEXT_ACTION.md` | 下一步行動 |
| `DECISIONS.md` | 已確認的專案決策 |

## 禁止行為

- 不得 force push 到 `main` 或 `master`。
- 未經使用者明確要求，不得 commit、push、merge、發布或部署。
- 不得自行修改 `docs/spec.md` 的功能範圍或任何 `[FROZEN]` 契約。
- 不得將 token、API key、密碼或其他憑證寫入 repository。
- 不得把待確認假設或 Open Questions 當作已裁決決策。

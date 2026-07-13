# Brunel

Brunel 是一個面向 Windows x64 的實驗性 coding harness，用來驗證：當模型具備足夠的自主推理與工具使用能力時，harness 是否能聚焦於工具、邊界、透明度與完成證據，而不需要強制固定工作流。

## 專案狀態

目前處於 Alpha 1 初期實作階段。正式需求、架構不變式、凍結介面與驗收條件請參閱 [`docs/spec.md`](docs/spec.md)。

## 技術環境

- Go 1.22（目前實作基線；Alpha 1 v1.2 目標基線為 Go 1.25.x，TUI 實作前需同步）
- Windows x64
- PowerShell 7 (`pwsh`)
- `CGO_ENABLED=0` 靜態編譯

## 開發指引

- Coding Agent 先閱讀 `AGENTS.md`、`MAZE_PROJECT.md`、`STATUS.md` 與 `NEXT_ACTION.md`。
- 不得自行修改規格中的 `[FROZEN]` 契約；變更須走規格修訂與使用者裁決。
- 後續變更使用功能分支與 Pull Request，不直接推送 `main`。

## Session 資料安全

Session 會以未加密檔案保存在本機。Brunel 會在寫入前遮罩已知的 API key、Authorization header 與 `.env` 憑證模式，但這僅是 best-effort，無法保證辨識所有敏感內容；API key 與 token 不得寫入 Session，正式憑證只由 Windows Credential Manager 提供。

## 授權

本專案採用 [Apache License 2.0](LICENSE)。

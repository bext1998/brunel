# Brunel

Brunel 是一個面向 Windows x64 的實驗性 coding harness，用來驗證：當模型具備足夠的自主推理與工具使用能力時，harness 是否能聚焦於工具、邊界、透明度與完成證據，而不需要強制固定工作流。

## 專案狀態

目前處於 Alpha 1 規格審查階段，尚未開始程式碼實作。正式需求、架構不變式、凍結介面與驗收條件請參閱 [`docs/spec.md`](docs/spec.md)。

## 技術環境

- Go 1.24.x
- Windows x64
- PowerShell 7 (`pwsh`)
- `CGO_ENABLED=0` 靜態編譯

## 開發指引

- Coding Agent 先閱讀 `AGENTS.md`、`MAZE_PROJECT.md`、`STATUS.md` 與 `NEXT_ACTION.md`。
- 不得自行修改規格中的 `[FROZEN]` 契約；變更須走規格修訂與使用者裁決。
- 後續變更使用功能分支與 Pull Request，不直接推送 `main`。

## 授權

本專案採用 [Apache License 2.0](LICENSE)。

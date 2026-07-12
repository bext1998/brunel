# Brunel — 專案說明

> 建立日期：2026-07-13
> 最後更新：2026-07-13

## 一句話說明

Brunel 是一個薄型 coding harness，用可量測的 Alpha 1 實驗驗證模型是否能在明確工具、安全邊界、透明度與完成證據下自主完成程式工作。

## 核心問題

現有 coding harness 常以固定工作流、強制規劃與大型系統提示替模型決策。Brunel 要驗證：能力足夠的模型是否能在較薄的 harness 中維持可靠、安全且可驗收的工作閉環。

## 技術棧

- **語言**：Go 1.24.x（零 CGO、靜態編譯）
- **框架 / 主要套件**：Go CLI；具體依賴依規格與後續裁決
- **資料存儲**：本機檔案系統、append-only JSONL session event log、Windows Credential Manager（僅憑證）
- **目標平台**：Windows x64、PowerShell 7 (`pwsh`)

## Coding Agent 工具

- **主要工具**：Codex
- **備用工具**：Claude Code

## 相關文件

- 規格書：docs/spec.md
- 當前狀態：STATUS.md
- 下一步：NEXT_ACTION.md
- 決策紀錄：DECISIONS.md

## 重要限制

- 不得自行修改 `docs/spec.md` 中的 `[FROZEN]` 介面、資料結構、模組邊界與安全不變式。
- Alpha 1 僅支援 Windows x64 與 PowerShell 7，不實作 MCP、plugin system、subagent、LSP 或跨平台支援。
- 專案設定不得保存憑證或降低硬性安全政策。
- 規格中的 Open Questions 與待確認假設在使用者裁決前不得視為正式決策。

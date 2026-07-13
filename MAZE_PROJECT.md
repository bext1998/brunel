# MAZE_PROJECT — Brunel 定位與工作流設定

> 由 `maze-project-init` 建立。Agent 讀取規格前必須先由此取得實際路徑。
> 文件搬移或設定變更時才更新；不得記錄 token、API key、密碼或私密憑證。

## 專案資訊

- 專案名稱：Brunel
- 目標工具：Codex、Claude Code
- 建立日期：2026-07-13

## 文件

- Spec：docs/spec.md
- Product Brief：PROJECT_BRIEF.md
- Product Proposal：docs/Brunel_產品提案.md
- Runtime ADR：docs/adr/ADR-001-runtime-language.md
- Status：STATUS.md
- Next Action：NEXT_ACTION.md
- Decisions：DECISIONS.md

## 自適應 Guidance

- Default profile：standard
- Model overlay：none
- Host capabilities：依執行中的 Codex 或 Claude Code 環境；不假設可用 Subagent、平行工具或額外 Context。
- Profile escalation evidence：僅在發生具體失敗時記錄。

## GitHub

- Repository：https://github.com/bext1998/brunel
- Issue tracking：enabled
- Spec to Issues：enabled（本次初始化不建立 Issues）
- Priority label convention：`priority: P1`
- Category label convention：`type: bug`
- Default assignee policy：specified (`bext1998`)
- Allow label creation：yes

## 備注

- `docs/spec.md` 的前置文件與 `docs/adr/ADR-001-runtime-language.md` 已存在；規格、ADR 與實作基線的版本差異須以 `STATUS.md` 與 `NEXT_ACTION.md` 明確記錄。

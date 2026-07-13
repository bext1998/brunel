# ADR-001 執行環境與語言選型

## 狀態

Accepted

## 日期

2026-07-13

## 脈絡

Alpha 1 有兩項硬需求 ——
(a) 乾淨 Windows x64 環境下載單一 brunel.exe 即可執行，不需預裝任何 runtime；
(b) 導入 Windows Job Object 以實作程序樹終止與資源上限。

## 評估選項

| 選項 | 單檔 exe | Job Object 取得難度 | AI 代理產出穩定度 | 與既有專案棧一致性 | 依賴負債 |
|---|---|---|---|---|---|
| Go 1.24 | 原生靜態編譯 | 低，可直接使用 Win32 API binding | 高 | 高 | 低 |
| TypeScript + Node 22 | SEA / bun compile 有限制 | 高，需 N-API native addon | 高 | 高 | 高 |
| Rust | 原生靜態編譯 | 低，可直接使用 Windows API crate | 較低 | 低 | 中 |

## 決策

Go 1.24.x，golang.org/x/sys/windows，CGO_ENABLED=0，靜態編譯，GOOS=windows GOARCH=amd64。

## 理由

1. Job Object 是 Alpha 1 硬需求。Node 需透過 N-API native addon 才能取用 Win32 API，這與單檔 exe（SEA / bun compile）在 Windows 上高度衝突。
2. TS 原本唯一的優勢（tokenizer / SDK 生態）在 Brunel 的設計下不成立：provider 只接 OpenRouter（純 HTTP + SSE），且使用者可自選任意模型，本地精確 tokenize 本來就不可能，只能吃 provider 回報的 usage。
3. Brunel 的產品命題就是「不要框架」，引入 agent loop 套件是自我否定。
4. Go 靜態編譯單檔零 runtime 依賴，直接滿足需求 (a)。

## 承擔的代價

SSE streaming 解析、agent loop、tool schema 產生器需自行實作。

## 推翻條件

若未來需要 in-process 執行 TS Language Service 等 JS 生態工具則需重新評估；但 spec §5.4 已明確排除 LSP。

## 被否決的選項與原因

- TypeScript / Node 22：native addon 與單檔 exe 衝突（見理由 1）。
- Rust：單檔與程序控制不輸 Go，但 AI 代理產出穩定度較低，且會成為唯一的 Rust 專案，跨專案慣例無法共用。

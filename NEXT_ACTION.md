# Brunel — 下一步行動

> 最後同步：2026-07-13

## 下一個 Session 目標

#3 的 symlink escape 測試已於 Developer Mode 環境驗證通過，AC-6 全數完成；開始 #4 stale-read 防護。

## 優先行動

1. `go test -v ./internal/workspace` 已在 Developer Mode 環境完整執行，8 個測試（含 symlink escape）全數通過，AC-6 可宣稱完成。
2. 另開工作處理 #4（F-4），以 workspace guard 為前置條件實作 stale-read 與 AC-5。
3. Review #3 的功能分支；AC-6 驗證限制已解除，依 review 結果判定合併並關閉 Issue。

## 阻塞與待決策

- AC-5 受 #4 未實作阻塞。

## 參考

- `docs/spec.md` §0、§15.4、§15.8、§16
- `MAZE_PROJECT.md`

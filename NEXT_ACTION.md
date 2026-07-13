# Brunel — 下一步行動

> 最後同步：2026-07-13

## 下一個 Session 目標

在具備 Windows symlink 建立權限的環境補跑 #3 的 symlink escape 測試，並開始 #4 stale-read 防護。

## 優先行動

1. 在啟用 Developer Mode 或具備 symlink privilege 的 Windows runner 執行 `go test -v ./internal/workspace`；通過前不得宣稱 AC-6 完成。
2. 另開工作處理 #4（F-4），以 workspace guard 為前置條件實作 stale-read 與 AC-5。
3. Review #3 的功能分支與驗證限制；symlink 測試補齊後再判定是否可合併及關閉 Issue。

## 阻塞與待決策

- #3 的 symlink escape 驗證受目前 Windows 帳號權限阻塞；junction／絕對路徑案例不代表 AC-6 全數通過。
- AC-5 受 #4 未實作阻塞。

## 參考

- `docs/spec.md` §0、§15.4、§15.8、§16
- `MAZE_PROJECT.md`

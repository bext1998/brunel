# Brunel — 下一步行動

> 最後同步：2026-07-13

## 下一個 Session 目標

完成 #10／#12 的 review 與合併，使 #2 CLI 具備可整合的 Session／Config 依賴。

## 優先行動

1. Review draft PR #16（Session）與 #17（Config），確認後轉為 ready 並依序合併。
2. 兩個依賴合併後更新本機 `main`，再從新功能分支實作 #2 的 CLI／TTY 契約。
3. 另行處理未追蹤規格文件與 §15.8 Open Questions；不得把待確認內容當成已裁決決策。

## 阻塞與待決策

- #2 受 #10、#12 尚未合併阻塞；本 session 未實作 #2。
- 前置文件缺漏：`Brunel_產品提案.md`、`ADR-001`。
- Open Questions 尚未裁決；受影響的發布門檻不得宣稱通過。

## 參考

- `docs/spec.md` §0、§15.4、§15.8、§16
- `MAZE_PROJECT.md`

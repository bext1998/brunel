# Brunel — 下一步行動

> 最後同步：2026-07-14

## 下一個 Session 目標

開始 #5（F-4 stale-read）：實作 expected_hash 前置條件與原子寫入，對應 v1.2 AC-7。

## 優先行動

1. 實作 #5：read_file 回傳全檔 SHA-256；apply_patch／write_file 驗證 expected_hash，失敗保留原檔。
2. 依 #5 的 AC 補 stale、patch conflict、取消與原子寫入回歸測試。
3. 之後處理 #7 的 AUTO／CONFIRM 安全入口，再解除 #4 與 #2 的相依阻塞。
4. 規劃 #2 前將 Go module 基線由 1.22 同步至 1.25.x，並引入 Bubble Tea v2；此項需另行實作授權。

## 阻塞與待決策

- 無 Alpha 1 硬阻塞；`docs/spec.md` §16 的 Open Questions 依各自裁決前行為處理。

## 參考

- `docs/spec.md` §4～§6、§8～§16
- `MAZE_PROJECT.md`

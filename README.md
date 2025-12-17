## 作业：Go 实现切片的删除操作

本项目演示并实现：
- 删除切片特定下标元素
- 高性能实现（稳定/不稳定删除）
- 泛型方法
- 支持缩容，并提供缩容机制

### 目录结构
- `sliceutil/delete.go`：实现代码
- `sliceutil/delete_test.go`：单元测试
- `main.go`：简单运行示例

### 主要 API
- `sliceutil.DeleteAtStable[T any](s []T, idx int, shrinkPolicy sliceutil.ShrinkPolicy) []T`
  - 稳定删除（保持顺序），内部使用 `copy` 原地搬移
- `sliceutil.DeleteAtUnstable[T any](s []T, idx int, shrinkPolicy sliceutil.ShrinkPolicy) []T`
  - 不稳定删除（不保证顺序），O(1) 删除
- `sliceutil.ShrinkPolicy`
  - 删除后可选缩容策略（阈值触发，按比例收缩）

（另外提供与参考答案风格一致的：`sliceutil.DeleteAt[T any](src []T, index int) ([]T, error)` 与 `sliceutil.Shrink[T any](src []T) []T`。）

### 运行
在项目根目录执行：

```bash
go test ./...
go run .
```

package sliceutil

import (
	"errors"
	"fmt"
)

// ErrIndexOutOfRange 老师参考实现里的越界错误（便于对齐作业要求/判题）。
var ErrIndexOutOfRange = errors.New("下标超出范围")

// ShrinkPolicy 控制删除后是否缩容以及如何缩容。
//
// 设计目标：
// - 删除通常是高频操作，默认不缩容（避免频繁分配）。
// - 当 cap 远大于 len 时再缩容，回收内存。
//
// 触发条件：cap(s) >= MinCap && cap(s) >= len(s)*Factor
// 缩容目标：newCap = max(len, cap/ToDivisor)
// 常用：Factor=4、ToDivisor=2（cap 至少是 len 的 4 倍才缩到一半）。
// 也可禁用：Enabled=false。
//
// 注意：缩容会分配新底层数组。
// 若 T 包含指针，为了帮助 GC 回收被删除元素引用，删除时会对尾部进行零值清理。
// 若希望更高性能且不关心 GC，可自行移除清理代码。
//
// Go 版本：泛型需要 Go 1.18+。
type ShrinkPolicy struct {
	Enabled   bool
	Factor    int // 触发阈值倍数，例如 4
	ToDivisor int // 缩容比例，例如 2 表示 cap/2
	MinCap    int // cap 小于该值时不缩容（避免小切片抖动）
}

// DeleteAt 删除指定位置的元素（稳定删除），越界返回 error（与老师参考答案风格一致）。
//
// 性能：内部使用 copy 搬移（通常比手写 for 更快/更简洁）。
// GC：会将尾部元素置零，避免悬挂引用。
func DeleteAt[T any](src []T, index int) ([]T, error) {
	length := len(src)
	if index < 0 || index >= length {
		return nil, fmt.Errorf("ekit: %w, 下标超出范围，长度 %d, 下标 %d",
			ErrIndexOutOfRange, length, index)
	}
	out := DeleteAtStable(src, index, ShrinkPolicy{Enabled: false})
	return out, nil
}

func (p ShrinkPolicy) normalize() ShrinkPolicy {
	if !p.Enabled {
		return p
	}
	if p.Factor <= 0 {
		p.Factor = 4
	}
	if p.ToDivisor <= 1 {
		p.ToDivisor = 2
	}
	if p.MinCap < 0 {
		p.MinCap = 0
	}
	return p
}

func (p ShrinkPolicy) shouldShrink(length, capacity int) bool {
	p = p.normalize()
	if !p.Enabled {
		return false
	}
	if capacity < p.MinCap {
		return false
	}
	// length==0 时：如果 capacity>=MinCap 也可缩到 0（返回空切片）
	return capacity >= length*p.Factor
}

func (p ShrinkPolicy) shrinkCap(length, capacity int) int {
	p = p.normalize()
	if length <= 0 {
		return 0
	}
	target := capacity / p.ToDivisor
	if target < length {
		return length
	}
	return target
}

// Shrink 这是老师参考答案同款缩容实现（独立缩容函数）。
//
// 规则：
// - cap<=64：不缩容
// - cap>2048 且 cap/len>=2：缩到 cap*0.625（5/8）
// - cap<=2048 且 cap/len>=4：缩到 cap/2
func Shrink[T any](src []T) []T {
	c, l := cap(src), len(src)
	n, changed := calCapacity(c, l)
	if !changed {
		return src
	}
	s := make([]T, 0, n)
	s = append(s, src...)
	return s
}

func calCapacity(c, l int) (int, bool) {
	if c <= 64 {
		return c, false
	}
	// len==0 时可以直接缩到 0（返回空切片）
	if l == 0 {
		return 0, true
	}
	if c > 2048 && (c/l >= 2) {
		factor := 0.625
		return int(float32(c) * float32(factor)), true
	}
	if c <= 2048 && (c/l >= 4) {
		return c / 2, true
	}
	return c, false
}

// DeleteAtStable 删除 s 中下标 idx 的元素，并保持元素相对顺序（稳定删除）。
//
// - **高性能点**：原地移动（copy），不额外分配；时间复杂度 O(n-idx)。
// - **安全性**：idx 越界会 panic，符合 Go 习惯（也便于测试暴露问题）。
// - **GC 友好**：将尾部一个元素置零，避免悬挂引用。
// - **缩容**：当 shrinkPolicy 触发时会重新分配并拷贝到新底层数组。
func DeleteAtStable[T any](s []T, idx int, shrinkPolicy ShrinkPolicy) []T {
	_ = s[idx] // bounds check
	copy(s[idx:], s[idx+1:])
	var zero T
	s[len(s)-1] = zero
	s = s[:len(s)-1]
	return maybeShrink(s, shrinkPolicy)
}

// DeleteAtUnstable 删除 s 中下标 idx 的元素，但不保证顺序（不稳定删除）。
//
// - **高性能点**：O(1) 交换到末尾再截断，适合不关心顺序的大切片。
// - **GC 友好**：同样清理尾部引用。
// - **缩容**：同 DeleteAtStable。
func DeleteAtUnstable[T any](s []T, idx int, shrinkPolicy ShrinkPolicy) []T {
	_ = s[idx] // bounds check
	last := len(s) - 1
	s[idx] = s[last]
	var zero T
	s[last] = zero
	s = s[:last]
	return maybeShrink(s, shrinkPolicy)
}

// DeleteRangeStable 删除区间 [from, to) 并保持顺序。
// from==to 时不变。
func DeleteRangeStable[T any](s []T, from, to int, shrinkPolicy ShrinkPolicy) []T {
	if from < 0 || to < from || to > len(s) {
		panic("sliceutil: invalid range")
	}
	if from == to {
		return maybeShrink(s, shrinkPolicy)
	}
	n := copy(s[from:], s[to:])
	// 清理尾部被“挤出”的元素，帮助 GC
	var zero T
	for i := len(s) - (to - from); i < len(s); i++ {
		s[i] = zero
	}
	s = s[:len(s)-(to-from)]
	_ = n
	return maybeShrink(s, shrinkPolicy)
}

func maybeShrink[T any](s []T, p ShrinkPolicy) []T {
	if !p.shouldShrink(len(s), cap(s)) {
		return s
	}
	newCap := p.shrinkCap(len(s), cap(s))
	if newCap == cap(s) {
		return s
	}
	if len(s) == 0 {
		return nil
	}
	out := make([]T, len(s), newCap)
	copy(out, s)
	return out
}

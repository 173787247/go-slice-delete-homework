package main

import (
	"fmt"

	"go-slice-delete-homework/sliceutil"
)

func main() {
	s := []string{"a", "b", "c", "d"}
	fmt.Println("before:", s, "len", len(s), "cap", cap(s))

	// 稳定删除（保持顺序）
	s = sliceutil.DeleteAtStable(s, 1, sliceutil.ShrinkPolicy{Enabled: false})
	fmt.Println("stable delete idx=1:", s)

	// 不稳定删除（更快，但会打乱顺序）
	s = sliceutil.DeleteAtUnstable(s, 1, sliceutil.ShrinkPolicy{Enabled: false})
	fmt.Println("unstable delete idx=1:", s)

	// 带缩容策略（cap 至少是 len 的 4 倍且 cap>=16 时触发；缩到 cap/2）
	big := make([]int, 0, 64)
	big = append(big, 1, 2, 3, 4)
	fmt.Println("big before:", big, "len", len(big), "cap", cap(big))
	big = sliceutil.DeleteAtStable(big, 0, sliceutil.ShrinkPolicy{Enabled: true, Factor: 4, ToDivisor: 2, MinCap: 16})
	fmt.Println("big after:", big, "len", len(big), "cap", cap(big))
}

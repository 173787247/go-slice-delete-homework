package sliceutil

import (
	"errors"
	"testing"
)

type ptrWrap struct{ P *int }

func TestDeleteAtStable_Int(t *testing.T) {
	s := []int{1, 2, 3, 4}
	s = DeleteAtStable(s, 1, ShrinkPolicy{Enabled: false})
	if got, want := len(s), 3; got != want {
		t.Fatalf("len=%d want=%d", got, want)
	}
	if got := (s); !(got[0] == 1 && got[1] == 3 && got[2] == 4) {
		t.Fatalf("got=%v", got)
	}
}

func TestDeleteAtUnstable_AllowsReorder(t *testing.T) {
	s := []int{1, 2, 3, 4}
	s = DeleteAtUnstable(s, 1, ShrinkPolicy{Enabled: false})
	if got, want := len(s), 3; got != want {
		t.Fatalf("len=%d want=%d", got, want)
	}
	// 不稳定删除：s[1] 变成原最后一个 4
	if got := s; !(got[0] == 1 && got[1] == 4) {
		t.Fatalf("got=%v", got)
	}
}

func TestDeleteRangeStable(t *testing.T) {
	s := []int{1, 2, 3, 4, 5}
	s = DeleteRangeStable(s, 1, 4, ShrinkPolicy{Enabled: false})
	if got := s; !(len(got) == 2 && got[0] == 1 && got[1] == 5) {
		t.Fatalf("got=%v", got)
	}
}

func TestShrinkPolicy_Triggered(t *testing.T) {
	// 构造一个 len 很小、cap 很大的切片
	s := make([]int, 0, 64)
	s = append(s, 1, 2, 3, 4)
	oldCap := cap(s)

	p := ShrinkPolicy{Enabled: true, Factor: 4, ToDivisor: 2, MinCap: 16}
	s = DeleteAtStable(s, 0, p) // len:4->3, cap:64 应触发 shrink

	if cap(s) >= oldCap {
		t.Fatalf("expected shrink: cap=%d old=%d", cap(s), oldCap)
	}
	if len(s) != 3 {
		t.Fatalf("len=%d", len(s))
	}
}

func TestGCFriendly_ZeroTail(t *testing.T) {
	// 验证删除后尾部被置零（帮助 GC）
	a, b, c := 1, 2, 3
	s := []ptrWrap{{&a}, {&b}, {&c}}
	s2 := DeleteAtStable(s, 1, ShrinkPolicy{Enabled: false})
	if len(s2) != 2 {
		t.Fatalf("len=%d", len(s2))
	}
	// s 是同一个底层数组，删除后原 len-1 位置应为零值
	if s[2].P != nil {
		t.Fatalf("tail not zeroed")
	}
}

func TestDeleteAt_ErrorOnOutOfRange(t *testing.T) {
	_, err := DeleteAt([]int{1, 2, 3}, 3)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, ErrIndexOutOfRange) {
		t.Fatalf("expected ErrIndexOutOfRange, got %v", err)
	}
}

func TestShrink_GeekTimeRules(t *testing.T) {
	// cap<=64 不缩
	a := make([]int, 0, 64)
	a = append(a, 1, 2, 3)
	a2 := Shrink(a)
	if cap(a2) != cap(a) {
		t.Fatalf("unexpected shrink for cap<=64: %d -> %d", cap(a), cap(a2))
	}

	// cap<=2048 且 cap/len>=4 缩到一半
	b := make([]int, 0, 128)
	b = append(b, 1, 2, 3) // 128/3 >= 4
	b2 := Shrink(b)
	if cap(b2) != 64 {
		t.Fatalf("expected cap 64, got %d", cap(b2))
	}

	// cap>2048 且 cap/len>=2 缩到 5/8
	c := make([]int, 0, 4096)
	for i := 0; i < 2000; i++ {
		c = append(c, i)
	}
	c2 := Shrink(c)
	// 4096 * 0.625 = 2560
	if cap(c2) != 2560 {
		t.Fatalf("expected cap 2560, got %d", cap(c2))
	}
}

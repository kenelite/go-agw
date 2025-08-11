package scheduler

import "testing"

func TestRoundRobinNext(t *testing.T) {
    rr := NewRoundRobin()
    got := []int{rr.Next(3), rr.Next(3), rr.Next(3), rr.Next(3), rr.Next(3), rr.Next(3)}
    want := []int{1, 2, 0, 1, 2, 0}
    for i := range want {
        if got[i] != want[i] {
            t.Fatalf("round-robin unexpected at %d: got=%d want=%d", i, got[i], want[i])
        }
    }
    if rr.Next(0) != -1 {
        t.Fatalf("expected -1 when no candidates")
    }
}


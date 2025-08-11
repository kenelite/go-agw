package scheduler

import (
    "sync/atomic"
)

// Scheduler selects next index given N backends.
type Scheduler interface {
    Next(numCandidates int) int
}

type RoundRobin struct { counter atomic.Int64 }

func NewRoundRobin() *RoundRobin { return &RoundRobin{} }

func (r *RoundRobin) Next(n int) int {
    if n <= 0 { return -1 }
    v := r.counter.Add(1)
    idx := int(v % int64(n))
    if idx < 0 { idx = -idx }
    return idx
}


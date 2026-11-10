package scanner

import "sync"

// taskStack is an unbounded LIFO of folders to list. LIFO keeps the walk
// depth-first-ish, so the set of in-flight folders (and memory) stays small.
type taskStack struct {
	mu     sync.Mutex
	cond   *sync.Cond
	items  []*dir
	closed bool
}

func newTaskStack() *taskStack {
	t := &taskStack{}
	t.cond = sync.NewCond(&t.mu)
	return t
}

func (t *taskStack) push(d *dir) {
	t.mu.Lock()
	t.items = append(t.items, d)
	t.mu.Unlock()
	t.cond.Signal()
}

func (t *taskStack) pop() (*dir, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	for len(t.items) == 0 && !t.closed {
		t.cond.Wait()
	}
	if t.closed {
		return nil, false
	}
	last := len(t.items) - 1
	d := t.items[last]
	t.items[last] = nil
	t.items = t.items[:last]
	return d, true
}

func (t *taskStack) close() {
	t.mu.Lock()
	t.closed = true
	t.mu.Unlock()
	t.cond.Broadcast()
}

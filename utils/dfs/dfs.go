package dfs

import "math"

type DirectedCycle struct {
	marked  map[string]bool
	onStack map[string]bool
	edgeTo  map[string]string
	cycle   []string
}

func NewDirectedCycle(m map[string][]string) *DirectedCycle {
	if m == nil {
		return nil
	}
	dc := &DirectedCycle{
		marked:  make(map[string]bool),
		onStack: make(map[string]bool),
		edgeTo:  make(map[string]string),
	}
	for k := range m {
		if ok, _ := dc.marked[k]; !ok {
			dc.dfs(m, k)
		}
	}
	return dc
}

func (dc *DirectedCycle) dfs(m map[string][]string, v string) {
	dc.marked[v] = true
	dc.onStack[v] = true
	for _, w := range m[v] {
		if len(dc.cycle) > 0 {
			return
		} else if ok, _ := dc.marked[w]; !ok {
			dc.edgeTo[w] = v
			dc.dfs(m, w)
		} else if on, _ := dc.onStack[w]; on {
			dc.cycle = make([]string, 0, len(m))
			for x := v; x != w; x = dc.edgeTo[x] {
				dc.cycle = append(dc.cycle, x)
			}
			dc.cycle = append(dc.cycle, w)
			dc.cycle = append(dc.cycle, v)
		}
	}
	dc.onStack[v] = false
}

func (dc *DirectedCycle) Cycle() []string {
	return dc.cycle
}

type BfsData struct {
	marked   map[string]bool
	distTo   map[string]int64
	edgeTo   map[string]string
	bfsOrder []string
}

func NewBFS(m map[string][]string, s string) *BfsData {
	bfs := &BfsData{
		marked:   make(map[string]bool),
		distTo:   make(map[string]int64),
		edgeTo:   make(map[string]string),
		bfsOrder: make([]string, 0, len(m)),
	}
	for v := range m {
		bfs.distTo[v] = math.MaxInt64
	}
	bfs.bfs(m, s)
	return bfs
}
func (b *BfsData) bfs(m map[string][]string, s string) {
	q := []string{s}
	b.marked[s] = true
	b.distTo[s] = 0

	for len(q) > 0 {
		var v string
		v, q = q[0], q[1:]
		b.bfsOrder = append(b.bfsOrder, v)
		for _, w := range m[v] {
			if m, _ := b.marked[w]; m {
				continue
			}
			b.edgeTo[w] = v
			b.distTo[w] = b.distTo[v] + 1
			b.marked[w] = true
			q = append(q, w)
		}
	}
}

func (b *BfsData) BfsOrder() []string {
	return b.bfsOrder
}

type DfsOrder struct {
	marked    map[string]bool
	pre       map[string]int
	post      map[string]int
	preOrder  []string
	postOrder []string
	preCount  int
	postCount int
}

func NewDfsOrder(m map[string][]string) *DfsOrder {
	if m == nil {
		return nil
	}
	do := &DfsOrder{
		marked:    make(map[string]bool),
		pre:       make(map[string]int),
		post:      make(map[string]int),
		preOrder:  []string{},
		postOrder: []string{},
	}
	for v := range m {
		if ok, _ := do.marked[v]; !ok {
			do.dfs(m, v)
		}
	}
	return do
}

func (do *DfsOrder) dfs(m map[string][]string, v string) {
	do.marked[v] = true
	do.pre[v] = do.preCount
	do.preCount++
	do.preOrder = append(do.preOrder, v)
	for _, w := range m[v] {
		if ok, _ := do.marked[w]; !ok {
			do.dfs(m, w)
		}
	}
	do.postOrder = append(do.postOrder, v)
	do.post[v] = do.postCount
	do.postCount++
}

func (do *DfsOrder) PostOrder() []string {
	return do.postOrder
}

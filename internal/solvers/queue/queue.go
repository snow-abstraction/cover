/*
 Copyright (C) 2024 Douglas Wayne Potter

 This program is free software: you can redistribute it and/or modify
 it under the terms of the GNU Affero General Public License as
 published by the Free Software Foundation, either version 3 of the
 License, or (at your option) any later version.

 This program is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 GNU Affero General Public License for more details.

 You should have received a copy of the GNU Affero General Public License
 along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package queue

import (
	"container/heap"

	"github.com/snow-abstraction/cover/internal/tree"
)

// LowerBoundPriorityQueue is a priority queue of nodes where
// nodes with lower lower bound are prioritized (i.e Pop'ed first).
type LowerBoundPriorityQueue struct {
	q pq
}

func MakeQueue() LowerBoundPriorityQueue {
	storage := make(pq, 0)
	return LowerBoundPriorityQueue{storage}
}
func (q *LowerBoundPriorityQueue) Push(node *tree.Node) {
	heap.Push(&q.q, &item{node: node})
}
func (q *LowerBoundPriorityQueue) Pop() *tree.Node {
	return heap.Pop(&q.q).(*item).node
}
func (q *LowerBoundPriorityQueue) Len() int {
	return q.q.Len()
}

// An item is a node with its heap index.
// Adapting from PriorityQueue example from https://pkg.go.dev/container/heap
type item struct {
	node  *tree.Node
	index int
}

// A pq (priority queue) implements heap.Interface. It is not intended to be used directly.
// Use LowerBoundPriorityQueue instead.
type pq []*item

func (q pq) Len() int { return len(q) }
func (q pq) Less(i, j int) bool {
	return q[i].node.LowerBound < q[j].node.LowerBound
}
func (q pq) Swap(i, j int) {
	q[i], q[j] = q[j], q[i]
	q[i].index = i
	q[j].index = j
}
func (pq *pq) Push(x any) {
	n := len(*pq)
	item := x.(*item)
	item.index = n
	*pq = append(*pq, item)
}
func (pq *pq) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // don't stop the GC from reclaiming the item eventually
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// Not needed yet.
// func (pq *pq) Update(item *item, node *tree.Node) {
// 	item.node = node
// 	heap.Fix(pq, item.index)
// }

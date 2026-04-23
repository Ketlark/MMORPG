package game

import (
	"container/heap"
	"math"

	pb "mmorpg/server/api/proto/gen"
)

var directions = [8][2]int32{
	{0, -1},  // N
	{1, 0},   // E
	{0, 1},   // S
	{-1, 0},  // W
	{1, -1},  // NE
	{1, 1},   // SE
	{-1, 1},  // SW
	{-1, -1}, // NW
}

type node struct {
	x, y   int32
	g      float64
	f      float64
	parent *node
	index  int
}

type openSet []*node

func (o openSet) Len() int           { return len(o) }
func (o openSet) Less(i, j int) bool { return o[i].f < o[j].f }
func (o openSet) Swap(i, j int) {
	o[i], o[j] = o[j], o[i]
	o[i].index = i
	o[j].index = j
}
func (o *openSet) Push(x interface{}) {
	n := x.(*node)
	n.index = len(*o)
	*o = append(*o, n)
}
func (o *openSet) Pop() interface{} {
	old := *o
	n := old[len(old)-1]
	old[len(old)-1] = nil
	n.index = -1
	*o = old[:len(old)-1]
	return n
}

// FindPath computes an A* path from (startX,startY) to (targetX,targetY).
// Returns nil if no path exists.
func FindPath(mapData *pb.MapData, startX, startY, targetX, targetY int32) []*pb.PathNode {
	if startX == targetX && startY == targetY {
		return []*pb.PathNode{{X: startX, Y: startY}}
	}
	if !IsWalkable(mapData, targetX, targetY) {
		return nil
	}

	maxIter := int(mapData.Width)*int(mapData.Height)*4 + 1
	key := func(x, y int32) int64 { return int64(x)<<32 | int64(y) }

	closed := make(map[int64]bool)
	lookup := make(map[int64]*node)

	start := &node{x: startX, y: startY, g: 0, f: octile(startX, startY, targetX, targetY)}
	lookup[key(startX, startY)] = start

	os := &openSet{start}
	heap.Init(os)

	iter := 0
	for os.Len() > 0 {
		iter++
		if iter > maxIter {
			return nil
		}

		current := heap.Pop(os).(*node)
		ck := key(current.x, current.y)
		if closed[ck] {
			continue
		}
		closed[ck] = true

		if current.x == targetX && current.y == targetY {
			return reconstructPath(current)
		}

		for _, dir := range directions {
			dx, dy := dir[0], dir[1]
			nx, ny := current.x+dx, current.y+dy

			if !IsWalkable(mapData, nx, ny) {
				continue
			}
			if closed[key(nx, ny)] {
				continue
			}

			isDiag := dx != 0 && dy != 0
			if isDiag {
				if !IsWalkable(mapData, current.x+dx, current.y) ||
					!IsWalkable(mapData, current.x, current.y+dy) {
					continue
				}
			}

			cost := 1.0
			if isDiag {
				cost = math.Sqrt2
			}
			ng := current.g + cost

			nk := key(nx, ny)
			existing, found := lookup[nk]
			if found && ng >= existing.g {
				continue
			}

			n := &node{
				x:      nx,
				y:      ny,
				g:      ng,
				f:      ng + octile(nx, ny, targetX, targetY),
				parent: current,
			}
			lookup[nk] = n
			heap.Push(os, n)
		}
	}

	return nil
}

func octile(x1, y1, x2, y2 int32) float64 {
	dx := math.Abs(float64(x2 - x1))
	dy := math.Abs(float64(y2 - y1))
	if dx > dy {
		return dx + (math.Sqrt2-1)*dy
	}
	return dy + (math.Sqrt2-1)*dx
}

func reconstructPath(goal *node) []*pb.PathNode {
	var path []*pb.PathNode
	for n := goal; n != nil; n = n.parent {
		path = append(path, &pb.PathNode{X: n.x, Y: n.y})
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}

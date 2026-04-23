package game

import (
	pb "mmorpg/server/api/proto/gen"
)

const (
	costCardinal = 10
	costDiagonal = 14 // ≈ √2 × 10
)

var dirs = [8][2]int32{
	{0, -1}, {1, 0}, {0, 1}, {-1, 0}, // N E S W
	{1, -1}, {1, 1}, {-1, 1}, {-1, -1}, // NE SE SW NW
}

// FindPath computes the shortest path from (sx,sy) to (tx,ty) using A*
// with octile heuristic and anti corner-cutting.
// Returns nil if no path exists.
func FindPath(mapData *pb.MapData, sx, sy, tx, ty int32) []*pb.PathNode {
	w, h := mapData.Width, mapData.Height
	size := int(w * h)
	if size == 0 || !inBounds(tx, ty, w, h) {
		return nil
	}
	if sx == tx && sy == ty {
		return []*pb.PathNode{{X: sx, Y: sy}}
	}

	start := int(sy*w + sx)
	target := int(ty*w + tx)

	// Flat walkability grid — one pass, direct index access.
	walk := make([]bool, size)
	for i, c := range mapData.Cells {
		walk[i] = c.Walkable
	}
	if !walk[target] {
		return nil
	}

	// A* state — flat arrays, zero per-node allocations.
	const inf = int(^uint(0) >> 1)
	g := make([]int, size)
	f := make([]int, size)
	from := make([]int, size)
	closed := make([]bool, size)
	for i := range g {
		g[i] = inf
		f[i] = inf
		from[i] = -1
	}
	g[start] = 0
	f[start] = heuristic(start, tx, ty, w)

	// Binary min-heap (flat []int, ordered by f[]).
	heap := make([]int, 0, 64)
	heap = append(heap, start)

	for len(heap) > 0 {
		cur := heapPop(&heap, f)
		if closed[cur] {
			continue
		}
		closed[cur] = true

		if cur == target {
			return tracePath(from, target, w)
		}

		cx, cy := int32(cur)%w, int32(cur)/w
		for _, d := range dirs {
			dx, dy := d[0], d[1]
			nx, ny := cx+dx, cy+dy
			if !inBounds(nx, ny, w, h) {
				continue
			}

			next := int(ny*w + nx)
			if !walk[next] || closed[next] {
				continue
			}

			// Anti corner-cutting: diagonal requires both cardinals walkable.
			if dx != 0 && dy != 0 {
				if !walk[int(cy*w+cx+dx)] || !walk[int((cy+dy)*w+cx)] {
					continue
				}
			}

			cost := costCardinal
			if dx != 0 && dy != 0 {
				cost = costDiagonal
			}
			ng := g[cur] + cost
			if ng < g[next] {
				g[next] = ng
				f[next] = ng + heuristic(next, tx, ty, w)
				from[next] = cur
				heapPush(&heap, next, f)
			}
		}
	}

	return nil
}

func inBounds(x, y, w, h int32) bool {
	return uint(x) < uint(w) && uint(y) < uint(h)
}

// heuristic returns octile distance as fixed-point integer.
// max(dx,dy)×10 + min(dx,dy)×4
func heuristic(idx int, tx, ty, w int32) int {
	dx := abs32(int32(idx)%w - tx)
	dy := abs32(int32(idx)/w - ty)
	if dx > dy {
		return int(dx)*costCardinal + int(dy)*(costDiagonal-costCardinal)
	}
	return int(dy)*costCardinal + int(dx)*(costDiagonal-costCardinal)
}

func tracePath(from []int, target int, w int32) []*pb.PathNode {
	var path []*pb.PathNode
	for i := target; i != -1; i = from[i] {
		path = append(path, &pb.PathNode{X: int32(i) % w, Y: int32(i) / w})
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}

func heapPush(h *[]int, idx int, f []int) {
	*h = append(*h, idx)
	i := len(*h) - 1
	for i > 0 {
		p := (i - 1) / 2
		if f[(*h)[p]] <= f[(*h)[i]] {
			break
		}
		(*h)[p], (*h)[i] = (*h)[i], (*h)[p]
		i = p
	}
}

func heapPop(h *[]int, f []int) int {
	n := len(*h)
	root := (*h)[0]
	(*h)[0] = (*h)[n-1]
	*h = (*h)[:n-1]
	i := 0
	for {
		s := i
		l, r := 2*i+1, 2*i+2
		if l < len(*h) && f[(*h)[l]] < f[(*h)[s]] {
			s = l
		}
		if r < len(*h) && f[(*h)[r]] < f[(*h)[s]] {
			s = r
		}
		if s == i {
			break
		}
		(*h)[i], (*h)[s] = (*h)[s], (*h)[i]
		i = s
	}
	return root
}

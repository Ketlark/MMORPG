package game

import (
	pb "mmorpg/server/api/proto/gen"
)

const (
	MapWidth  = 20
	MapHeight = 15

	TerrainGrass = 0
	TerrainRoad  = 1
	TerrainWall  = 2
	TerrainWater = 3
)

// terrainLayout is a 20x15 grid defining the village layout.
// Rows are Y (0=top), columns are X (0=left).
// The layout features:
//   - A village centre with buildings (walls) at rows 3-6, cols 8-17
//   - Roads forming paths between buildings and around the village
//   - A pond (water) in the south-east area
//   - Open grassland surrounding everything
var terrainLayout = [MapHeight][MapWidth]int32{
	// Row 0: top border with a road at col 9-10
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	// Row 1: open grass with road continuing
	{0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0},
	// Row 2: road splits, buildings start
	{0, 0, 0, 0, 0, 0, 1, 1, 1, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0},
	// Row 3: top of buildings, road along sides
	{0, 0, 0, 0, 0, 1, 1, 2, 2, 1, 2, 2, 1, 1, 0, 0, 0, 0, 0, 0},
	// Row 4: buildings with doorways
	{0, 0, 0, 0, 0, 1, 2, 2, 2, 1, 2, 2, 2, 1, 0, 0, 0, 0, 0, 0},
	// Row 5: buildings interior, road between them
	{0, 0, 0, 0, 0, 1, 2, 2, 2, 1, 2, 2, 2, 1, 0, 0, 0, 0, 0, 0},
	// Row 6: bottom of buildings
	{0, 0, 0, 0, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 0, 0},
	// Row 7: road going south
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0},
	// Row 8: road, some scattered trees (walls)
	{0, 0, 0, 0, 0, 0, 0, 2, 0, 1, 1, 0, 2, 0, 0, 0, 0, 0, 0, 0},
	// Row 9: road, water starts appearing
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 3, 3, 3, 0, 0, 0},
	// Row 10: road continues, pond expands
	{0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 0, 3, 3, 3, 3, 3, 0, 0},
	// Row 11: road, pond
	{0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 0, 3, 3, 3, 3, 3, 3, 0, 0},
	// Row 12: bottom of road, pond shrinking
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 3, 3, 3, 3, 0, 0, 0},
	// Row 13: open grass
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 3, 3, 0, 0, 0, 0},
	// Row 14: bottom border with road exit
	{0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0},
}

// walkableLookup provides fast walkability check per terrain type.
var walkableLookup = map[int32]bool{
	TerrainGrass: true,
	TerrainRoad:  true,
	TerrainWall:  false,
	TerrainWater: false,
}

// elevationLookup provides the elevation for each terrain type.
var elevationLookup = map[int32]int32{
	TerrainGrass: 0,
	TerrainRoad:  0,
	TerrainWall:  1,
	TerrainWater: -1,
}

// GenerateTestMap creates a MapData proto for the 20x15 village test map.
func GenerateTestMap() *pb.MapData {
	cells := make([]*pb.MapCell, 0, MapWidth*MapHeight)

	for y := int32(0); y < MapHeight; y++ {
		for x := int32(0); x < MapWidth; x++ {
			terrain := terrainLayout[y][x]
			cells = append(cells, &pb.MapCell{
				X:           x,
				Y:           y,
				TerrainType: terrain,
				Walkable:    walkableLookup[terrain],
				Elevation:   elevationLookup[terrain],
			})
		}
	}

	return &pb.MapData{
		Width:  MapWidth,
		Height: MapHeight,
		Cells:  cells,
	}
}

// IsWalkable checks whether the cell at (x, y) is walkable on the given map.
// Returns false if the coordinates are out of bounds.
func IsWalkable(mapData *pb.MapData, x, y int32) bool {
	if x < 0 || x >= mapData.Width || y < 0 || y >= mapData.Height {
		return false
	}

	idx := y*mapData.Width + x
	if int(idx) >= len(mapData.Cells) {
		return false
	}

	return mapData.Cells[idx].Walkable
}

// FindNearestWalkableCell searches outward from (startX, startY) in a spiral
// pattern and returns the first walkable cell. Useful for placing new players.
func FindNearestWalkableCell(mapData *pb.MapData, startX, startY int32) (int32, int32) {
	if IsWalkable(mapData, startX, startY) {
		return startX, startY
	}

	maxDist := mapData.Width
	if mapData.Height > maxDist {
		maxDist = mapData.Height
	}

	for dist := int32(1); dist < maxDist; dist++ {
		for dx := -dist; dx <= dist; dx++ {
			for dy := -dist; dy <= dist; dy++ {
				if abs32(dx) != dist && abs32(dy) != dist {
					continue // only check the perimeter at this distance
				}
				nx, ny := startX+dx, startY+dy
				if IsWalkable(mapData, nx, ny) {
					return nx, ny
				}
			}
		}
	}

	// Fallback: return (0,0) which is grass and always walkable
	return 0, 0
}

func abs32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

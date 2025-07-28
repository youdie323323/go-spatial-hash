package spatial_hash

import (
	"fmt"
	"math/rand/v2"
	"testing"
	"time"
)

type TestNode = Node[int, float64]

// Point is a basic implementation of the Node interface for testing.
type Point struct {
	id int

	x, y       float64
	oldX, oldY float64
}

var _ TestNode = (*Point)(nil) // *Point must implement TestNode

func (n *Point) GetId() int { return n.id }

func (n *Point) GetX() float64 { return n.x }
func (n *Point) GetY() float64 { return n.y }

func (n *Point) GetOldPos() (float64, float64) { return n.oldX, n.oldY }
func (n *Point) SetOldPos(x, y float64) {
	n.oldX = x
	n.oldY = y
}

// CreateTestNodes creates a slice of test nodes with random positions.
func CreateTestNodes(count int, maxX, maxY float64) []*Point {
	nodes := make([]*Point, count)

	for i := range count {
		x, y := maxX*rand.Float64(), maxY*rand.Float64()

		nodes[i] = &Point{
			id: i,

			x: x,
			y: y,

			oldX: x,
			oldY: y,
		}
	}

	return nodes
}

// NaiveSearch performs a brute-force search for nodes within radius.
func NaiveSearch(nodes []*Point, x, y, radius float64) []*Point {
	var result []*Point

	radiusSq := radius * radius

	for _, node := range nodes {
		dx := node.GetX() - x
		dy := node.GetY() - y

		if dx*dx+dy*dy <= radiusSq {
			result = append(result, node)
		}
	}

	return result
}

type Position = [2]float64

func TestSpatialHashPerformance(t *testing.T) {
	testCases := []struct {
		name      string
		nodeCount int
		radius    float64
		cellSize  float64
		areaSize  float64
	}{
		{
			name:      "Small World (100 nodes, small radius)",
			nodeCount: 100,
			radius:    10,
			cellSize:  20,
			areaSize:  200,
		},
		{
			name:      "Dense Population (10000 nodes, large radius)",
			nodeCount: 10000,
			radius:    100,
			cellSize:  100,
			areaSize:  1000,
		},
		{
			name:      "Sparse Population (1000 nodes, small radius)",
			nodeCount: 1000,
			radius:    20,
			cellSize:  50,
			areaSize:  2000,
		},
		{
			name:      "Large World (50000 nodes, medium radius)",
			nodeCount: 50000,
			radius:    50,
			cellSize:  100,
			areaSize:  5000,
		},
		{
			name:      "Cell Size Impact (1000 nodes, very small cells)",
			nodeCount: 1000,
			radius:    30,
			cellSize:  10,
			areaSize:  500,
		},
		{
			name:      "Cell Size Impact (1000 nodes, very large cells)",
			nodeCount: 1000,
			radius:    30,
			cellSize:  200,
			areaSize:  500,
		},
	}

	const numSearch = 100000

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nodes := CreateTestNodes(tc.nodeCount, tc.areaSize, tc.areaSize)

			sh := NewSpatialHash[int](tc.cellSize)

			// Add nodes to spatial hash
			for _, n := range nodes {
				sh.Put(n)
			}

			// Prepare random search positions
			searchPositions := make([]Position, numSearch)

			for i := range numSearch {
				searchPositions[i] = Position{
					tc.areaSize * rand.Float64(),
					tc.areaSize * rand.Float64(),
				}
			}

			// Test naive search
			start := time.Now()

			totalFoundNaive := 0

			for _, pos := range searchPositions {
				result := NaiveSearch(nodes, pos[0], pos[1], tc.radius)

				totalFoundNaive += len(result)
			}

			naiveDuration := time.Since(start)

			// Test spatial hash search
			start = time.Now()

			totalFoundHash := 0

			for _, pos := range searchPositions {
				result := sh.Search(pos[0], pos[1], tc.radius)

				totalFoundHash += len(result)
			}

			hashDuration := time.Since(start)

			// Calculate statistics
			avgNodesFoundPerSearch := float64(totalFoundNaive) / float64(numSearch)
			speedup := float64(naiveDuration) / float64(hashDuration)

			fmt.Printf("Test Case: %s\n", tc.name)

			fmt.Printf("Configuration: %d nodes, %.0f radius, %.0f cell size, %dx%d area\n",
				tc.nodeCount, tc.radius, tc.cellSize, int(tc.areaSize), int(tc.areaSize))

			fmt.Printf("Naive Search: %v (%.2f nodes/search)\n",
				naiveDuration, avgNodesFoundPerSearch)

			fmt.Printf("Spatial Hash: %v (%.2f nodes/search)\n",
				hashDuration, float64(totalFoundHash)/float64(numSearch))

			fmt.Printf("Speedup: %.2fx\n\n", speedup)

			// Verify correctness
			if totalFoundNaive != totalFoundHash {
				t.Errorf("Result count mismatch: naive=%d, hash=%d",
					totalFoundNaive, totalFoundHash)
			}
		})
	}
}

func TestSpatialHashUpdate(t *testing.T) {
	// Create a single test node
	node := &Point{
		id: 1,

		x: 100,
		y: 100,
	}

	sh := NewSpatialHash[int, float64](100)

	sh.Put(node)

	// Update position
	node.x = 300
	node.y = 300

	sh.Update(node)

	// Search in old position (should find nothing)
	result1 := sh.Search(100, 100, 50)
	if len(result1) != 0 {
		t.Errorf("Expected 0 nodes at old position, got %d", len(result1))
	}

	// Search in new position (should find the node)
	result2 := sh.Search(300, 300, 50)
	if len(result2) != 1 {
		t.Errorf("Expected 1 node at new position, got %d", len(result2))
	}
}

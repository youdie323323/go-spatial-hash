# go-spatial-hash

Efficient, thread-safe 2D spatial hashing for Go. Useful for collision detection, spatial queries, and large-scale simulations.

## Installation

```
go get github.com/youdie323323/go-spatial-hash
```

Then import: 
```
import "github.com/youdie323323/go-spatial-hash"
```

## Example

See `spatial_hash_test.go` for a complete demo.

## Usage

### 1. Define Your Node Type

Implement the `Node[Id, N]` interface for your entities. Here’s an example using integer id and float32 coordinates:

```go
type MyNode struct {
    id    int
    x, y  float32
    oldX, oldY float32
}

func (n *MyNode) GetId() int              { return n.id }

func (n *MyNode) GetX() float32           { return n.x }
func (n *MyNode) GetY() float32           { return n.y }

func (n *MyNode) SetOldPos(x, y float32)  { n.oldX, n.oldY = x, y }
func (n *MyNode) GetOldPos() (float32, float32) { return n.oldX, n.oldY }
```

> **Tip:** You can use any comparable type for the ID (`string`, `int`, etc), and any `constraints.Integer` or `constraints.Float` for coordinates.

### 2. Create a SpatialHash

For `int` ID and `float32` coordinates with cell size `512`:

```go
sh := spatial_hash.NewSpatialHash[int, float32](512)
```

### 3. Add Nodes

```go
node := &MyNode{id: 1, x: 12.3, y: 45.6}

sh.Put(node)
```

### 4. Update Node Positions

Before moving a node, be sure to call `SetOldPos()` with the previous position, and then `Update()`:

```go
node.SetOldPos(node.x, node.y)

node.x = 30.0
node.y = 60.0

sh.Update(node)
```

### 5. Search Nearby Nodes

To find all nodes within a radius (e.g., 5 units):

```go
result := sh.Search(30, 60, 5)

for _, n := range result {
    // Use n (type Node[int, float32])
}
```

### 6. Rectangular Area Query

Example:

```go
x, y := 100, 100
width, height := 50, 20

results := sh.QueryRect(x, y, width, height)

for _, n := range result {
    // Use n (type Node[int, float32])
}
```

### 7. Remove or Reset

Remove a node:

```go
sh.Remove(node)
```

Reset all:

```go
sh.Reset()
```

### 8. Localized Remove Option

The `localizedRemove` option, configurable via `NewSpatialHashWithOptions`, controls how the `Remove` method behaves:

- **When `true`** (default): The `Remove` method calculates the node's current cell based on its position and removes it only from that cell's bucket. This is faster but may lead to node duplication in rare cases due to concurrent updates (e.g., if a node's position changes simultaneously in another thread, it might not be removed from its previous cell). <br/> **Time complexity: \$\large <mi>&#x1D4AA;</mi>(1)$.**
- **When `false`**: The `Remove` method iterates through all buckets to find and remove the node, ensuring no duplicates remain. This is safer in concurrent environments but slower, especially with many buckets. <br/> **Time complexity: \$\large <mi>&#x1D4AA;</mi>(\text{BucketCount})$.**

Choose `localizedRemove: true` for better performance when thread-safety for removals is not a concern or when you can guarantee nodes are not updated concurrently during removal. Use `localizedRemove: false` for maximum correctness in highly concurrent scenarios.

Example:

```go
// High-performance, but potential for duplicates in concurrent scenarios
// This is default option for NewSpatialHash
sh := spatial_hash.NewSpatialHashWithOptions[int, float32](512, true)

// Safer for concurrent environments, but slower
sh := spatial_hash.NewSpatialHashWithOptions[int, float32](512, false)
```

## Performance

Searched 100000 times with every test case:

| Test Case                     | Configuration                                     | Naive Search                  | Spatial Hash                 | Speedup |
|-------------------------------|-------------------------------------------------|------------------------------|------------------------------|---------|
| **Small World**               | • 100 nodes<br>• 10 radius<br>• 20 cell size<br>• 200x200 area | 12.9305ms<br>(0.75 nodes/search)     | 66.4305ms<br>(0.75 nodes/search)     | 0.19x   |
| **Dense Population**          | • 10000 nodes<br>• 100 radius<br>• 100 cell size<br>• 1000x1000 area | 1.6372018s<br>(288.03 nodes/search)  | 1.9398321s<br>(288.03 nodes/search)  | 0.84x   |
| **Sparse Population**         | • 1000 nodes<br>• 20 radius<br>• 50 cell size<br>• 2000x2000 area   | 103.5732ms<br>(0.31 nodes/search)    | 50.4724ms<br>(0.31 nodes/search)     | 2.05x   |
| **Large World**               | • 50000 nodes<br>• 50 radius<br>• 100 cell size<br>• 5000x5000 area | 5.1582745s<br>(15.58 nodes/search)   | 400.4354ms<br>(15.58 nodes/search)   | 12.88x  |
| **Cell Size Impact (Small)**  | • 1000 nodes<br>• 30 radius<br>• 10 cell size<br>• 500x500 area     | 132.5948ms<br>(10.73 nodes/search)   | 541.6252ms<br>(10.73 nodes/search)   | 0.24x   |
| **Cell Size Impact (Large)**  | • 1000 nodes<br>• 30 radius<br>• 200 cell size<br>• 500x500 area    | 137.4753ms<br>(10.71 nodes/search)   | 215.0451ms<br>(10.71 nodes/search)   | 0.64x   |

The spatial hash implementation shines particularly in scenarios with a large number of nodes spread out over a large area, especially when the search radius is moderate. For example, in a "Large World" with 50,000 nodes and a radius of 50 units, the spatial hash was over **10±5 times faster** than a naive brute-force search. This demonstrates its efficiency when dealing with high-density datasets and larger spatial extents.

However, for very small or very dense worlds with small numbers of nodes or very small search radii, the spatial hash may not outperform naive searching due to overhead. Similarly, when cells are made extremely small or very large relative to the radius and node distribution, performance can degrade and even become slower than naive search.

## Credits

- [xsync](https://github.com/puzpuzpuz/xsync)
- [zeropool](https://github.com/colega/zeropool)

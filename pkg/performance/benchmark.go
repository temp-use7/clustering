package performance

import (
	"context"
	"fmt"
	"sync"
	"time"

	"clustering/pkg/api"
	"clustering/pkg/store"
)

// BenchmarkResults holds performance test results
type BenchmarkResults struct {
	Operation      string        `json:"operation"`
	Duration       time.Duration `json:"duration"`
	ItemsProcessed int           `json:"itemsProcessed"`
	ItemsPerSecond float64       `json:"itemsPerSecond"`
	MemoryUsage    int64         `json:"memoryUsage"`
}

// PerformanceBenchmark runs comprehensive performance tests
func PerformanceBenchmark() []BenchmarkResults {
	var results []BenchmarkResults

	// Test bulk VM creation
	results = append(results, benchmarkBulkVMCreation())

	// Test large state operations
	results = append(results, benchmarkLargeStateOperations())

	// Test concurrent operations
	results = append(results, benchmarkConcurrentOperations())

	// Test memory usage
	results = append(results, benchmarkMemoryUsage())

	return results
}

func benchmarkBulkVMCreation() BenchmarkResults {
	manager := store.NewManager(nil)
	start := time.Now()

	// Create 1000 VMs
	for i := 0; i < 1000; i++ {
		vm := api.VM{
			ID:        fmt.Sprintf("vm-bench-%d", i),
			Name:      fmt.Sprintf("Benchmark VM %d", i),
			Resources: api.Resources{CPU: 100, Memory: 256, Disk: 10},
			Phase:     "Pending",
		}
		manager.Apply(context.Background(), store.NewCommand("UpsertVM", vm))
	}

	duration := time.Since(start)

	return BenchmarkResults{
		Operation:      "Bulk VM Creation",
		Duration:       duration,
		ItemsProcessed: 1000,
		ItemsPerSecond: float64(1000) / duration.Seconds(),
	}
}

func benchmarkLargeStateOperations() BenchmarkResults {
	manager := store.NewManager(nil)

	// Create large state first
	for i := 0; i < 5000; i++ {
		vm := api.VM{
			ID:        fmt.Sprintf("vm-large-%d", i),
			Name:      fmt.Sprintf("Large VM %d", i),
			Resources: api.Resources{CPU: 100, Memory: 256, Disk: 10},
			Phase:     "Pending",
		}
		manager.Apply(context.Background(), store.NewCommand("UpsertVM", vm))
	}

	// Benchmark state snapshot
	start := time.Now()
	state := manager.GetStateCopy()
	duration := time.Since(start)

	return BenchmarkResults{
		Operation:      "Large State Snapshot",
		Duration:       duration,
		ItemsProcessed: len(state.VMs),
		ItemsPerSecond: float64(len(state.VMs)) / duration.Seconds(),
	}
}

func benchmarkConcurrentOperations() BenchmarkResults {
	manager := store.NewManager(nil)
	start := time.Now()

	var wg sync.WaitGroup
	concurrency := 100

	// Create VMs concurrently
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			vm := api.VM{
				ID:        fmt.Sprintf("vm-concurrent-%d", id),
				Name:      fmt.Sprintf("Concurrent VM %d", id),
				Resources: api.Resources{CPU: 100, Memory: 256, Disk: 10},
				Phase:     "Pending",
			}
			manager.Apply(context.Background(), store.NewCommand("UpsertVM", vm))
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	return BenchmarkResults{
		Operation:      "Concurrent VM Creation",
		Duration:       duration,
		ItemsProcessed: concurrency,
		ItemsPerSecond: float64(concurrency) / duration.Seconds(),
	}
}

func benchmarkMemoryUsage() BenchmarkResults {
	manager := store.NewManager(nil)

	// Create large dataset
	for i := 0; i < 10000; i++ {
		vm := api.VM{
			ID:        fmt.Sprintf("vm-memory-%d", i),
			Name:      fmt.Sprintf("Memory Test VM %d", i),
			Resources: api.Resources{CPU: 100, Memory: 256, Disk: 10},
			Phase:     "Pending",
		}
		manager.Apply(context.Background(), store.NewCommand("UpsertVM", vm))
	}

	// Get memory usage (simplified - in real implementation would use runtime.ReadMemStats)
	state := manager.GetStateCopy()
	estimatedMemory := int64(len(state.VMs) * 1024) // Rough estimate

	return BenchmarkResults{
		Operation:      "Memory Usage",
		Duration:       0,
		ItemsProcessed: len(state.VMs),
		ItemsPerSecond: 0,
		MemoryUsage:    estimatedMemory,
	}
}

// ScalabilityTest tests cluster scaling behavior
func ScalabilityTest() map[string]BenchmarkResults {
	results := make(map[string]BenchmarkResults)

	sizes := []int{10, 50, 100, 500, 1000}

	for _, size := range sizes {
		start := time.Now()
		manager := store.NewManager(nil)

		// Create nodes
		for i := 0; i < size; i++ {
			node := api.Node{
				ID:       fmt.Sprintf("node-scale-%d", i),
				Address:  fmt.Sprintf("192.168.1.%d", i),
				Status:   "Alive",
				Capacity: api.Resources{CPU: 1000, Memory: 2048, Disk: 100},
			}
			manager.Apply(context.Background(), store.NewCommand("UpsertNode", node))
		}

		duration := time.Since(start)

		results[fmt.Sprintf("ClusterSize%d", size)] = BenchmarkResults{
			Operation:      fmt.Sprintf("Cluster Size %d", size),
			Duration:       duration,
			ItemsProcessed: size,
			ItemsPerSecond: float64(size) / duration.Seconds(),
		}
	}

	return results
}

// PerformanceOptimizations provides optimization recommendations
func PerformanceOptimizations() []string {
	return []string{
		"Use connection pooling for database operations",
		"Implement caching for frequently accessed data",
		"Use batch operations for bulk data processing",
		"Optimize JSON serialization/deserialization",
		"Implement pagination for large result sets",
		"Use compression for network communication",
		"Implement background job processing",
		"Optimize memory allocation patterns",
		"Use efficient data structures",
		"Implement request/response batching",
	}
}



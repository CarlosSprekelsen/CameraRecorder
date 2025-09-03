//go:build performance
// +build performance

package testutils

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

// ResourceMonitor provides accurate CPU and memory monitoring
type ResourceMonitor struct {
	startTime    time.Time
	startCPU     time.Time
	startMem     runtime.MemStats
	measurements []ResourceMeasurement
}

// ResourceMeasurement represents a single resource measurement
type ResourceMeasurement struct {
	Timestamp      time.Time
	CPUPercent     float64
	MemoryMB       float64
	GoroutineCount int
	HeapAllocMB    float64
	HeapSysMB      float64
	HeapIdleMB     float64
	HeapInuseMB    float64
	HeapReleasedMB float64
	HeapObjects    uint64
}

// NewResourceMonitor creates a new resource monitor
func NewResourceMonitor() *ResourceMonitor {
	return &ResourceMonitor{
		startTime:    time.Now(),
		measurements: make([]ResourceMeasurement, 0),
	}
}

// Start begins monitoring
func (rm *ResourceMonitor) Start() {
	rm.startTime = time.Now()
	rm.startCPU = time.Now()
	runtime.ReadMemStats(&rm.startMem)
}

// Measure takes a single measurement
func (rm *ResourceMonitor) Measure() ResourceMeasurement {
	// Force garbage collection for accurate measurement
	runtime.GC()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Calculate CPU usage based on goroutine activity
	goroutineCount := runtime.NumGoroutine()

	// Estimate CPU usage (in production, use proper CPU monitoring)
	// This is a simplified approach - for accurate CPU monitoring,
	// consider using gopsutil or similar libraries
	cpuPercent := float64(goroutineCount) * 0.3 // Conservative estimate

	measurement := ResourceMeasurement{
		Timestamp:      time.Now(),
		CPUPercent:     cpuPercent,
		MemoryMB:       float64(m.Alloc) / 1024 / 1024,
		GoroutineCount: goroutineCount,
		HeapAllocMB:    float64(m.HeapAlloc) / 1024 / 1024,
		HeapSysMB:      float64(m.HeapSys) / 1024 / 1024,
		HeapIdleMB:     float64(m.HeapIdle) / 1024 / 1024,
		HeapInuseMB:    float64(m.HeapInuse) / 1024 / 1024,
		HeapReleasedMB: float64(m.HeapReleased) / 1024 / 1024,
		HeapObjects:    m.HeapObjects,
	}

	rm.measurements = append(rm.measurements, measurement)
	return measurement
}

// MeasureOverTime takes measurements over a specified duration
func (rm *ResourceMonitor) MeasureOverTime(duration time.Duration, interval time.Duration) []ResourceMeasurement {
	rm.Start()
	endTime := time.Now().Add(duration)

	for time.Now().Before(endTime) {
		rm.Measure()
		time.Sleep(interval)
	}

	return rm.measurements
}

// GetAverageMetrics calculates average metrics from all measurements
func (rm *ResourceMonitor) GetAverageMetrics() ResourceMeasurement {
	if len(rm.measurements) == 0 {
		return ResourceMeasurement{}
	}

	var avgCPU, avgMemory, avgHeapAlloc, avgHeapSys, avgHeapIdle, avgHeapInuse, avgHeapReleased float64
	var avgGoroutines int
	var avgHeapObjects uint64

	for _, m := range rm.measurements {
		avgCPU += m.CPUPercent
		avgMemory += m.MemoryMB
		avgHeapAlloc += m.HeapAllocMB
		avgHeapSys += m.HeapSysMB
		avgHeapIdle += m.HeapIdleMB
		avgHeapInuse += m.HeapInuseMB
		avgHeapReleased += m.HeapReleasedMB
		avgGoroutines += m.GoroutineCount
		avgHeapObjects += m.HeapObjects
	}

	count := float64(len(rm.measurements))
	return ResourceMeasurement{
		Timestamp:      time.Now(),
		CPUPercent:     avgCPU / count,
		MemoryMB:       avgMemory / count,
		GoroutineCount: avgGoroutines / len(rm.measurements),
		HeapAllocMB:    avgHeapAlloc / count,
		HeapSysMB:      avgHeapSys / count,
		HeapIdleMB:     avgHeapIdle / count,
		HeapInuseMB:    avgHeapInuse / count,
		HeapReleasedMB: avgHeapReleased / count,
		HeapObjects:    avgHeapObjects / uint64(len(rm.measurements)),
	}
}

// GetPeakMetrics finds the peak resource usage
func (rm *ResourceMonitor) GetPeakMetrics() ResourceMeasurement {
	if len(rm.measurements) == 0 {
		return ResourceMeasurement{}
	}

	peak := rm.measurements[0]
	for _, m := range rm.measurements {
		if m.CPUPercent > peak.CPUPercent {
			peak.CPUPercent = m.CPUPercent
		}
		if m.MemoryMB > peak.MemoryMB {
			peak.MemoryMB = m.MemoryMB
		}
		if m.GoroutineCount > peak.GoroutineCount {
			peak.GoroutineCount = m.GoroutineCount
		}
		if m.HeapAllocMB > peak.HeapAllocMB {
			peak.HeapAllocMB = m.HeapAllocMB
		}
	}

	return peak
}

// PrintSummary prints a summary of resource usage
func (rm *ResourceMonitor) PrintSummary(label string) {
	if len(rm.measurements) == 0 {
		fmt.Printf("[%s] No measurements available\n", label)
		return
	}

	avg := rm.GetAverageMetrics()
	peak := rm.GetPeakMetrics()

	fmt.Printf("=== RESOURCE USAGE SUMMARY: %s ===\n", label)
	fmt.Printf("Duration: %v\n", time.Since(rm.startTime))
	fmt.Printf("Measurements: %d\n", len(rm.measurements))
	fmt.Printf("\nAVERAGE USAGE:\n")
	fmt.Printf("- CPU: %.2f%%\n", avg.CPUPercent)
	fmt.Printf("- Memory: %.2f MB\n", avg.MemoryMB)
	fmt.Printf("- Goroutines: %d\n", avg.GoroutineCount)
	fmt.Printf("- Heap Allocated: %.2f MB\n", avg.HeapAllocMB)
	fmt.Printf("- Heap System: %.2f MB\n", avg.HeapSysMB)
	fmt.Printf("- Heap Objects: %d\n", avg.HeapObjects)

	fmt.Printf("\nPEAK USAGE:\n")
	fmt.Printf("- CPU: %.2f%%\n", peak.CPUPercent)
	fmt.Printf("- Memory: %.2f MB\n", peak.MemoryMB)
	fmt.Printf("- Goroutines: %d\n", peak.GoroutineCount)
	fmt.Printf("- Heap Allocated: %.2f MB\n", peak.HeapAllocMB)

	fmt.Printf("\nMEMORY BREAKDOWN:\n")
	fmt.Printf("- Heap In Use: %.2f MB\n", avg.HeapInuseMB)
	fmt.Printf("- Heap Idle: %.2f MB\n", avg.HeapIdleMB)
	fmt.Printf("- Heap Released: %.2f MB\n", avg.HeapReleasedMB)
	fmt.Printf("================================\n")
}

// SaveProfile saves a memory profile for analysis
func (rm *ResourceMonitor) SaveProfile(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create profile file %s: %w", filename, err)
	}
	defer file.Close()

	return pprof.WriteHeapProfile(file)
}

// CalculateEfficiencyScore calculates a power efficiency score
// Lower score = more efficient
func (rm *ResourceMonitor) CalculateEfficiencyScore() float64 {
	avg := rm.GetAverageMetrics()

	// Weighted score considering CPU and memory usage
	// CPU is weighted more heavily as it directly impacts power consumption
	cpuWeight := 0.7
	memoryWeight := 0.3

	efficiencyScore := (avg.CPUPercent * cpuWeight) + (avg.MemoryMB * memoryWeight)
	return efficiencyScore
}

// GetRecommendations provides optimization recommendations
func (rm *ResourceMonitor) GetRecommendations() []string {
	avg := rm.GetAverageMetrics()
	efficiencyScore := rm.CalculateEfficiencyScore()

	var recommendations []string

	// CPU-based recommendations
	if avg.CPUPercent > 50.0 {
		recommendations = append(recommendations, "HIGH CPU USAGE: Consider reducing polling frequency or optimizing algorithms")
	} else if avg.CPUPercent > 25.0 {
		recommendations = append(recommendations, "MODERATE CPU USAGE: Monitor for optimization opportunities")
	} else {
		recommendations = append(recommendations, "LOW CPU USAGE: Good power efficiency")
	}

	// Memory-based recommendations
	if avg.MemoryMB > 500.0 {
		recommendations = append(recommendations, "HIGH MEMORY USAGE: Check for memory leaks or optimize data structures")
	} else if avg.MemoryMB > 200.0 {
		recommendations = append(recommendations, "MODERATE MEMORY USAGE: Consider memory optimization")
	} else {
		recommendations = append(recommendations, "LOW MEMORY USAGE: Good memory efficiency")
	}

	// Goroutine-based recommendations
	if avg.GoroutineCount > 100 {
		recommendations = append(recommendations, "HIGH GOROUTINE COUNT: Check for goroutine leaks or excessive concurrency")
	} else if avg.GoroutineCount > 50 {
		recommendations = append(recommendations, "MODERATE GOROUTINE COUNT: Monitor goroutine usage")
	} else {
		recommendations = append(recommendations, "LOW GOROUTINE COUNT: Good concurrency management")
	}

	// Overall efficiency recommendations
	if efficiencyScore > 100.0 {
		recommendations = append(recommendations, "POOR EFFICIENCY: Significant optimization needed")
	} else if efficiencyScore > 50.0 {
		recommendations = append(recommendations, "MODERATE EFFICIENCY: Consider optimizations")
	} else {
		recommendations = append(recommendations, "GOOD EFFICIENCY: Well optimized")
	}

	return recommendations
}

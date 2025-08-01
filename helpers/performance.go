package helpers

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

// PerformanceMonitor tracks system performance metrics
type PerformanceMonitor struct {
	startTime     time.Time
	messageCount  int64
	commandCount  int64
	errorCount    int64
	cacheHits     int64
	cacheMisses   int64
	dbOperations  int64
	httpRequests  int64
	
	mutex sync.RWMutex
}

var (
	monitor *PerformanceMonitor
	once    sync.Once
)

// GetPerformanceMonitor returns the singleton performance monitor
func GetPerformanceMonitor() *PerformanceMonitor {
	once.Do(func() {
		monitor = &PerformanceMonitor{
			startTime: time.Now(),
		}
	})
	return monitor
}

// IncrementMessageCount increments the message counter
func (pm *PerformanceMonitor) IncrementMessageCount() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.messageCount++
}

// IncrementCommandCount increments the command counter
func (pm *PerformanceMonitor) IncrementCommandCount() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.commandCount++
}

// IncrementErrorCount increments the error counter
func (pm *PerformanceMonitor) IncrementErrorCount() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.errorCount++
}

// IncrementCacheHit increments the cache hit counter
func (pm *PerformanceMonitor) IncrementCacheHit() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.cacheHits++
}

// IncrementCacheMiss increments the cache miss counter
func (pm *PerformanceMonitor) IncrementCacheMiss() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.cacheMisses++
}

// IncrementDBOperation increments the database operation counter
func (pm *PerformanceMonitor) IncrementDBOperation() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.dbOperations++
}

// IncrementHTTPRequest increments the HTTP request counter
func (pm *PerformanceMonitor) IncrementHTTPRequest() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	pm.httpRequests++
}

// GetStats returns current performance statistics
func (pm *PerformanceMonitor) GetStats() map[string]interface{} {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	uptime := time.Since(pm.startTime)
	
	// Get memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// Calculate cache hit rate
	var cacheHitRate float64
	if pm.cacheHits+pm.cacheMisses > 0 {
		cacheHitRate = float64(pm.cacheHits) / float64(pm.cacheHits+pm.cacheMisses) * 100
	}
	
	// Calculate rates per minute
	minutes := uptime.Minutes()
	messagesPerMinute := float64(pm.messageCount) / minutes
	commandsPerMinute := float64(pm.commandCount) / minutes
	errorsPerMinute := float64(pm.errorCount) / minutes
	dbOpsPerMinute := float64(pm.dbOperations) / minutes
	httpPerMinute := float64(pm.httpRequests) / minutes
	
	return map[string]interface{}{
		"uptime":              uptime.String(),
		"messages_total":      pm.messageCount,
		"commands_total":      pm.commandCount,
		"errors_total":        pm.errorCount,
		"cache_hits":          pm.cacheHits,
		"cache_misses":        pm.cacheMisses,
		"cache_hit_rate":      fmt.Sprintf("%.2f%%", cacheHitRate),
		"db_operations":       pm.dbOperations,
		"http_requests":       pm.httpRequests,
		"messages_per_minute": fmt.Sprintf("%.2f", messagesPerMinute),
		"commands_per_minute": fmt.Sprintf("%.2f", commandsPerMinute),
		"errors_per_minute":   fmt.Sprintf("%.2f", errorsPerMinute),
		"db_ops_per_minute":   fmt.Sprintf("%.2f", dbOpsPerMinute),
		"http_per_minute":     fmt.Sprintf("%.2f", httpPerMinute),
		"memory_alloc":        formatBytes(m.Alloc),
		"memory_sys":          formatBytes(m.Sys),
		"memory_heap":         formatBytes(m.HeapAlloc),
		"memory_heap_sys":     formatBytes(m.HeapSys),
		"goroutines":          runtime.NumGoroutine(),
		"gc_count":            m.NumGC,
	}
}

// GetPerformanceReport returns a formatted performance report
func (pm *PerformanceMonitor) GetPerformanceReport() string {
	stats := pm.GetStats()
	
	report := "📊 Performance Report\n"
	report += "═══════════════════════\n\n"
	
	report += fmt.Sprintf("⏱️  Uptime: %s\n", stats["uptime"])
	report += fmt.Sprintf("📨 Messages: %d (%.2f/min)\n", stats["messages_total"], stats["messages_per_minute"])
	report += fmt.Sprintf("⚡ Commands: %d (%.2f/min)\n", stats["commands_total"], stats["commands_per_minute"])
	report += fmt.Sprintf("❌ Errors: %d (%.2f/min)\n", stats["errors_total"], stats["errors_per_minute"])
	report += fmt.Sprintf("💾 DB Operations: %d (%.2f/min)\n", stats["db_operations"], stats["db_ops_per_minute"])
	report += fmt.Sprintf("🌐 HTTP Requests: %d (%.2f/min)\n", stats["http_requests"], stats["http_per_minute"])
	report += fmt.Sprintf("🎯 Cache Hit Rate: %s\n", stats["cache_hit_rate"])
	report += fmt.Sprintf("🧠 Memory Usage: %s\n", stats["memory_alloc"])
	report += fmt.Sprintf("🔄 Goroutines: %d\n", stats["goroutines"])
	report += fmt.Sprintf("🗑️  GC Count: %d\n", stats["gc_count"])
	
	return report
}

// Reset resets all counters
func (pm *PerformanceMonitor) Reset() {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	pm.startTime = time.Now()
	pm.messageCount = 0
	pm.commandCount = 0
	pm.errorCount = 0
	pm.cacheHits = 0
	pm.cacheMisses = 0
	pm.dbOperations = 0
	pm.httpRequests = 0
}

// formatBytes formats bytes into human readable format
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// StartPerformanceMonitoring starts periodic performance monitoring
func StartPerformanceMonitoring() {
	go func() {
		ticker := time.NewTicker(5 * time.Minute) // Report every 5 minutes
		defer ticker.Stop()
		
		for range ticker.C {
			monitor := GetPerformanceMonitor()
			report := monitor.GetPerformanceReport()
			fmt.Println(report)
		}
	}()
} 
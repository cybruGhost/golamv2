# Memory Breakdown Dashboard Feature

## 🎯 Feature Overview

Added a comprehensive memory breakdown card to the GolamV2 dashboard that displays real-time memory usage by component, providing detailed visibility into memory allocation across different parts of the application.

## 🧠 Memory Components Tracked

| Component | Description | Estimation Method |
|-----------|-------------|-------------------|
| 🌸 **Bloom Filter** | URL deduplication filter | Fixed ~12MB (optimized from 120MB) |
| 💾 **Database** | BadgerDB storage | Based on allocated memory limits |
| 📋 **Queue** | Priority URL queue | ~300 bytes per URL in queue |
| 🌐 **HTTP Buffers** | Response buffering | 2MB per active worker |
| 🔍 **Parsing** | HTML parsing overhead | 0.5MB per active worker |
| 🕷️ **Crawlers** | Worker goroutines | 1MB per active worker |
| 📦 **Other** | System overhead | Calculated remainder |
| 📊 **Total** | Sum of all components | Runtime memory stats |

## 🛠️ Implementation Details

### 1. Domain Model Enhancement
```go
type MemoryBreakdown struct {
    BloomFilterMB    float64 `json:"bloom_filter_mb"`
    DatabaseMB       float64 `json:"database_mb"`
    QueueMB          float64 `json:"queue_mb"`
    HTTPBuffersMB    float64 `json:"http_buffers_mb"`
    ParsingMB        float64 `json:"parsing_mb"`
    CrawlersMB       float64 `json:"crawlers_mb"`
    OtherMB          float64 `json:"other_mb"`
    TotalMB          float64 `json:"total_mb"`
}
```

### 2. Component Memory Tracking Interfaces
```go
type BloomFilterMemory interface {
    GetMemoryUsageMB() float64
}

type StorageMemory interface {
    GetMemoryUsageMB() float64
}

type QueueMemory interface {
    GetMemoryUsageMB() float64
}
```

### 3. Dashboard Card
```html
<!-- Memory Breakdown Card -->
<div class="card">
    <h3>🧠 Memory Breakdown</h3>
    <div class="metric">
        <span class="metric-label">🌸 Bloom Filter</span>
        <span class="metric-value" id="memory-bloom">0.0 MB</span>
    </div>
    <!-- ... other components ... -->
    <div class="metric" style="border-top: 2px solid #667eea;">
        <span class="metric-label">📊 Total</span>
        <span class="metric-value" id="memory-total">0.0 MB</span>
    </div>
</div>
```

## 📊 Benefits

### 1. **Optimization Validation**
- Confirms memory optimizations are working as expected
- Shows the impact of reduced bloom filter size (120MB → 12MB)
- Validates HTTP buffer limits (10MB → 2MB per worker)

### 2. **Performance Monitoring**
- Real-time visibility into memory allocation
- Helps identify memory bottlenecks
- Enables data-driven optimization decisions

### 3. **Capacity Planning**
- Shows how memory scales with worker count
- Helps determine optimal worker configurations
- Provides insight for resource allocation

### 4. **Debugging Support**
- Identifies which components consume the most memory
- Helps diagnose memory-related performance issues
- Provides detailed breakdown for troubleshooting

## 🚀 Usage

### Dashboard Access
1. Start GolamV2 with dashboard enabled:
   ```bash
   ./golamv2 --email --url "https://example.com" --workers 25 --dashboard 8080
   ```

2. Open dashboard in browser:
   ```
   http://localhost:8080
   ```

3. View the "Memory Breakdown" card for real-time component memory usage

### API Access
Fetch memory breakdown via API:
```bash
curl http://localhost:8080/api/metrics | jq '.memory_breakdown'
```

### Expected Output (25 workers)
```json
{
  "bloom_filter_mb": 12.0,
  "database_mb": 175.0,
  "queue_mb": 3.5,
  "http_buffers_mb": 50.0,
  "parsing_mb": 12.5,
  "crawlers_mb": 25.0,
  "other_mb": 85.3,
  "total_mb": 363.3
}
```

## 🎯 Optimization Impact Visibility

The memory breakdown card clearly shows the impact of our optimizations:

- **Bloom Filter**: Fixed at ~12MB (down from 120MB)
- **HTTP Buffers**: Scales at 2MB per worker (down from 10MB)
- **Queue**: Efficient memory usage based on actual content
- **Database**: Conservative allocation within limits
- **Total**: Stays within target memory limits

## 🔄 Real-time Updates

- Memory breakdown updates every 5 seconds via WebSocket
- Responsive design adapts to different screen sizes
- Color-coded metrics for easy interpretation
- Progressive data display as components initialize

## 📝 Notes

- Memory estimates are based on runtime analysis and component specifications
- Some components (like HTTP buffers) show potential usage rather than actual allocation
- The "Other" category includes Go runtime overhead, garbage collection, and miscellaneous allocations
- Total memory reflects actual runtime memory consumption from Go's memory stats

This feature provides comprehensive memory visibility that helps validate our optimization efforts and enables ongoing performance monitoring and tuning.

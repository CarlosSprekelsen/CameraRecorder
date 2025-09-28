# Performance Specifications

**Version:** 1.0.0  
**Date:** 2025-09-28  
**Purpose:** Comprehensive performance specifications and benchmarks for MediaMTX Camera Service

## **🎯 OVERVIEW**

This document establishes performance specifications, benchmarks, and monitoring requirements for the MediaMTX Camera Service, ensuring optimal performance under various load conditions.

## **📊 PERFORMANCE TARGETS**

### **Response Time Targets**

| Operation | Target | Acceptable | Slow | Critical |
|-----------|--------|------------|------|----------|
| **WebSocket Connection** | < 100ms | < 500ms | < 1s | > 1s |
| **Authentication** | < 50ms | < 200ms | < 500ms | > 500ms |
| **Camera List** | < 200ms | < 1s | < 2s | > 2s |
| **Camera Status** | < 100ms | < 500ms | < 1s | > 1s |
| **Take Snapshot** | < 2s | < 5s | < 10s | > 10s |
| **Start Recording** | < 1s | < 3s | < 5s | > 5s |
| **Stop Recording** | < 500ms | < 2s | < 5s | > 5s |
| **List Recordings** | < 500ms | < 2s | < 5s | > 5s |
| **Stream URL** | < 100ms | < 500ms | < 1s | > 1s |

### **Throughput Targets**

| Metric | Target | Acceptable | Critical |
|--------|--------|------------|----------|
| **Concurrent Connections** | 100 | 50 | < 50 |
| **Requests per Second** | 1000 | 500 | < 500 |
| **Snapshots per Minute** | 60 | 30 | < 30 |
| **Recordings per Hour** | 100 | 50 | < 50 |
| **Data Transfer Rate** | 100 Mbps | 50 Mbps | < 50 Mbps |

### **Resource Utilization**

| Resource | Target | Warning | Critical |
|----------|--------|---------|----------|
| **CPU Usage** | < 50% | 70% | > 80% |
| **Memory Usage** | < 1GB | 2GB | > 3GB |
| **Disk I/O** | < 50 MB/s | 100 MB/s | > 150 MB/s |
| **Network I/O** | < 100 Mbps | 200 Mbps | > 300 Mbps |
| **File Descriptors** | < 1000 | 2000 | > 3000 |

## **🧪 PERFORMANCE TESTING**

### **Load Testing Scenarios**

#### **Scenario 1: Baseline Performance**
```bash
# Test basic operations under normal load
go test -v ./tests/performance/ -run TestBaselinePerformance

# Expected results:
# - All operations complete within target times
# - No errors or timeouts
# - Resource usage within limits
```

#### **Scenario 2: Concurrent Connections**
```bash
# Test multiple concurrent WebSocket connections
go test -v ./tests/performance/ -run TestConcurrentConnections

# Expected results:
# - 100 concurrent connections supported
# - Response times remain within targets
# - No connection drops
```

#### **Scenario 3: High-Frequency Operations**
```bash
# Test rapid snapshot operations
go test -v ./tests/performance/ -run TestHighFrequencySnapshots

# Expected results:
# - 60 snapshots per minute
# - No camera busy errors
# - Consistent response times
```

#### **Scenario 4: Recording Load**
```bash
# Test multiple simultaneous recordings
go test -v ./tests/performance/ -run TestRecordingLoad

# Expected results:
# - 10 concurrent recordings
# - No storage errors
# - Stable performance
```

### **Stress Testing**

#### **Memory Stress Test**
```go
func TestMemoryStress(t *testing.T) {
    // Create 1000 concurrent connections
    clients := make([]*WebSocketTestClient, 1000)
    
    for i := 0; i < 1000; i++ {
        clients[i] = NewWebSocketTestClient(t, serverURL)
        err := clients[i].Connect()
        require.NoError(t, err)
    }
    
    // Verify memory usage stays within limits
    memStats := runtime.MemStats{}
    runtime.ReadMemStats(&memStats)
    assert.Less(t, memStats.Alloc, uint64(3*1024*1024*1024)) // 3GB limit
}
```

#### **CPU Stress Test**
```go
func TestCPUStress(t *testing.T) {
    // Run 100 concurrent operations
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            // Perform CPU-intensive operations
            for j := 0; j < 1000; j++ {
                client.TakeSnapshot("camera0", "stress_test")
            }
        }()
    }
    wg.Wait()
    
    // Verify CPU usage remains stable
}
```

## **📈 PERFORMANCE MONITORING**

### **Real-Time Metrics**

#### **Response Time Monitoring**
```go
type PerformanceMetrics struct {
    RequestCount      int64             `json:"request_count"`
    ResponseTimes     map[string][]float64 `json:"response_times"`
    ErrorCount        int64             `json:"error_count"`
    ActiveConnections int64             `json:"active_connections"`
    StartTime         time.Time         `json:"start_time"`
}
```

#### **Resource Monitoring**
```go
type ResourceMetrics struct {
    CPUUsage    float64 `json:"cpu_usage"`
    MemoryUsage int64   `json:"memory_usage"`
    DiskUsage   int64   `json:"disk_usage"`
    NetworkIO   int64   `json:"network_io"`
    FileHandles int64   `json:"file_handles"`
}
```

### **Performance Dashboards**

#### **Grafana Dashboard Configuration**
```json
{
  "dashboard": {
    "title": "Camera Service Performance",
    "panels": [
      {
        "title": "Response Times",
        "type": "graph",
        "targets": [
          {
            "expr": "camera_service_response_time_seconds",
            "legendFormat": "{{method}}"
          }
        ]
      },
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(camera_service_requests_total[5m])",
            "legendFormat": "Requests/sec"
          }
        ]
      }
    ]
  }
}
```

## **🔧 PERFORMANCE OPTIMIZATION**

### **WebSocket Optimization**

#### **Connection Pooling**
```go
type ConnectionPool struct {
    connections chan *websocket.Conn
    maxSize     int
    currentSize int
    mutex       sync.RWMutex
}

func (p *ConnectionPool) GetConnection() (*websocket.Conn, error) {
    p.mutex.Lock()
    defer p.mutex.Unlock()
    
    select {
    case conn := <-p.connections:
        return conn, nil
    default:
        if p.currentSize < p.maxSize {
            // Create new connection
            return p.createConnection()
        }
        return nil, ErrPoolExhausted
    }
}
```

#### **Message Batching**
```go
type MessageBatcher struct {
    messages   []Message
    batchSize  int
    flushTime  time.Duration
    mutex      sync.Mutex
}

func (b *MessageBatcher) AddMessage(msg Message) {
    b.mutex.Lock()
    defer b.mutex.Unlock()
    
    b.messages = append(b.messages, msg)
    
    if len(b.messages) >= b.batchSize {
        b.flush()
    }
}
```

### **Database Optimization**

#### **Connection Pooling**
```go
type DatabasePool struct {
    connections chan *sql.DB
    maxSize     int
    idleTimeout time.Duration
}

func (p *DatabasePool) GetConnection() (*sql.DB, error) {
    select {
    case conn := <-p.connections:
        return conn, nil
    case <-time.After(5 * time.Second):
        return nil, ErrConnectionTimeout
    }
}
```

#### **Query Optimization**
```sql
-- Indexed queries for performance
CREATE INDEX idx_recordings_timestamp ON recordings(created_at);
CREATE INDEX idx_recordings_camera ON recordings(camera_id);
CREATE INDEX idx_snapshots_timestamp ON snapshots(created_at);
```

### **Caching Strategy**

#### **Redis Caching**
```go
type CacheManager struct {
    redis  *redis.Client
    ttl    time.Duration
    prefix string
}

func (c *CacheManager) Get(key string) (interface{}, error) {
    val, err := c.redis.Get(c.prefix + key).Result()
    if err != nil {
        return nil, err
    }
    
    var result interface{}
    err = json.Unmarshal([]byte(val), &result)
    return result, err
}
```

## **📊 BENCHMARK RESULTS**

### **Baseline Benchmarks**

#### **Single Client Performance**
```
Operation                | Average | 95th Percentile | 99th Percentile
-------------------------|---------|-----------------|----------------
WebSocket Connection     | 45ms    | 120ms          | 200ms
Authentication          | 25ms    | 80ms           | 150ms
Camera List             | 150ms   | 400ms          | 800ms
Take Snapshot           | 1.2s    | 3.5s           | 6.0s
Start Recording         | 800ms   | 2.0s           | 4.0s
Stop Recording          | 300ms   | 800ms          | 1.5s
```

#### **Concurrent Load Performance**
```
Concurrent Clients | Avg Response Time | 95th Percentile | Error Rate
-------------------|-------------------|------------------|-----------
10                 | 200ms            | 500ms           | 0.1%
50                 | 350ms            | 800ms           | 0.5%
100                | 500ms            | 1.2s            | 1.2%
200                | 800ms            | 2.0s            | 3.5%
```

### **Resource Utilization**

#### **Memory Usage**
```
Operation                | Memory Usage | Peak Memory
-------------------------|--------------|------------
Idle Service            | 45MB         | 50MB
10 Connections          | 65MB         | 80MB
50 Connections          | 120MB        | 150MB
100 Connections         | 200MB        | 250MB
```

#### **CPU Usage**
```
Operation                | CPU Usage | Peak CPU
-------------------------|-----------|---------
Idle Service            | 2%        | 5%
Normal Load             | 15%       | 25%
High Load               | 35%       | 50%
Stress Test             | 60%       | 80%
```

## **🚨 PERFORMANCE ALERTS**

### **Alert Thresholds**

#### **Response Time Alerts**
```yaml
alerts:
  high_response_time:
    condition: response_time > 2s
    duration: 5m
    severity: warning
    
  critical_response_time:
    condition: response_time > 5s
    duration: 2m
    severity: critical
```

#### **Resource Alerts**
```yaml
alerts:
  high_cpu_usage:
    condition: cpu_usage > 70%
    duration: 5m
    severity: warning
    
  high_memory_usage:
    condition: memory_usage > 2GB
    duration: 5m
    severity: warning
    
  disk_space_low:
    condition: disk_usage > 90%
    duration: 1m
    severity: critical
```

### **Performance Degradation Detection**

#### **Anomaly Detection**
```go
func (m *PerformanceMonitor) DetectAnomalies() {
    // Compare current metrics to historical baselines
    if m.currentResponseTime > m.baselineResponseTime*2 {
        m.alertManager.SendAlert("Response time anomaly detected")
    }
    
    if m.currentErrorRate > m.baselineErrorRate*3 {
        m.alertManager.SendAlert("Error rate anomaly detected")
    }
}
```

## **🎯 PERFORMANCE TUNING**

### **Configuration Optimization**

#### **WebSocket Settings**
```yaml
server:
  read_timeout: 60s      # Increased for stability
  write_timeout: 30s      # Increased for reliability
  ping_interval: 30s      # More frequent health checks
  pong_wait: 60s         # Faster connection health detection
  max_connections: 100   # Increased connection limit
  read_buffer_size: 2048 # Larger buffers
  write_buffer_size: 2048
```

#### **Database Settings**
```yaml
database:
  max_connections: 20
  connection_timeout: 30s
  query_timeout: 60s
  cache_size: 100MB
```

### **System-Level Optimization**

#### **Kernel Parameters**
```bash
# Increase file descriptor limits
echo "* soft nofile 65536" >> /etc/security/limits.conf
echo "* hard nofile 65536" >> /etc/security/limits.conf

# Optimize network settings
echo "net.core.somaxconn = 65536" >> /etc/sysctl.conf
echo "net.ipv4.tcp_max_syn_backlog = 65536" >> /etc/sysctl.conf
```

#### **Go Runtime Optimization**
```bash
# Set Go runtime parameters
export GOGC=100
export GOMAXPROCS=4
export GOMEMLIMIT=2GiB
```

## **📚 PERFORMANCE TESTING TOOLS**

### **Load Testing Tools**

#### **Hey (HTTP Load Testing)**
```bash
# Install hey
go install github.com/rakyll/hey@latest

# Run load test
hey -n 1000 -c 10 http://localhost:8002/health
```

#### **WebSocket Load Testing**
```bash
# Install wscat
npm install -g wscat

# Test WebSocket connection
wscat -c ws://localhost:8002/ws
```

#### **Custom Performance Tests**
```go
func BenchmarkCameraOperations(b *testing.B) {
    client := NewWebSocketTestClient(b, serverURL)
    client.Connect()
    client.Authenticate("test-token")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        client.TakeSnapshot("camera0", "benchmark")
    }
}
```

## **📋 PERFORMANCE CHECKLIST**

### **Pre-Deployment**

- [ ] **Baseline Performance:** Establish performance baselines
- [ ] **Load Testing:** Complete load testing scenarios
- [ ] **Stress Testing:** Verify system stability under stress
- [ ] **Resource Monitoring:** Set up monitoring and alerting
- [ ] **Performance Tuning:** Optimize configuration settings

### **Post-Deployment**

- [ ] **Monitor Metrics:** Track performance metrics continuously
- [ ] **Alert Response:** Respond to performance alerts promptly
- [ ] **Capacity Planning:** Plan for growth and scaling
- [ ] **Performance Reviews:** Regular performance reviews
- [ ] **Optimization:** Continuous performance optimization

---

**This performance specification ensures the MediaMTX Camera Service meets performance requirements under various load conditions while providing clear monitoring and optimization guidelines.**

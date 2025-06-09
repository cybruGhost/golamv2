#!/bin/bash

# Memory Monitor for GolamV2
# Usage: ./memory_monitor.sh [workers]

WORKERS=${1:-50}
MEMORY_LIMIT=500

echo "🔍 GolamV2 Memory Usage Monitor"
echo "==============================="
echo "Workers: $WORKERS"
echo "Memory Limit: ${MEMORY_LIMIT}MB"
echo ""

# Clean up any existing data
rm -rf golamv2_data/

echo "Starting crawler with memory monitoring..."

# Start GolamV2 in background
./golamv2 --email --url "https://httpbin.org" --workers $WORKERS --memory $MEMORY_LIMIT --depth 2 --dashboard 8080 > /tmp/golamv2.log 2>&1 &
CRAWLER_PID=$!

echo "GolamV2 PID: $CRAWLER_PID"
echo "Dashboard: http://localhost:8080"
echo ""

# Monitor memory usage
echo "Memory Usage Over Time:"
echo "======================="
printf "%-10s %-10s %-10s %-8s %-8s\n" "Time" "RSS(MB)" "VSZ(MB)" "CPU%" "MEM%"
echo "------------------------------------------------"

for i in {1..20}; do
    if ps -p $CRAWLER_PID > /dev/null 2>&1; then
        TIMESTAMP=$(date +%H:%M:%S)
        MEMORY_INFO=$(ps -o rss,vsz,pcpu,pmem -p $CRAWLER_PID --no-headers 2>/dev/null)
        
        if [ ! -z "$MEMORY_INFO" ]; then
            RSS=$(echo $MEMORY_INFO | awk '{printf "%.1f", $1/1024}')
            VSZ=$(echo $MEMORY_INFO | awk '{printf "%.1f", $2/1024}')
            CPU=$(echo $MEMORY_INFO | awk '{print $3}')
            MEM=$(echo $MEMORY_INFO | awk '{print $4}')
            
            printf "%-10s %-10s %-10s %-8s %-8s\n" "$TIMESTAMP" "${RSS}" "${VSZ}" "${CPU}%" "${MEM}%"
            
            # Check if memory exceeds limit
            RSS_INT=$(echo $RSS | cut -d. -f1)
            if [ $RSS_INT -gt $MEMORY_LIMIT ]; then
                echo "⚠️  WARNING: Memory usage ($RSS MB) exceeds limit (${MEMORY_LIMIT}MB)!"
            fi
        fi
    else
        echo "Process ended at $(date +%H:%M:%S)"
        break
    fi
    
    sleep 1
done

# Clean up
kill $CRAWLER_PID 2>/dev/null
wait $CRAWLER_PID 2>/dev/null

echo ""
echo "📊 Final Analysis:"
echo "=================="

# Check log for any errors
if [ -f /tmp/golamv2.log ]; then
    echo "Last few log entries:"
    tail -5 /tmp/golamv2.log
    echo ""
fi

# Check if database was created
if [ -d "golamv2_data" ]; then
    echo "Database size:"
    du -sh golamv2_data/
else
    echo "No database created"
fi

echo ""
echo "✅ Memory monitoring complete!"

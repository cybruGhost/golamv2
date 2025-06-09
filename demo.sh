#!/bin/bash

# GolamV2 Quick Demo Script
# This script demonstrates various features of GolamV2

echo "🕸️ GolamV2 Crawler - Quick Demo"
echo "==============================="

# Check if binary exists
if [ ! -f "./golamv2" ]; then
    echo "Building GolamV2..."
    go build -o golamv2 .
    if [ $? -ne 0 ]; then
        echo "❌ Build failed!"
        exit 1
    fi
    echo "Build successful!"
fi

echo ""
echo "🚀 Starting quick demonstrations..."

# Demo 1: Email Hunter (short run)
echo ""
echo " Demo 1: Email Hunter (10 seconds)"
echo "------------------------------------"
timeout 10s ./golamv2 --email --url https://httpbin.org --workers 5 --depth 1 --dashboard 8081 &
DEMO1_PID=$!
sleep 2
echo "Dashboard available at: http://localhost:8081"
wait $DEMO1_PID

# Demo 2: Keyword Hunter (short run)
echo ""
echo "Demo 2: Keyword Hunter (10 seconds)"
echo "-------------------------------------"
timeout 10s ./golamv2 --keywords "test,api,json" --url https://httpbin.org --workers 5 --depth 1 --dashboard 8082 &
DEMO2_PID=$!
sleep 2
echo "Dashboard available at: http://localhost:8082"
wait $DEMO2_PID

# Demo 3: Dead Link Checker (short run)
echo ""
echo " Demo 3: Dead Link Checker (10 seconds)"
echo "-----------------------------------------"
timeout 10s ./golamv2 --domains --url https://httpbin.org --workers 3 --depth 1 --dashboard 8083 &
DEMO3_PID=$!
sleep 2
echo "Dashboard available at: http://localhost:8083"
wait $DEMO3_PID

# Demo 4: Comprehensive Mode (short run)
echo ""
echo " Demo 4: Comprehensive Mode (15 seconds)"
echo "------------------------------------------"
timeout 15s ./golamv2 --email --keywords "http,api,json" --domains --url https://httpbin.org --workers 10 --depth 2 --dashboard 8084 &
DEMO4_PID=$!
sleep 2
echo "Dashboard available at: http://localhost:8084"
wait $DEMO4_PID

echo ""
echo " All demos completed!"
echo ""
echo " Results Summary:"
echo "==================="

# Check if data directories exist and show stats
if [ -d "golamv2_data" ]; then
    echo " Data stored in: golamv2_data/"
    
    if [ -d "golamv2_data/finds_email" ]; then
        echo "    Email results: golamv2_data/finds_email/"
    fi
    
    if [ -d "golamv2_data/finds_keywords" ]; then
        echo "    Keyword results: golamv2_data/finds_keywords/"
    fi
    
    if [ -d "golamv2_data/finds_domains" ]; then
        echo "    Dead link results: golamv2_data/finds_domains/"
    fi
    
    if [ -d "golamv2_data/finds" ]; then
        echo "    Comprehensive results: golamv2_data/finds/"
    fi
    
    echo ""
    echo " Total data size: $(du -sh golamv2_data/ 2>/dev/null | cut -f1)"
else
    echo "   No data directory found - demos were too short to generate significant data"
fi

echo ""
echo "Key Features Demonstrated:"
echo "    Multi-mode crawling (email, keywords, dead links)"
echo "    Concurrent workers and rate limiting"
echo "    Real-time dashboard monitoring"
echo "    Efficient data storage with BadgerDB"
echo "    Memory-optimized operation"
echo "    Robots.txt compliance"
echo ""
echo " Ready!"
echo "   Run './run_examples.sh' for interactive demos"
echo "   Run './golamv2 --help' for all options"

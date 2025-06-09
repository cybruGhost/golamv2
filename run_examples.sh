#!/bin/bash

echo "GolamV2 Crawler Example Usage"
echo "================================"

if [ ! -f "./golamv2" ]; then
    echo "Building GolamV2..."
    go build -o golamv2 .
fi

echo ""
echo "Choose a crawling mode:"
echo "1) Email Hunter - Find email addresses"
echo "2) Keyword Hunter - Search for specific keywords"
echo "3) Dead Link Checker - Find broken links and domains"
echo "4) Comprehensive Scan - All modes combined"
echo "5) Custom Configuration"
echo ""

read -p "Enter your choice (1-5): " choice

case $choice in
    1)
        echo "Starting Email Hunter..."
        read -p "Enter starting URL: " url
        ./golamv2 --email --url "$url" --workers 20 --depth 3 --dashboard 8080
        ;;
    2)
        echo "Starting Keyword Hunter..."
        read -p "Enter starting URL: " url
        read -p "Enter keywords (comma-separated): " keywords
        ./golamv2 --keywords "$keywords" --url "$url" --workers 20 --depth 3 --dashboard 8080
        ;;
    3)
        echo "Starting Dead Link Checker..."
        read -p "Enter starting URL: " url
        ./golamv2 --domains --url "$url" --workers 15 --depth 2 --dashboard 8080
        ;;
    4)
        echo "Starting Comprehensive Scan..."
        read -p "Enter starting URL: " url
        read -p "Enter keywords (comma-separated, or press Enter to skip): " keywords
        if [ -z "$keywords" ]; then
            ./golamv2 --email --domains --url "$url" --workers 30 --depth 3 --dashboard 8080
        else
            ./golamv2 --email --keywords "$keywords" --domains --url "$url" --workers 30 --depth 3 --dashboard 8080
        fi
        ;;
    5)
        echo "Custom Configuration..."
        read -p "Enter starting URL: " url
        read -p "Enable email hunting? (y/n): " email_flag
        read -p "Enter keywords (or press Enter to skip): " keywords
        read -p "Enable dead link checking? (y/n): " domains_flag
        read -p "Number of workers (default 50): " workers
        read -p "Max depth (default 5): " depth
        read -p "Memory limit MB (default 500): " memory
        read -p "Dashboard port (default 8080): " port
        
        workers=${workers:-50}
        depth=${depth:-5}
        memory=${memory:-500}
        port=${port:-8080}
        
        cmd="./golamv2 --url \"$url\" --workers $workers --depth $depth --memory $memory --dashboard $port"
        
        if [ "$email_flag" = "y" ] || [ "$email_flag" = "Y" ]; then
            cmd="$cmd --email"
        fi
        
        if [ -n "$keywords" ]; then
            cmd="$cmd --keywords \"$keywords\""
        fi
        
        if [ "$domains_flag" = "y" ] || [ "$domains_flag" = "Y" ]; then
            cmd="$cmd --domains"
        fi
        
        echo "Executing: $cmd"
        eval $cmd
        ;;
    *)
        echo "Invalid choice. Exiting."
        exit 1
        ;;
esac

echo ""
echo "Crawling completed!"
echo "Check the dashboard at http://localhost:8080 for results"
echo "Results are stored in golamv2_data/ directory"

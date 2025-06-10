package infrastructure

import (
	"context"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"golamv2/internal/domain"
	"golamv2/pkg/metrics"

	"github.com/PuerkitoBio/goquery"
)

// ContentExtractor implements domain.ContentExtractor
type ContentExtractor struct {
	emailRegex      *regexp.Regexp
	httpClient      *http.Client
	deadLinkClient  *http.Client // Separate client with aggressive timeout for dead link checking
	mu              sync.RWMutex
	deadLinkCache   map[string]bool
	deadDomainCache map[string]bool // Cache for domain-level checks

	// Async dead link checking - results go directly to storage
	linkQueue chan linkCheckRequest
	storage   domain.Storage            // Direct access to storage for async updates
	metrics   *metrics.MetricsCollector // Direct access to metrics for updates
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

type linkCheckRequest struct {
	url       string
	sourceURL string
}

// NewContentExtractor creates a new content extractor
func NewContentExtractor() *ContentExtractor {
	ctx, cancel := context.WithCancel(context.Background())

	extractor := &ContentExtractor{
		emailRegex: regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 3 {
					return http.ErrUseLastResponse
				}
				return nil
			},
		},
		// Aggressive timeout for dead link checking
		deadLinkClient: &http.Client{
			Timeout: 2 * time.Second, // Very fast timeout for dead link checks
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse // Don't follow redirects for speed
			},
		},
		deadLinkCache:   make(map[string]bool),
		deadDomainCache: make(map[string]bool),
		linkQueue:       make(chan linkCheckRequest, 1000), // Buffered queue
		ctx:             ctx,
		cancel:          cancel,
	}

	// Start background workers for async dead link checking
	numWorkers := 3 // Reduced from 10 workers per page
	for i := 0; i < numWorkers; i++ {
		extractor.wg.Add(1)
		go extractor.asyncDeadLinkWorker()
	}

	return extractor
}

// SetStorage allows setting the storage reference after creation
func (e *ContentExtractor) SetStorage(storage domain.Storage) {
	e.storage = storage
}

// SetMetrics allows setting the metrics collector reference after creation
func (e *ContentExtractor) SetMetrics(metrics *metrics.MetricsCollector) {
	e.metrics = metrics
}

// extracts email addresses
func (e *ContentExtractor) ExtractEmails(content string) []string {
	matches := e.emailRegex.FindAllString(content, -1)

	// Deduplicate emails
	emailMap := make(map[string]bool)
	var emails []string

	for _, email := range matches {
		email = strings.ToLower(email)
		if !emailMap[email] {
			emailMap[email] = true
			emails = append(emails, email)
		}
	}

	return emails
}

// searches for specific keywords in content and counts occurrences
func (e *ContentExtractor) ExtractKeywords(content string, keywords []string) map[string]int {
	results := make(map[string]int)
	contentLower := strings.ToLower(content)

	for _, keyword := range keywords {
		keywordLower := strings.ToLower(keyword)
		count := strings.Count(contentLower, keywordLower)
		if count > 0 {
			results[keyword] = count
		}
	}

	return results
}

// extracts all links from HTML content
func (e *ContentExtractor) ExtractLinks(content, baseURL string) []string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return nil
	}

	baseU, err := url.Parse(baseURL)
	if err != nil {
		return nil
	}

	var links []string
	linkMap := make(map[string]bool)

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}

		// Resolve relative URLs
		linkURL, err := url.Parse(href)
		if err != nil {
			return
		}

		absoluteURL := baseU.ResolveReference(linkURL)
		urlStr := absoluteURL.String()

		// Filter valid URLs and deduplicate
		if domain.IsValidURL(urlStr) && !linkMap[urlStr] {
			linkMap[urlStr] = true
			links = append(links, urlStr)
		}
	})

	// Extract links from src attributes (images, scripts, etc.)
	doc.Find("[src]").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if !exists {
			return
		}

		srcURL, err := url.Parse(src)
		if err != nil {
			return
		}

		absoluteURL := baseU.ResolveReference(srcURL)
		urlStr := absoluteURL.String()

		if domain.IsValidURL(urlStr) && !linkMap[urlStr] {
			linkMap[urlStr] = true
			links = append(links, urlStr)
		}
	})

	return links
}

// extracts the page title from HTML content
func (e *ContentExtractor) ExtractTitle(content string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
	if err != nil {
		return ""
	}

	title := doc.Find("title").First().Text()
	return strings.TrimSpace(title)
}

// CheckDeadLinks queues links for async checking and returns empty results immediately
func (e *ContentExtractor) CheckDeadLinks(links []string, sourceURL string) ([]string, []string) {
	// Sample 20% of links for async processing
	sampledLinks := e.sampleLinks(links, 0.2)

	// Queue all sampled links for background processing
	e.queueLinksForChecking(sampledLinks, sourceURL)

	// Return empty results immediately - dead links will be stored in DB by async workers
	return []string{}, []string{}
}

// sampleLinks randomly selects a percentage of links
func (e *ContentExtractor) sampleLinks(links []string, percentage float64) []string {
	if percentage >= 1.0 {
		return links
	}

	numToSample := int(float64(len(links)) * percentage)
	if numToSample == 0 && len(links) > 0 {
		numToSample = 1 // Always sample at least 1 link if any exist
	}

	// Shuffle and take first N
	shuffled := make([]string, len(links))
	copy(shuffled, links)

	// Simple Fisher-Yates shuffle
	for i := len(shuffled) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled[:numToSample]
}

// queueLinksForChecking adds links to the async checking queue
func (e *ContentExtractor) queueLinksForChecking(links []string, sourceURL string) {
	for _, link := range links {
		select {
		case e.linkQueue <- linkCheckRequest{url: link, sourceURL: sourceURL}:
			// Successfully queued
		default:
			// Queue is full, skip this link
		}
	}
}

// asyncDeadLinkWorker processes links in the background
func (e *ContentExtractor) asyncDeadLinkWorker() {
	defer e.wg.Done()

	for {
		select {
		case <-e.ctx.Done():
			return
		case req := <-e.linkQueue:
			e.processLinkAsync(req)
		}
	}
}

// isDeadLinkFast checks if a link is dead with aggressive timeout (URL-level check)
func (e *ContentExtractor) isDeadLinkFast(urlStr string) bool {
	// Check cache first
	e.mu.RLock()
	if cached, exists := e.deadLinkCache[urlStr]; exists {
		e.mu.RUnlock()
		return cached
	}
	e.mu.RUnlock()

	// Use HEAD request only (no GET fallback for speed)
	req, err := http.NewRequest("HEAD", urlStr, nil)
	if err != nil {
		e.cacheDeadLink(urlStr, false)
		return false
	}
	req.Header.Set("User-Agent", "GolamV2-Crawler/1.0")

	resp, err := e.deadLinkClient.Do(req)
	if err != nil {
		// This could be domain-level or URL-level issue
		// We'll let the domain check handle domain-level issues
		e.cacheDeadLink(urlStr, true)
		return true
	}
	defer resp.Body.Close()

	// Only consider HTTP error status codes as dead (not connection issues)
	isDead := resp.StatusCode == 404 || resp.StatusCode == 410 || resp.StatusCode >= 500
	e.cacheDeadLink(urlStr, isDead)

	return isDead
}

// Close shuts down the async workers
func (e *ContentExtractor) Close() {
	e.cancel()
	close(e.linkQueue)
	e.wg.Wait()
}

// processLinkAsync checks if a link is dead and stores result directly in database
func (e *ContentExtractor) processLinkAsync(req linkCheckRequest) {
	if e.storage == nil {
		return // No storage available
	}

	// Extract domain first
	domainName := domain.GetDomain(req.url)
	if domainName == "" {
		return // Invalid URL
	}

	// Check if domain is dead first (optimization)
	isDomainDead := e.isDomainDead(domainName)
	if isDomainDead {
		// Domain is dead, so URL is automatically dead too
		result := domain.CrawlResult{
			URL:         req.sourceURL,
			ProcessedAt: time.Now(),
			DeadLinks:   []string{req.url},
			DeadDomains: []string{domainName},
		}

		e.storage.StoreResult(result)

		// Update metrics if available
		if e.metrics != nil {
			e.metrics.UpdateDeadLinksFound(1)
			e.metrics.UpdateDeadDomainsFound(1)
		}
		return
	}

	// Domain is alive, check specific URL
	isURLDead := e.isDeadLinkFast(req.url)
	if isURLDead {
		// URL is dead but domain is alive
		result := domain.CrawlResult{
			URL:         req.sourceURL,
			ProcessedAt: time.Now(),
			DeadLinks:   []string{req.url},
			DeadDomains: []string{}, // Domain is NOT dead
		}

		e.storage.StoreResult(result)

		// Update metrics if available
		if e.metrics != nil {
			e.metrics.UpdateDeadLinksFound(1)
			// Don't increment dead domains since domain is alive
		}
	}
}

// isDomainDead checks if an entire domain is unreachable (DNS/connection level)
func (e *ContentExtractor) isDomainDead(domainName string) bool {
	// Check cache first
	e.mu.RLock()
	if cached, exists := e.deadDomainCache[domainName]; exists {
		e.mu.RUnlock()
		return cached
	}
	e.mu.RUnlock()

	// Try to connect to domain root
	testURL := "https://" + domainName
	req, err := http.NewRequest("HEAD", testURL, nil)
	if err != nil {
		e.cacheDomainStatus(domainName, true)
		return true
	}
	req.Header.Set("User-Agent", "GolamV2-Crawler/1.0")

	resp, err := e.deadLinkClient.Do(req)
	if err != nil {
		// Connection failed - domain is likely dead
		e.cacheDomainStatus(domainName, true)
		return true
	}
	defer resp.Body.Close()

	// If we get any HTTP response, domain is alive
	e.cacheDomainStatus(domainName, false)
	return false
}

// cacheDomainStatus caches the domain alive/dead status
func (e *ContentExtractor) cacheDomainStatus(domainName string, isDead bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(e.deadDomainCache) > 1000 {
		e.deadDomainCache = make(map[string]bool)
	}

	e.deadDomainCache[domainName] = isDead
}

func (e *ContentExtractor) cacheDeadLink(urlStr string, isDead bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(e.deadLinkCache) > 5000 {
		e.deadLinkCache = make(map[string]bool)
	}

	e.deadLinkCache[urlStr] = isDead
}

package infrastructure

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"golamv2/internal/domain"

	"github.com/PuerkitoBio/goquery"
)

// ContentExtractor implements domain.ContentExtractor
type ContentExtractor struct {
	emailRegex    *regexp.Regexp
	httpClient    *http.Client
	mu            sync.RWMutex
	deadLinkCache map[string]bool
}

// NewContentExtractor creates a new content extractor
func NewContentExtractor() *ContentExtractor {
	return &ContentExtractor{
		emailRegex: regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		deadLinkCache: make(map[string]bool),
	}
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

// CheckDeadLinks checks which links are dead (return 404) and separates dead domains
func (e *ContentExtractor) CheckDeadLinks(links []string) ([]string, []string) {
	var deadLinks []string
	var deadDomains []string
	domainMap := make(map[string]bool)

	// Use channels for concurrent checking
	type linkResult struct {
		url    string
		isDead bool
		domain string
	}

	linkChan := make(chan string, len(links))
	resultChan := make(chan linkResult, len(links))

	// Start workers
	numWorkers := 10
	for i := 0; i < numWorkers; i++ {
		go func() {
			for link := range linkChan {
				isDead := e.isDeadLink(link)
				domain := domain.GetDomain(link)
				resultChan <- linkResult{
					url:    link,
					isDead: isDead,
					domain: domain,
				}
			}
		}()
	}

	// Send links to workers
	for _, link := range links {
		linkChan <- link
	}
	close(linkChan)

	// Collect results
	for i := 0; i < len(links); i++ {
		result := <-resultChan
		if result.isDead {
			deadLinks = append(deadLinks, result.url)
			if !domainMap[result.domain] {
				domainMap[result.domain] = true
				deadDomains = append(deadDomains, result.domain)
			}
		}
	}

	return deadLinks, deadDomains
}

// checks if a link returns 404 or is unreachable --- Slow response times can cause issues with false positives
func (e *ContentExtractor) isDeadLink(urlStr string) bool {
	e.mu.RLock()
	if cached, exists := e.deadLinkCache[urlStr]; exists {
		e.mu.RUnlock()
		return cached
	}
	e.mu.RUnlock()

	req, err := http.NewRequest("HEAD", urlStr, nil)
	if err != nil {
		e.cacheDeadLink(urlStr, false)
		return false
	}

	req.Header.Set("User-Agent", "GolamV2-Crawler/1.0")
	resp, err := e.httpClient.Do(req)
	if err != nil {
		e.cacheDeadLink(urlStr, false)
		return false
	}
	defer resp.Body.Close()

	isDead := resp.StatusCode == 404
	e.cacheDeadLink(urlStr, isDead)

	return isDead
}

func (e *ContentExtractor) cacheDeadLink(urlStr string, isDead bool) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if len(e.deadLinkCache) > 5000 {
		count := 0
		for k := range e.deadLinkCache {
			delete(e.deadLinkCache, k)
			count++
			if count >= 2500 {
				break
			}
		}
	}

	e.deadLinkCache[urlStr] = isDead
}

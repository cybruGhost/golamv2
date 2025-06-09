package domain

import (
	"net/url"
	"time"
)

// CrawlMode represents different crawling modes
type CrawlMode string

const (
	ModeEmail    CrawlMode = "email"
	ModeDomains  CrawlMode = "domains"
	ModeKeywords CrawlMode = "keywords"
	ModeAll      CrawlMode = "all"
)

// URLTask represents a URL to be crawled
type URLTask struct {
	URL       string    `json:"url"`
	Depth     int       `json:"depth"`
	Timestamp time.Time `json:"timestamp"`
	Retries   int       `json:"retries"`
}

// represents the result of crawling a URL
type CrawlResult struct {
	URL         string         `json:"url"`
	StatusCode  int            `json:"status_code"`
	Title       string         `json:"title"`
	Emails      []string       `json:"emails,omitempty"`
	Keywords    map[string]int `json:"keywords,omitempty"`
	DeadLinks   []string       `json:"dead_links,omitempty"`
	DeadDomains []string       `json:"dead_domains,omitempty"`
	NewURLs     []string       `json:"new_urls,omitempty"`
	ProcessedAt time.Time      `json:"processed_at"`
	ProcessTime time.Duration  `json:"process_time"`
	Error       string         `json:"error,omitempty"`
}

// represents crawler performance metrics
type CrawlMetrics struct {
	URLsProcessed    int64     `json:"urls_processed"`
	URLsInQueue      int64     `json:"urls_in_queue"`
	URLsInDB         int64     `json:"urls_in_db"`
	EmailsFound      int64     `json:"emails_found"`
	KeywordsFound    int64     `json:"keywords_found"`
	DeadLinksFound   int64     `json:"dead_links_found"`
	DeadDomainsFound int64     `json:"dead_domains_found"`
	ActiveWorkers    int       `json:"active_workers"`
	MemoryUsageMB    float64   `json:"memory_usage_mb"`
	URLsPerSecond    float64   `json:"urls_per_second"`
	StartTime        time.Time `json:"start_time"`
	LastUpdateTime   time.Time `json:"last_update_time"`
	Errors           int64     `json:"errors"`
	// Memory breakdown by component
	MemoryBreakdown MemoryBreakdown `json:"memory_breakdown"`
}

// MemoryBreakdown represents memory usage by component -- Something is off though not much of a breakdown-may cause an iinflated memory usage in the dashboard
type MemoryBreakdown struct {
	BloomFilterMB float64 `json:"bloom_filter_mb"`
	DatabaseMB    float64 `json:"database_mb"`
	QueueMB       float64 `json:"queue_mb"`
	HTTPBuffersMB float64 `json:"http_buffers_mb"`
	ParsingMB     float64 `json:"parsing_mb"`
	CrawlersMB    float64 `json:"crawlers_mb"`
	OtherMB       float64 `json:"other_mb"`
	TotalMB       float64 `json:"total_mb"`
}

// interface for the efficient URL queue
type URLQueue interface {
	Push(task URLTask) error
	Pop() (URLTask, error)
	Size() int
	IsFull() bool
	IsEmpty() bool
	Close() error
}

// BloomFilter
type BloomFilter interface {
	Add(url string)
	Test(url string) bool
	EstimateCount() uint64
	Reset()
}

// Storage interface for persistent storage
type Storage interface {
	StoreURL(task URLTask) error
	GetURLs(limit int) ([]URLTask, error)
	StoreResult(result CrawlResult) error
	GetResults(mode CrawlMode, limit int) ([]CrawlResult, error)
	GetMetrics() (*CrawlMetrics, error)
	UpdateMetrics(metrics *CrawlMetrics) error
	Close() error
}

// RobotsChecker interface for robots.txt compliance
type RobotsChecker interface {
	CanFetch(userAgent, urlStr string) bool
	GetSitemaps(domain string) []string
	GetCrawlDelay(userAgent, domain string) time.Duration
}

// ContentExtractor interface for extracting data from HTML
type ContentExtractor interface {
	ExtractEmails(content string) []string
	ExtractKeywords(content string, keywords []string) map[string]int
	ExtractLinks(content, baseURL string) []string
	ExtractTitle(content string) string
	CheckDeadLinks(links []string) ([]string, []string) // deadLinks, deadDomains
}

// IsValidURL checks if a URL is valid
func IsValidURL(urlStr string) bool {
	if urlStr == "" {
		return false
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	return u.Scheme == "http" || u.Scheme == "https"
}

// GetDomain extracts domain from URL
func GetDomain(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}
	return u.Host
}

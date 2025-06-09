package infrastructure

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/temoto/robotstxt"
)

// RobotsChecker implements domain.RobotsChecker
type RobotsChecker struct {
	mu        sync.RWMutex
	cache     map[string]*robotstxt.RobotsData
	client    *http.Client
	userAgent string
}

// NewRobotsChecker creates a new robots.txt checker
func NewRobotsChecker(userAgent string) *RobotsChecker {
	return &RobotsChecker{
		cache:     make(map[string]*robotstxt.RobotsData),
		userAgent: userAgent,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CanFetch checks if the given URL can be fetched according to robots.txt
func (r *RobotsChecker) CanFetch(userAgent, urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	domain := u.Host
	robots := r.getRobots(domain)
	if robots == nil {
		return true // If we can't get robots.txt, assume we good!
	}

	group := robots.FindGroup(userAgent)
	if group == nil {
		group = robots.FindGroup("*")
	}

	if group == nil {
		return true
	}

	return group.Test(u.Path)
}

// GetSitemaps returns sitemap URLs from robots.txt
func (r *RobotsChecker) GetSitemaps(domain string) []string {
	robots := r.getRobots(domain)
	if robots == nil {
		return nil
	}

	var sitemaps []string
	for _, sitemap := range robots.Sitemaps {
		sitemaps = append(sitemaps, sitemap)
	}

	return sitemaps
}

// GetCrawlDelay returns the crawl delay for the given user agent and domain
func (r *RobotsChecker) GetCrawlDelay(userAgent, domain string) time.Duration {
	robots := r.getRobots(domain)
	if robots == nil {
		return 0
	}

	group := robots.FindGroup(userAgent)
	if group == nil {
		group = robots.FindGroup("*")
	}

	if group == nil {
		return 0
	}

	return time.Duration(group.CrawlDelay) * time.Second
}

// getRobots fetches and caches robots.txt for a domain
func (r *RobotsChecker) getRobots(domain string) *robotstxt.RobotsData {
	r.mu.RLock()
	if robots, exists := r.cache[domain]; exists {
		r.mu.RUnlock()
		return robots
	}
	r.mu.RUnlock()

	// Fetch robots.txt
	robotsURL := fmt.Sprintf("https://%s/robots.txt", domain)
	resp, err := r.client.Get(robotsURL)
	if err != nil {
		// Try HTTP if HTTPS fails
		robotsURL = fmt.Sprintf("http://%s/robots.txt", domain)
		resp, err = r.client.Get(robotsURL)
		if err != nil {
			r.cacheRobots(domain, nil)
			return nil
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		r.cacheRobots(domain, nil)
		return nil
	}

	robots, err := robotstxt.FromResponse(resp)
	if err != nil {
		r.cacheRobots(domain, nil)
		return nil
	}

	r.cacheRobots(domain, robots)
	return robots
}

// cacheRobots caches robots.txt data for a domain
func (r *RobotsChecker) cacheRobots(domain string, robots *robotstxt.RobotsData) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache[domain] = robots
}

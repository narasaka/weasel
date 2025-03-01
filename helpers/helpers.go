package helpers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

type Link struct {
	url        *url.URL
	statusCode int
	err        error
	parent     string
	traversed  bool
}

type StatusMap struct {
	mu     sync.Mutex
	store  map[string]Link
	errors int
}

func (s *StatusMap) set(key string, link Link) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[key] = link
	if link.err != nil {
		s.errors++
	}
}

func (s *StatusMap) get(key string) (Link, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	status, exists := s.store[key]
	return status, exists
}

var statusMap StatusMap = StatusMap{store: make(map[string]Link)}

func isSameDomain(baseURL, checkURL *url.URL) bool {
	return baseURL.Hostname() == checkURL.Hostname()
}

func extractLinks(body io.Reader, baseURL *url.URL) ([]string, error) {
	links := []string{}
	tokenizer := html.NewTokenizer(body)

	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			err := tokenizer.Err()

			// end of document
			if err == io.EOF {
				return links, nil
			}
			return nil, fmt.Errorf("error tokenizing HTML: %w", err)
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokenizer.Token()
			if token.Data == "a" {
				for _, attr := range token.Attr {
					if attr.Key == "href" {
						absoluteURL, err := url.Parse(attr.Val)

						// invalid url handling
						if err != nil {
							fmt.Fprintf(os.Stderr, "Invalid URL: %s - %v\n", attr.Val, err)
							continue
						}

						// valid url handling
						absoluteURL = baseURL.ResolveReference(absoluteURL)
						if absoluteURL.Scheme != "mailto" {
							links = append(links, absoluteURL.String())
						}
					}
				}
			}
		}
	}
}

func getLinkStatus(link *url.URL, sourcePage string) Link {
	if link.Scheme == "mailto" {
		return Link{url: link, statusCode: http.StatusOK, err: nil, parent: sourcePage}
	}
	resp, err := http.Get(link.String())
	if err != nil {
		return Link{url: link, statusCode: 0, err: err, parent: sourcePage}
	}
	return Link{url: link, statusCode: resp.StatusCode, err: nil, parent: sourcePage}
}

// i want `example.com` to be treated the same as `example.com/`.
// trimming the leading and trailing spaces too because i've seen quite a few like: ` example.com/`
func normalizeURL(u string) string {
	if idx := strings.Index(u, "#"); idx != -1 {
		u = u[:idx]
	}
	if strings.HasSuffix(u, "/") {
		normalized := strings.TrimSuffix(strings.TrimSpace(u), "/")
		return normalized
	}
	return strings.TrimSpace(u)
}

// dfs but on a webpage
func traverse(targetURL *url.URL, recursive bool) {
	normalizedTargetURL := normalizeURL(targetURL.String())

	// if target is seen, skip it
	if currentPage, exists := statusMap.get(normalizedTargetURL); exists && currentPage.traversed {
		fmt.Printf("%-70s [SEEN, SKIPPING]\n", normalizedTargetURL)
		return
	}

	// avoid 429 status codes
	time.Sleep(100 * time.Millisecond)

	fmt.Printf("\nChecking links on %s\n", normalizedTargetURL)

	resp, err := http.Get(normalizedTargetURL)
	if err != nil {
		log.Printf("error fetching target url (%v): %v", normalizedTargetURL, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("error: http status code: %d\n", resp.StatusCode)
		return
	}

	statusMap.set(normalizedTargetURL, Link{
		url:        targetURL,
		statusCode: resp.StatusCode,
		err:        nil,
		traversed:  true,
	})

	// extract the links (<a> tags only for now) from the page
	// and mark it as "seen"
	links, err := extractLinks(resp.Body, targetURL)
	if err != nil {
		log.Printf("error extracting links: %v\n", err)
		return
	}

	if len(links) == 0 {
		fmt.Println("No links to check")
		return
	}

	// list of links to check only including those not already seen
	var linksToCheck []*url.URL
	for _, link := range links {
		normalizedLink := normalizeURL(link)
		if _, exists := statusMap.get(normalizedLink); !exists {
			lURL, err := url.Parse(link)
			if err != nil {
				log.Printf("error parsing url: %v", err)
				continue
			}
			linksToCheck = append(linksToCheck, lURL)
		}
	}

	var checkWG sync.WaitGroup
	for _, link := range linksToCheck {
		checkWG.Add(1)
		go func(l *url.URL) {
			defer checkWG.Done()
			normalizedLink := normalizeURL(l.String())
			if _, exists := statusMap.get(normalizedLink); exists {
				fmt.Printf("%-70s [SEEN, SKIPPING]\n", normalizedTargetURL)
				return
			}
			status := getLinkStatus(l, normalizedTargetURL)
			statusMap.set(normalizedLink, status)
			if status.err != nil {
				fmt.Printf("%-70s [ERROR: %v]\n", status.url, status.err)
			} else if status.statusCode != http.StatusOK {
				fmt.Printf("%-70s [ERROR: Status %d]\n", status.url, status.statusCode)
			} else {
				fmt.Printf("%-70s [OK]\n", status.url)
			}
		}(link)
	}
	checkWG.Wait()

	if recursive {
		for _, link := range linksToCheck {
			if isSameDomain(targetURL, link) {
				traverse(link, recursive)
			}
		}
	}
}

func Check(target string, recursive bool) {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("error parsing base url: %v\n", err)
	}

	traverse(targetURL, recursive)

	if statusMap.errors > 0 {
		fmt.Printf("\n=== Summary of All Problematic Links (%d) ===\n", statusMap.errors)
		for _, link := range statusMap.store {
			if link.err != nil {
				fmt.Printf("%-70s [ERROR: %v] (found on %s)\n", link.url, link.err, link.parent)
			} else {
				fmt.Printf("%-70s [Status: %d] (found on %s)\n", link.url, link.statusCode, link.parent)
			}
		}
	} else {
		fmt.Println("\nNo broken links found!")
	}
}

package helpers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"golang.org/x/net/html"
)

var seen map[string]bool

type LinkStatus struct {
	url        string
	statusCode int
	err        error
}

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

func getLinkStatus(link string) LinkStatus {
	parsedURL, err := url.Parse(link)
	if err != nil {
		return LinkStatus{url: link, statusCode: 0, err: err}
	}
	if parsedURL.Scheme == "mailto" {
		return LinkStatus{url: link, statusCode: http.StatusOK, err: nil}
	}
	resp, err := http.Get(link)
	if err != nil {
		return LinkStatus{url: link, statusCode: 0, err: err}
	}
	defer resp.Body.Close()
	return LinkStatus{url: link, statusCode: resp.StatusCode, err: nil}
}

// i want `example.com` to be treated the same as `example.com/`
func normalizeURL(u string) string {
	if strings.HasSuffix(u, "/") {
		return strings.TrimSuffix(u, "/")
	}
	return u
}

// dfs but on a webpage
func traverse(target string, baseURL *url.URL, recursive bool) {
	normalizedTarget := normalizeURL(target)
	if seen[normalizedTarget] {
		return
	}
	seen[normalizedTarget] = true

	resp, err := http.Get(target)
	if err != nil {
		log.Printf("error fetching target url (%v): %v", target, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("error: http status code: %d\n", resp.StatusCode)
		return
	}

	links, err := extractLinks(resp.Body, baseURL)
	if err != nil {
		log.Printf("error extracting links: %v\n", err)
		return
	}

	var problematicLinks []LinkStatus
	fmt.Printf("\nChecking links on %s\n", target)

	results := make(chan LinkStatus, len(links))
	for _, link := range links {
		if !seen[link] {
			go func(l string) {
				results <- getLinkStatus(l)
			}(link)
		}
	}

	uniqueLinks := 0
	for _, link := range links {
		if !seen[link] {
			uniqueLinks++
			status := <-results
			if status.err != nil {
				fmt.Printf("%-70s [ERROR: %v]\n", status.url, status.err)
				problematicLinks = append(problematicLinks, status)
			} else if status.statusCode != http.StatusOK {
				fmt.Printf("%-70s [ERROR: Status %d]\n", status.url, status.statusCode)
				problematicLinks = append(problematicLinks, status)
			} else {
				fmt.Printf("%-70s [OK]\n", status.url)
			}

			// handle recursion
			if recursive {
				linkURL, err := url.Parse(link)
				if err == nil && isSameDomain(baseURL, linkURL) {
					traverse(link, baseURL, recursive)
				}
			}
		}
	}

	if len(problematicLinks) > 0 {
		fmt.Printf("\n=== Summary of Problematic Links for %s (%d) ===\n", target, len(problematicLinks))
		for _, bad := range problematicLinks {
			if bad.err != nil {
				fmt.Printf("%-70s [ERROR: %v]\n", bad.url, bad.err)
			} else {
				fmt.Printf("%-70s [Status: %d]\n", bad.url, bad.statusCode)
			}
		}
	}
}

func Check(target string, recursive bool) {
	seen = make(map[string]bool)
	baseURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("error parsing base url: %v\n", err)
	}
	traverse(target, baseURL, recursive)
}

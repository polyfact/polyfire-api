package webrequest

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"

	"github.com/cixtor/readability"
	"github.com/gocolly/colly/v2"
	"github.com/tmc/langchaingo/llms"
)

const (
	baseUrl              = "https://html.duckduckgo.com/html/?q=%s&no_redirect=1"
	duckDuckGoPrefix     = "//duckduckgo.com/l/?uddg="
	maxSitesToVisit      = 7
	answerTokenLength    = 300
	additionalTokenSpace = 300
	urlPattern           = `read:(https?://[^\s]+)` // Extract URL with "read:" prefix.
)

var (
	WebsiteExceedsLimit    error = errors.New("error_website_exceeds_limit")
	WebsitesContentExceeds error = errors.New("error_websites_content_exceeds")
	NoContentFound         error = errors.New("error_no_content_found")
	FetchWebpageError      error = errors.New("error_fetch_webpage")
	ParseContentError      error = errors.New("error_parse_content")
	VisitBaseURLError      error = errors.New("error_visit_base_url")
)

func prepareURL(u string) (string, error) {
	if strings.HasPrefix(u, duckDuckGoPrefix) {
		rawURL := strings.TrimPrefix(u, duckDuckGoPrefix)
		decodedURL, err := url.QueryUnescape(rawURL)
		if err != nil {
			return "", err
		}
		return strings.Split(decodedURL, "&rut=")[0], nil
	}
	return u, nil
}

func removeUselessWhitespaces(s string) string {
	re := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
	return re.ReplaceAllString(strings.TrimSpace(s), " ")
}

func fetchContent(link string) (string, error) {
	res, err := http.Get(link)
	if err != nil {
		return "", FetchWebpageError
	}
	defer res.Body.Close()

	r := readability.New()
	parsed, err := r.Parse(res.Body, link)
	if err != nil {
		return "", ParseContentError
	}

	return removeUselessWhitespaces(parsed.TextContent), nil
}

func containsURL(content string) ([]string, bool) {
	re := regexp.MustCompile(urlPattern)
	matches := re.FindAllStringSubmatch(content, -1)

	var urls []string
	for _, match := range matches {
		if len(match) > 1 {
			urls = append(urls, match[1])
		}
	}

	if len(urls) > 0 {
		return urls, true
	}
	return nil, false
}

func WebRequest(query string, model string) (string, error) {
	var accumulatedText strings.Builder

	contextSize := llms.GetModelContextSize(model) - answerTokenLength

	urlsFound, ok := containsURL(query)
	if ok {
		for _, urlFound := range urlsFound {
			content, err := fetchContent(urlFound)
			if err != nil {
				fmt.Println()
				continue
			}

			contentTokenCount := llms.CountTokens(model, content)
			if contentTokenCount > contextSize {
				return "", WebsiteExceedsLimit
			}

			totalTokens := llms.CountTokens(model, accumulatedText.String()+content)
			if totalTokens > contextSize {
				return "", WebsitesContentExceeds
			}

			accumulatedText.WriteString(content + "\n==========\n")
		}

		if accumulatedText.Len() == 0 {
			return "", NoContentFound
		}
		return accumulatedText.String(), nil
	}

	c := colly.NewCollector()
	var sitesVisited int
	var wg sync.WaitGroup

	c.OnHTML(".result", func(e *colly.HTMLElement) {
		wg.Add(1)
		defer wg.Done()

		if sitesVisited >= maxSitesToVisit {
			return
		}

		linkAttr := e.ChildAttr(".result__title .result__a", "href")
		link, err := prepareURL(linkAttr)
		if err != nil {
			fmt.Println("Error preparing URL:", err)
			return
		}

		content, err := fetchContent(link)
		if err != nil {
			fmt.Println("Error fetching content:", err)
			return
		}

		formattedContent := fmt.Sprintf("Site %d (%s): %s\n==========\n", sitesVisited+1, link, content)

		totalTokens := llms.CountTokens(model, accumulatedText.String()+formattedContent)

		if totalTokens <= contextSize && totalTokens+additionalTokenSpace <= contextSize {
			accumulatedText.WriteString(formattedContent)
			sitesVisited++
		}
	})

	err := c.Visit(fmt.Sprintf(baseUrl, url.QueryEscape(query)))
	if err != nil {
		return "", VisitBaseURLError
	}

	wg.Wait()

	if accumulatedText.Len() == 0 {
		return "", NoContentFound
	}

	return accumulatedText.String(), nil
}

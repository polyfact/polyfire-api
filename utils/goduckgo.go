package utils

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

var baseUrl = "https://html.duckduckgo.com/html/?q=%s&no_redirect=1"

func prepareURL(u string) string {
	prefix := "//duckduckgo.com/l/?uddg="
	if strings.HasPrefix(u, prefix) {
		rawURL := strings.TrimPrefix(u, prefix)
		decodedURL, err := url.QueryUnescape(rawURL)
		if err != nil {
			return u
		}
		finalURL := strings.Split(decodedURL, "&rut=")[0]
		return finalURL
	}
	return u
}
func removeUselessWhitespaces(s string) string {
	s = strings.TrimSpace(s)

	re := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
	return re.ReplaceAllString(s, " ")
}

func WebRequest(query string, model string) (string, error) {
	c := colly.NewCollector()
	var accumulatedText strings.Builder
	var maxSitesToVisit = 7
	var sitesVisited = 0

	var wg sync.WaitGroup

	c.OnHTML(".result", func(e *colly.HTMLElement) {

		wg.Add(1)
		defer wg.Done()

		if !(sitesVisited < maxSitesToVisit) {
			return
		}

		link := e.ChildAttr(".result__title .result__a", "href")

		res, err := http.Get(prepareURL(link))

		if err != nil {
			fmt.Println("Failed to fetch the webpage:", err)
			return
		}

		defer res.Body.Close()

		r := readability.New()
		parsed, err := r.Parse(res.Body, link)

		if err != nil {
			fmt.Println("Failed to parse the content with readability:", err)
			return
		}

		content := removeUselessWhitespaces(parsed.TextContent)

		content = fmt.Sprintf("Site %d: %s\n==========\n", sitesVisited+1, content)

		totalTokens := llms.CountTokens(model, accumulatedText.String()+content)
		contextSize := llms.GetModelContextSize(model) - 500 // 500 is token length for the answer

		if totalTokens > contextSize {
			return
		} else if totalTokens+300 > contextSize {
			return
		}

		accumulatedText.WriteString(content)
		sitesVisited++

	})

	err := c.Visit(fmt.Sprintf(baseUrl, url.QueryEscape(query)))
	if err != nil {
		return "", fmt.Errorf("Error visiting base URL: %v", err)
	}

	wg.Wait()

	if accumulatedText.Len() == 0 {
		return "", errors.New("No visible text accumulated from results")
	}

	return accumulatedText.String(), nil
}

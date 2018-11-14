package main

import (
	"encoding/xml"
	"flag"
	"html-link-parser-gophercies/link"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const xmlns = "https://www.sitempas.org/schemas/sitemap/0.9"

type loc struct {
	Value string `xml:"loc"`
}

type urlset struct {
	Urls  []loc  `xml:"url"`
	Xmlns string `xml:"xmlns,attr"`
}

func main() {
	urlFlag := flag.String("url", "https://gophercises.com", "the url that you want to build a sitemap for")
	maxDepthFlag := flag.Int("depth", 3, "the maximum number of links deep to traverse")
	flag.Parse()

	// pages := requestPage(*urlFlag)
	pages := bfs(*urlFlag, *maxDepthFlag)

	toXml := urlset{
		Xmlns: xmlns,
	}
	for _, page := range pages {
		toXml.Urls = append(toXml.Urls, loc{page})
	}

	enc := xml.NewEncoder(os.Stdout)
	enc.Indent("", "  ")
	if err := enc.Encode(toXml); err != nil {
		panic(err)
	}

}

func bfs(url string, maxDepth int) []string {
	urlVisited := map[string]bool{}
	urlQueue := map[string]bool{}
	nextUrlQueue := map[string]bool{
		url: true,
	}
	for i := 0; i <= maxDepth; i++ {
		urlQueue, nextUrlQueue = nextUrlQueue, make(map[string]bool)
		if len(urlQueue) == 0 {
			break
		}
		for url := range urlQueue {
			_, ok := urlVisited[url]
			if ok {
				continue
			}
			urlVisited[url] = true
			hrefs := requestPage(url)
			for _, link := range hrefs {
				if _, ok := urlVisited[link]; !ok {
					nextUrlQueue[link] = true
				}
				nextUrlQueue[link] = true
			}
		}
	}

	listUrls := make([]string, 0, len(urlVisited))

	for url, _ := range urlVisited {
		listUrls = append(listUrls, url)
	}

	return listUrls
}

func gethrefs(r io.Reader, base string) []string {
	links, _ := link.Parse(r)

	var hrefs []string
	for _, l := range links {
		switch {
		// handle /something
		case strings.HasPrefix(l.Href, "/"):
			hrefs = append(hrefs, base+l.Href)
		// handle https/http
		case strings.HasPrefix(l.Href, "http"):
			hrefs = append(hrefs, l.Href)
		}
	}
	return hrefs
}

func requestPage(urlStr string) []string {
	resp, err := http.Get(urlStr)
	if err != nil {
		return []string{}
	}
	defer resp.Body.Close()

	reqUrl := resp.Request.URL

	baseUrl := &url.URL{
		Scheme: reqUrl.Scheme,
		Host:   reqUrl.Host,
	}

	base := baseUrl.String()
	hrefs := gethrefs(resp.Body, base)
	filterResultFn := filterWithPrefix(base)
	return filter(hrefs, filterResultFn)
}

func filter(links []string, filterPrefix func(string) bool) []string {
	var hrefs []string

	for _, link := range links {
		filterPrefixValue := filterPrefix(link)
		if filterPrefixValue {
			hrefs = append(hrefs, link)
		}
	}
	return hrefs

}

func filterWithPrefix(prfx string) func(string) bool {
	return func(link string) bool {
		return strings.HasPrefix(link, prfx)
	}
}

package main

import (
	"flag"
	"fmt"
	"html-link-parser-gophercies/link"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func main() {
	urlFlag := flag.String("url", "https://gophercises.com", "the url that you want to build a sitemap for")
	flag.Parse()

	pages := requestPage(*urlFlag)

	for _, p := range pages {
		fmt.Println(p)
	}

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
		panic(err)
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

package parser

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/yanzay/log"

	"golang.org/x/net/html"
)

var retre = regexp.MustCompile(`location.replace\("(.*)"\)`)

func (p *Parser) GetLink(serial, season, episode int) (string, error) {
	retreLink, err := p.linkToRetre(serial, season, episode)
	if err != nil {
		return "", err
	}
	tracktorLink, err := p.linkToTracktor(retreLink)
	if err != nil {
		return "", err
	}
	return tracktorLink, nil
}

func (p *Parser) linkToRetre(serial, season, episode int) (string, error) {
	lostLink := fmt.Sprintf("http://lostfilm.tv/nrdr2.php?c=%d&s=%d&e=%d", serial, season, episode)
	req, err := http.NewRequest("GET", lostLink, nil)
	if err != nil {
		return "", err
	}
	req.AddCookie(&http.Cookie{Name: "uid", Value: p.userID})
	req.AddCookie(&http.Cookie{Name: "pass", Value: p.password})
	resp, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	content, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return "", err
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf(string(content))
	}
	matches := retre.FindStringSubmatch(string(content))
	if len(matches) < 2 {
		return "", fmt.Errorf("Can't find link to retre.org:\n%s", string(content))
	}
	return matches[1], nil
}

func (p *Parser) linkToTracktor(retreLink string) (string, error) {
	resp, err := p.client.Get(retreLink)
	if err != nil {
		return "", err
	}
	doc, err := html.Parse(resp.Body)
	if err != nil {
		return "", err
	}
	var f func(n *html.Node, quality string) string
	f = func(n *html.Node, quality string) string {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, a := range n.Attr {
				if a.Key == "href" {
					linkText := n.FirstChild.Data
					log.Tracef("%s: %s", linkText, a.Val)
					if strings.Contains(linkText, quality) {
						return a.Val
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			link := f(c, quality)
			if link != "" {
				return link
			}
		}
		return ""
	}
	link := f(doc, "1080p")
	if link != "" {
		return link, nil
	}
	link = f(doc, "720p")
	if link != "" {
		return link, nil
	}
	writer := &bytes.Buffer{}
	html.Render(writer, doc)
	return "", fmt.Errorf("Can't get link to tracktor from retre link: %s, %v", retreLink, writer.String())
}

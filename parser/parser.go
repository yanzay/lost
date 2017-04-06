package parser

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/yanzay/log"
	"github.com/yanzay/lost"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

const FULL_SEASON_NUMBER = 99

var rx = regexp.MustCompile(`ShowAllReleases\('\d*','(\d*)','([\d\-]*)'\)`)

type Parser struct {
	userID   string
	password string
	client   *http.Client
}

func NewParser(userID, password string) *Parser {
	p := &Parser{}
	p.userID = userID
	p.password = password
	p.client = &http.Client{}
	return p
}

func (p *Parser) ListAllEpisodes(id int) ([]*lost.Episode, error) {
	browseLink := fmt.Sprintf("http://www.lostfilm.tv/browse.php?cat=%d", id)
	doc, err := goquery.NewDocument(browseLink)
	if err != nil {
		return nil, err
	}
	episodes := make([]*lost.Episode, 0)
	doc.Find("div.mid div.t_row tbody").Each(func(i int, s *goquery.Selection) {
		episode, err := parseEpisode(s)
		if err != nil {
			log.Error(err)
		} else {
			if episode != nil {
				episodes = append(episodes, episode)
			}
		}
	})
	return episodes, nil
}

func parseEpisode(s *goquery.Selection) (*lost.Episode, error) {
	date := s.Find("span.micro").Children().First().Text()
	parsedDate, err := time.Parse("02.01.2006", date)
	if err != nil {
		return nil, fmt.Errorf("Can't parse date: %s", err)
	}
	onclick, ok := s.Find("td.t_episode_title").Attr("onclick")
	if !ok {
		return nil, fmt.Errorf("onclick not found")
	}
	log.Debug(onclick)
	match := rx.FindStringSubmatch(onclick)
	if len(match) < 3 {
		return nil, fmt.Errorf("Can't match episode number: %s", onclick)
	}
	seasonNum, err := strconv.Atoi(match[1])
	if err != nil {
		return nil, err
	}
	episodeNum, err := strconv.Atoi(match[2])
	if err != nil {
		return nil, err
	}
	// Full season, skip
	if episodeNum == FULL_SEASON_NUMBER {
		return nil, nil
	}
	title := cp1251toUTF8(s.Find("td.t_episode_title nobr span").Text())
	episode := &lost.Episode{
		Season: seasonNum,
		Number: episodeNum,
		Date:   parsedDate,
		Name:   title,
	}
	return episode, nil
}

func episodeFromMatch(match string) (int, error) {
	if strings.Contains(match, "-") {
		return strconv.Atoi(strings.Split(match, "-")[0])
	}
	return strconv.Atoi(match)
}

func cp1251toUTF8(str string) string {
	sr := strings.NewReader(str)
	tr := transform.NewReader(sr, charmap.Windows1251.NewDecoder())
	buf, err := ioutil.ReadAll(tr)
	if err != nil {
		log.Error(err)
	}

	return strings.TrimSpace(string(buf))
}

package crawler

import (
	"GoCrawl/internal/crypto"
	"GoCrawl/internal/log"
	"GoCrawl/internal/redis"
	"GoCrawl/internal/robot"
	"GoCrawl/internal/utils"
	"GoCrawl/model"
	"bytes"
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"math"
	"os"
	"strings"
	"syscall"
	"time"
)

var err error

const CRAWLER_RETRY = 3
const CRAWLER_BLOOM_KEY = "crawler:bloom"
const CRAWLER_TASK_QUEUE = "crawler:tasks"
const CRAWLER_RESULTS_QUEUE = "crawler:results"
const CRAWLER_HOST_COUNTER = "crawler:host:counter"
const CRAWLER_REQUEST_HOSTNAME_STATS = "crawler:request:hostname:stats"

var Sigchan = make(chan os.Signal, 1)

func Exit() {
	Sigchan <- syscall.SIGTERM
}

func recoverFromPanic() {
	if r := recover(); r != nil {
		log.Warn("Recovered from panic: %s", r)
		return
	}
}

func init() {
	if err != nil {
		panic(err)
	}

	reserveBloomKey()
}

func Scrape(workerCount int) {

	for i := 0; i < workerCount; i++ {
		log.Info("Starting worker %d", i)

		go func(id int) {
			for {
				scrape()
			}
		}(i)
	}

	select {}
}

func scrape() {
	defer recoverFromPanic()

	url := redis.LPop(CRAWLER_TASK_QUEUE)

	if len(url) == 0 {
		log.Warn("No URL in the queue")
		time.Sleep(5 * time.Second)
		return
	}

	var hostname = utils.GetHostname(url)

	if !utils.IsValidHostname(hostname) {
		log.Error("illegal url %s, hostname %s", url, hostname)
		return
	}

	robotTxt := robot.New(url)

	if !robotTxt.AgentAllowed("GoogleBot", url) {
		log.Error("URL is not allowed to visit in robots.txt. URL: %s", url)
		return
	}

	var c = colly.NewCollector(
		colly.UserAgent(os.Getenv("GO_BOT_UA")),
		colly.AllowURLRevisit(),
	)
	collySetupRetry(c, CRAWLER_RETRY)

	c.Limit(&colly.LimitRule{
		RandomDelay: 2 * time.Second,
	})

	c.OnRequest(func(r *colly.Request) {
		log.Info("Visiting %s", r.URL.String())
	})

	c.OnResponse(func(r *colly.Response) {
		if err = redis.HIncryBy(CRAWLER_REQUEST_HOSTNAME_STATS, hostname, 1); err != nil {
			log.Error(err.Error())
		}

		content := parseCollyResponseToModel(r)
		content.Domain = hostname
		content.URL = r.Request.URL.String()

		b, err := json.Marshal(content)
		if err != nil {
			log.Error("failed to marshal struct to json, url: %s", content.URL)
			return
		}

		if err := redis.LPush(CRAWLER_RESULTS_QUEUE, b); err != nil {
			log.Error(err.Error())
		}
	})

	c.OnHTML("a", func(e *colly.HTMLElement) {
		linkFinder(e, url)
	})

	c.Visit(url)
}

func collySetupRetry(c *colly.Collector, maxRetries int) {
	c.OnError(func(r *colly.Response, err error) {
		retriesLeft := maxRetries

		if retries, ok := r.Ctx.GetAny("retriesLeft").(int); ok {
			retriesLeft = retries
		}

		log.Error("error %s |  retriesLeft %d", err.Error(), retriesLeft)

		if retriesLeft > 0 {

			r.Ctx.Put("retriesLeft", retriesLeft-1)
			time.Sleep(time.Duration(math.Exp(float64(CRAWLER_RETRY-retriesLeft+1))) * time.Second)

			err := r.Request.Retry()
			if err != nil {
				return
			}
		} else {
			log.Error("Max retries reached for %s", r.Request.URL.String())
		}
	})
}

func linkFinder(e *colly.HTMLElement, currentUrl string) {

	link := e.Request.AbsoluteURL(e.Attr("href"))

	currentHostname := utils.GetHostname(currentUrl)
	linkHostname := utils.GetHostname(link)

	if len(link) == 0 {
		return
	}

	if !utils.IsValidHostname(linkHostname) {
		return
	}

	link = utils.UrlCleaner(link)

	bloomKey := crypto.Md5(strings.Replace(strings.TrimSuffix(link, "/"), "http://", "https://", 1))
	ok, err := redis.BloomAdd(CRAWLER_BLOOM_KEY, bloomKey)

	if err != nil {
		log.Error(err.Error())
		return
	}

	if !ok {
		log.Debug("URL has been crawled already, %s", link)
		return
	}

	log.Info("Found new Url %s", link)

	if linkHostname == currentHostname {
		err := redis.RPush(CRAWLER_TASK_QUEUE, link)

		if err != nil {
			log.Error(err.Error())
			return
		}

	} else {
		err := redis.LPush(CRAWLER_TASK_QUEUE, link)

		if err != nil {
			log.Error(err.Error())
			return
		}

		if err := redis.HIncryBy(CRAWLER_HOST_COUNTER, linkHostname, 1); err != nil {
			log.Error(err.Error())
		}
	}
}

func reserveBloomKey() {
	if ok := redis.Exists(CRAWLER_BLOOM_KEY); ok {
		return
	}

	if err := redis.BloomReserve(CRAWLER_BLOOM_KEY, 0.0000001, 1000000000); err != nil {
		log.Error("failed to reserve redis bloom key %s, err:%s", CRAWLER_BLOOM_KEY, err.Error())
		os.Exit(1)
	}

	log.Info("redis bloom key %s is reserved successfully!", CRAWLER_BLOOM_KEY)
}

func parseCollyResponseToModel(response *colly.Response) *model.Content {
	content := &model.Content{}

	doc, _ := goquery.NewDocumentFromReader(bytes.NewReader(response.Body))
	title := doc.Find("h1").Text()

	if len(title) > 0 {
		content.Title = title
	}

	var desc string
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		if val, _ := s.Attr("name"); strings.Contains(val, "description") {
			desc, _ = s.Attr("content")
		}
		if len(desc) == 0 {
			if val, _ := s.Attr("property"); strings.Contains(val, "description") {
				desc, _ = s.Attr("content")
			}
		}
	})

	if len(desc) > 0 {
		content.Desc = desc
	}

	// loop through the http request header tags and return all in one string
	var headers []model.Header

	for k, v := range *response.Headers {
		headers = append(headers, model.Header{
			Key:   k,
			Value: v[0],
		})
	}

	content.HttpHeaders = headers

	return content
}

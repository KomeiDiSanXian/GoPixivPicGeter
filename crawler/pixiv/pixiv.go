package pixiv

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

	"GoPixivPicGeter/crawler"
	"GoPixivPicGeter/global"
	"GoPixivPicGeter/model"

	"github.com/tidwall/gjson"
)

var ipTables = map[string]string{
	"pixiv.net":   "210.140.92.183:443",
	"i.pximg.net": "210.140.92.142:443",
}

var (
	DailyIllust = "https://www.pixiv.net/ranking.php?mode=daily&content=illust&format=json&p=%d" //max 10 pages
	// bug:
	// 需要登陆以后才能访问
	DailyR18Illust = "https://www.pixiv.net/ranking.php?mode=daily_r18&content=illust&format=json&p=%d" //max 2	pages
)

const (
	R18PageSize    = 2
	NormalPageSize = 10
	referer        = "https://www.pixiv.net/"
	filePath       = "./data/PixivPic/normal/" //下载路径
	PixivImg       = "https://i.pximg.net/"
	Original       = "img-original/img/"            // 原图
	Regular        = "img-master/img/"              // 1200
	Small          = "c/540x540_70/img-master/img/" // 540x540
	Thumb          = "c/250x250_80_a2/img-master/img/"
	Mini           = "c/48x48/img-master/img/"
)

func NewPixivClient(dialAdress string) *http.Client {
	// 设置req client
	return &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
			// 隐藏 sni 标志
			TLSClientConfig: &tls.Config{
				ServerName:         "-",
				InsecureSkipVerify: true,
			},
			// 更改 dns
			Dial: func(network, addr string) (net.Conn, error) {
				return net.Dial("tcp", dialAdress)
			},
		},
	}
}

func pixivBodyWriteInMySQL(illusts []model.Illust) {
	for _, illust := range illusts {
		illust.Create(global.DBEngine)
	}
}

func readPixivBody(resBody io.Reader) (illusts []model.Illust) {
	body, _ := io.ReadAll(resBody)
	var mu sync.RWMutex
	var wg sync.WaitGroup
	var r18 = strings.Contains(gjson.GetBytes(body, "mode").Str, "r18")
	gjson.GetBytes(body, "contents").ForEach(func(_, value gjson.Result) bool {
		wg.Add(1)
		go func() {
			illust := jsonToIllust(value)
			illust.R18 = r18
			mu.Lock()
			illusts = append(illusts, illust)
			mu.Unlock()
			wg.Done()
		}()
		return true
	})
	wg.Wait()
	return illusts
}

func jsonToIllust(json gjson.Result) model.Illust {
	var tags []model.Tag
	json.Get("tags").ForEach(func(_, v gjson.Result) bool {
		tags = append(tags, model.Tag{TagName: v.Str})
		return true
	})
	illust := model.Illust{
		Title:           json.Get("title").Str,
		Author:          json.Get("user_name").Str,
		IllustID:        uint(json.Get("illust_id").Uint()),
		AuthorID:        uint(json.Get("user_id").Uint()),
		UploadTimestamp: json.Get("illust_upload_timestamp").Int(),
		Tags:            tags,
		PageCount:       int(json.Get("illust_page_count").Int()),
	}
	return illust
}

func WritePixivPicInfoInMySQL(pageSize int, url string) {
	var wg sync.WaitGroup
	var mu sync.RWMutex
	for i := 0; i < pageSize; i++ {
		wg.Add(1)
		go func(i int) {
			res, err := requestPage(url, i+1)
			if err != nil {
				return
			}
			defer res.Body.Close()
			mu.Lock()
			pixivBodyWriteInMySQL(readPixivBody(res.Body))
			mu.Unlock()
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func requestPage(url string, page int) (*http.Response, error) {
	c := NewPixivClient(ipTables["pixiv.net"])
	link := fmt.Sprintf(url, page)
	req, _ := http.NewRequest("GET", link, nil)
	req.Header.Set("User-Agent", crawler.GenerateRandomUA())
	res, err := c.Do(req)
	if err != nil {
		log.Println("请求", page, "页时发生err: ", err)
		return nil, err
	}
	return res, nil
}

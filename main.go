package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// P站 无污染 IP 地址
var IPTables = map[string]string{
	"pixiv.net":   "210.140.92.183:443",
	"i.pximg.net": "210.140.92.142:443",
}

var FilePath string = "./data/normal/" //下载路径

//日榜json
type Daily struct {
	Contents []struct {
		Title             string   `json:"title"`
		Date              string   `json:"date"`
		Tags              []string `json:"tags"`
		URL               string   `json:"url"`
		IllustType        string   `json:"illust_type"`
		IllustBookStyle   string   `json:"illust_book_style"`
		IllustPageCount   string   `json:"illust_page_count"`
		UserName          string   `json:"user_name"`
		ProfileImg        string   `json:"profile_img"`
		IllustContentType struct {
			Sexual     int  `json:"sexual"`
			Lo         bool `json:"lo"`
			Grotesque  bool `json:"grotesque"`
			Violent    bool `json:"violent"`
			Homosexual bool `json:"homosexual"`
			Drug       bool `json:"drug"`
			Thoughts   bool `json:"thoughts"`
			Antisocial bool `json:"antisocial"`
			Religion   bool `json:"religion"`
			Original   bool `json:"original"`
			Furry      bool `json:"furry"`
			Bl         bool `json:"bl"`
			Yuri       bool `json:"yuri"`
		} `json:"illust_content_type"`
		IllustSeries          interface{} `json:"illust_series"`
		IllustID              int         `json:"illust_id"`
		Width                 int         `json:"width"`
		Height                int         `json:"height"`
		UserID                int         `json:"user_id"`
		Rank                  int         `json:"rank"`
		YesRank               int         `json:"yes_rank"`
		RatingCount           int         `json:"rating_count"`
		ViewCount             int         `json:"view_count"`
		IllustUploadTimestamp int         `json:"illust_upload_timestamp"`
		Attr                  string      `json:"attr"`
	} `json:"contents"`
	Mode      string      `json:"mode"`
	Content   string      `json:"content"`
	Page      int         `json:"page"`
	Prev      interface{} `json:"prev"`
	Next      interface{} `json:"next"`
	Date      string      `json:"date"`
	PrevDate  interface{} `json:"prev_date"`
	NextDate  interface{} `json:"next_date"`
	RankTotal int         `json:"rank_total"`
}

//数据库结构体
type PictureDB struct {
	gorm.Model
	Title    string
	Date     string
	URL      string
	Tags     string
	Pid      int
	UserName string
	UserID   int
}

//标签
type Tag struct {
	Tags []string
}

func main() {
	cron2 := cron.New()
	cron1 := cron.New()
	fmt.Print("定时器启动中...\n")
	err := cron2.AddFunc("0 0 6 * * ?", DailyDownload)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Print("日榜计时器启动成功！\n将在每天6点时自动获取图片，并将信息写入数据库\n")
	}
	err = cron1.AddFunc("0 0 6 * * ?", MonthlyDownload)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Print("月榜计时器启动成功！\n将在每天6点时自动获取图片，并将信息写入数据库\n")
	}
	cron2.Start()
	cron1.Start()
	//启动时自动运行一次
	fmt.Print("10秒后爬取一次图片...")
	time.Sleep(10 * time.Second)
	go DailyDownload()
	go MonthlyDownload()
	select {}
}

//月榜下载
func MonthlyDownload() {
	var link, name [50]string
	var w sync.WaitGroup
	var n int = 0
	//创建数据库
	db, err := gorm.Open(sqlite.Open("Pixiv.db"), &gorm.Config{})
	if err != nil {
		panic("链接数据库失败")
	}
	db.AutoMigrate(&PictureDB{})
	//获取链接
	log.Printf("获取链接中...\n")
	for i := 0; i < 5; i++ {
		json, err := IdGet("https://www.pixiv.net/ranking.php?mode=monthly&content=illust&format=json&p=" + strconv.Itoa(i+1))
		if err != nil {
			log.Print(err)
			continue
		}
		log.Printf("获取到月排行榜第%d页...\n", i+1)
		for i := 0; i < len(json.Contents); i++ {
			link[i] = Replace(json.Contents[i].URL)
			name[i] = strconv.Itoa(json.Contents[i].IllustID) + ".jpg"
			n++
			log.Printf("获取了%d条图片信息\n", n)
		}
		//开始下载
		log.Printf("正在下载图片到 %s 目录...\n", FilePath)
		for i := 0; i < len(link); i++ {
			//检查文件是否已经下载
			a, _ := IsCreated(name[i])
			if a {
				log.Printf("%s已经下载，跳过...\n", name[i])
			} else {
				//写入数据库
				url := "https://www.pixiv.net/artworks/" + strconv.Itoa(json.Contents[i].IllustID)
				title := json.Contents[i].Title
				date := json.Contents[i].Date
				username := json.Contents[i].UserName
				pid := json.Contents[i].IllustID
				uid := json.Contents[i].UserID
				//无论是否存在条目都写入...待修复
				err := CreatePicData(db, title, date, url, username, pid, uid, &Tag{Tags: json.Contents[i].Tags})
				if err != nil {
					log.Println("创建数据库条目", pid, "失败：", err)
				}
				w.Add(1)
				go GetPicture(link[i], name[i], &w)
				log.Printf("下载%s完成\n", name[i])
			}
		}
		w.Wait()
	}
	log.Printf("月榜图片全部下载完成，下载了%d张图片\n", n)
}

//日榜下载
func DailyDownload() {
	var link, name [50]string
	var w sync.WaitGroup
	var n int = 0
	//创建数据库
	db, err := gorm.Open(sqlite.Open("Pixiv.db"), &gorm.Config{})
	if err != nil {
		panic("链接数据库失败")
	}
	db.AutoMigrate(&PictureDB{})
	//获取链接
	log.Printf("获取链接中...\n")
	for i := 0; i < 10; i++ {
		json, err := IdGet("https://www.pixiv.net/ranking.php?p=" + strconv.Itoa(i+1) + "&format=json")
		if err != nil {
			log.Print(err)
			continue
		}
		log.Printf("获取到日排行榜第%d页...\n", i+1)
		for i := 0; i < len(json.Contents); i++ {
			link[i] = Replace(json.Contents[i].URL)
			name[i] = strconv.Itoa(json.Contents[i].IllustID) + ".jpg"
			n++
			log.Printf("获取了%d条图片信息\n", n)
		}
		//开始下载
		log.Printf("正在下载图片到 %s 目录...\n", FilePath)
		for i := 0; i < len(link); i++ {
			//检查文件是否已经下载
			a, _ := IsCreated(name[i])
			if a {
				log.Printf("%s已经下载，跳过...\n", name[i])
			} else {
				//写入数据库
				url := "https://www.pixiv.net/artworks/" + strconv.Itoa(json.Contents[i].IllustID)
				title := json.Contents[i].Title
				date := json.Contents[i].Date
				username := json.Contents[i].UserName
				pid := json.Contents[i].IllustID
				uid := json.Contents[i].UserID
				//无论是否存在条目都写入...待修复
				err := CreatePicData(db, title, date, url, username, pid, uid, &Tag{Tags: json.Contents[i].Tags})
				if err != nil {
					log.Println("创建数据库条目", pid, "失败：", err)
				}
				w.Add(1)
				go GetPicture(link[i], name[i], &w)
				log.Printf("下载%s完成\n", name[i])
			}
		}
		w.Wait()
	}
	log.Printf("日榜图片全部下载完成，下载了%d张图片\n", n)
}

func IdGet(link string) (*Daily, error) {
	// 获取IP地址
	domain, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	// P站特殊客户端
	client := &http.Client{
		// 解决中国大陆无法访问的问题
		Transport: &http.Transport{
			DisableKeepAlives: true,
			// 隐藏 sni 标志
			TLSClientConfig: &tls.Config{
				ServerName:         "-",
				InsecureSkipVerify: true,
			},
			// 更改 dns
			Dial: func(network, addr string) (net.Conn, error) {
				return net.Dial("tcp", IPTables["pixiv.net"])
			},
		},
	}
	// 网络请求
	request, _ := http.NewRequest("GET", link, nil)
	request.Header.Set("Host", domain.Host)
	request.Header.Set("Accept", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:6.0) Gecko/20100101 Firefox/6.0")
	res, err := client.Do(request)
	if err != nil {
		return nil, errors.New("请求错误")
	}
	defer res.Body.Close()
	result := &Daily{}
	if err := json.NewDecoder(res.Body).Decode(result); err != nil {
		panic(err)
	}
	return result, nil
}

//检查图片是否下载
func IsCreated(name string) (bool, error) {
	files, err := ioutil.ReadDir(FilePath)
	if err != nil {
		return false, err
	}
	for _, curfile := range files {
		if curfile.IsDir() {
			IsCreated(name + "/" + curfile.Name())
		} else {
			if curfile.Name() == name {
				return true, nil
			}
		}
	}
	return false, err
}

//使用代理
func Replace(url string) string {
	url = strings.ReplaceAll(url, "/c/240x480", "") //使用1200的图片，原图极易被tx风控
	return url
}

//下载图片（并发）
func GetPicture(link, name string, w *sync.WaitGroup) {
	// 获取IP地址
	domain, err := url.Parse(link)
	if err != nil {
		log.Printf("url.Parse -> %v", err)
	}
	// P站特殊客户端
	client := &http.Client{
		// 解决中国大陆无法访问的问题
		Transport: &http.Transport{
			DisableKeepAlives: true,
			// 隐藏 sni 标志
			TLSClientConfig: &tls.Config{
				ServerName:         "-",
				InsecureSkipVerify: true,
			},
			// 更改 dns
			Dial: func(network, addr string) (net.Conn, error) {
				return net.Dial("tcp", IPTables["i.pximg.net"])
			},
		},
	}
	// 网络请求
	request, _ := http.NewRequest("GET", link, nil)
	request.Header.Set("Host", domain.Host)
	//UA伪装
	request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:98.0) Gecko/20100101 Firefox/98.0")
	request.Header.Set("Referer", "https://www.pixiv.net/")
	res, err := client.Do(request)

	if err != nil {
		log.Printf("http.Get -> %v", err)
		defer res.Body.Close()
		log.Printf("文件传输时发生错误，重新下载%s中...", name)
		w.Add(1)
		GetPicture(link, name, w)
	}
	defer res.Body.Close()

	//放入文件夹
	_ = os.MkdirAll(FilePath, 0777)
	out, err := os.Create(FilePath + name)
	if err != nil {
		log.Printf("os.Create -> %v", err)
	}
	wt := bufio.NewWriter(out)
	defer out.Close()
	n, err := io.Copy(wt, res.Body)
	log.Println(name, "-> write", n)
	if err != nil {
		log.Printf("io.Copy -> %v", err)
		os.Remove(FilePath + name)
		log.Printf("文件传输时发生错误，重新下载%s中...", name)
		w.Add(1)
		GetPicture(link, name, w)
	}
	wt.Flush()
	w.Done()
}

//新数据写入数据库
func CreatePicData(db *gorm.DB, title, date, url, username string, pid, userid int, mutitags *Tag) error {
	tags, err := mutitags.Json2String()
	if err != nil {
		return err
	}
	pic := PictureDB{
		Title:    title,
		Date:     date,
		URL:      url,
		Tags:     tags,
		Pid:      pid,
		UserName: username,
		UserID:   userid,
	}
	err = db.Debug().Create(&pic).Error
	if err != nil {
		return err
	}
	log.Println("数据条目创建成功")
	return nil
}

func (p *Tag) Json2String() (string, error) {
	bs, err := json.Marshal(p)
	if err != nil {
		return "{}", err
	}
	return string(bs), nil
}

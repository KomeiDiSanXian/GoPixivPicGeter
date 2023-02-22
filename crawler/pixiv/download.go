package pixiv

import (
	"GoPixivPicGeter/crawler"
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// 由pid和上传时间戳组成下载链接
func NewPixivDownloadURL(pid uint, timestamp int64, sacle string) string {
	form := "2006/01/02/15/04/05/"
	// 改为日本时区
	timestamp += 3600
	t := time.Unix(timestamp, 0).Format(form)
	switch sacle {
	case Original:
		return fmt.Sprintf("%s%s%s%d_p0.jpg", PixivImg, Original, t, pid)
	case Regular:
		return fmt.Sprintf("%s%s%s%d_p0_master1200.jpg", PixivImg, Regular, t, pid)
	case Small:
		return fmt.Sprintf("%s%s%s%d_p0_master1200.jpg", PixivImg, Small, t, pid)
	case Thumb:
		return fmt.Sprintf("%s%s%s%d_p0_square1200.jpg", PixivImg, Thumb, t, pid)
	case Mini:
		return fmt.Sprintf("%s%s%s%d_p0_square1200.jpg", PixivImg, Mini, t, pid)
	}
	return ""
}

// 下载同一pid的多张图片（图集）
func DownloadPics(link string, pages int) {
	var wg sync.WaitGroup
	for i := 0; i < pages; i++ {
		wg.Add(1)
		go func(i int) {
			url := urlToDownload(link, i)
			DownloadPic(url)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

// 同一pid的第n张图片下载链接
func urlToDownload(link string, n int) string {
	return strings.ReplaceAll(link, "_p0", "_p"+strconv.Itoa(n))
}

func DownloadPic(link string) {
	cli := NewPixivClient(ipTables["i.pximg.net"])
	req, _ := http.NewRequest("GET", link, nil)
	req.Header.Set("User-Agent", crawler.GenerateRandomUA())
	req.Header.Set("Referer", referer)
	resp, err := requestPic(cli, req)
	if resp == nil {
		log.Printf("Resp body is nil!\nRequest %s failed: %v", link, err)
		return
	}
	defer resp.Body.Close()

	// 三次保存机会
	for i := 0; i < 3; i++ {
		err := savePicture(link, resp)
		if err == nil {
			return
		}
		// 保存失败，刷新 resp
		resp, err = requestPic(cli, req)
		if resp == nil {
			log.Printf("Resp body is nil!\nRequest %s failed: %v", link, err)
			return
		}
	}
}

// 创建文件夹
func createDir(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	os.Chmod(path, os.ModePerm)
	return nil
}

func requestPic(cli *http.Client, req *http.Request) (*http.Response, error) {
	// 三次重试机会
	var resp *http.Response
	var err error
	for i := 0; i < 3; i++ {
		resp, err = cli.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			break
		}
	}
	return resp, err
}

func savePicture(link string, resp *http.Response) error {
	createDir(filePath)
	s := strings.Split(link, "/")
	name := s[len(s)-1]
	fileName := filePath + name
	out, err := os.Create(fileName)
	if err != nil {
		log.Printf("%s os.Create err: %v", name, err)
		return err
	}
	wt := bufio.NewWriter(out)
	defer out.Close()

	n, err := io.Copy(wt, resp.Body)
	log.Println(name, "-> write", n)
	if n != resp.ContentLength {
		err = errors.New("response content length mismatch")
		log.Printf("%s io.Copy err: %v", name, err)
		os.Remove(fileName)
		return err
	}
	if err != nil {
		log.Printf("%s io.Copy err: %v", name, err)
		os.Remove(fileName)
		return err
	}
	return nil
}

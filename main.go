package main

// waiting for reconstruction
/*
import (
	"fmt"
	"log"
	"time"

	"GoPixivPicGeter/global"
	"GoPixivPicGeter/model"
	"github.com/robfig/cron"
)

func init() {
	db, err := model.NewDBEngine()
	if err != nil {
		log.Panicf("model.NewDBEngine failed: %v", err)
	}
	global.DBEngine = db
}

func main() {
	cron2 := cron.New()
	cron1 := cron.New()
	fmt.Print("定时器启动中...\n")
	err := cron2.AddFunc("0 0 6 * * ?", func() { Download(daily) })
	if err != nil {
		WriteLog(err.Error())
		fmt.Println(err)
	} else {
		fmt.Print("日榜计时器启动成功！\n将在每天6点时自动获取图片，并将信息写入数据库\n")
	}
	err = cron1.AddFunc("0 0 6 * * ?", func() { Download(monthly) })
	if err != nil {
		WriteLog(err.Error())
		fmt.Println(err)
	} else {
		fmt.Print("月榜计时器启动成功！\n将在每天6点时自动获取图片，并将信息写入数据库\n")
	}
	cron2.Start()
	cron1.Start()
	//启动时自动运行一次
	fmt.Print("10秒后爬取一次图片...")
	time.Sleep(10 * time.Second)
	go Download(daily)
	go Download(monthly)
	select {}
}
*/

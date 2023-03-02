package pixiv

import (
	"sync"

	"GoPixivPicGeter/global"
	"GoPixivPicGeter/model"
)

// GetR18PicInfo returns *model.Illust array.
// n_offset is the number of the offset.
func GetR18PicInfo(n_offset int) (illusts []*model.Illust) {
	return GetPicInfo(dayR18, n_offset, true)
}

// GetDailyPicInfo returns *model.Illust array.
// n_offset is the number of the offset.
func GetDailyPicInfo(n_offset int) (illusts []*model.Illust) {
	return GetPicInfo(day, n_offset, false)
}

// WritePixivPicInfoInMySQL will write picture info in MySQL.
// The number of writes depends on n_offset (exact num: n_offset*Offset).
// getpic is a function that returns *model.Illust array, which needs n_offset.
// For example, GetR18PicInfo is a function body that satisfies getpic
func WritePixivPicInfoInMySQL(n_offset int, getpic func(n_offset int) []*model.Illust) {
	var mu sync.RWMutex
	pic := getpic(n_offset)
	for _, img := range pic {
		mu.Lock()
		img.Create(global.DBEngine)
		mu.Unlock()
	}
}

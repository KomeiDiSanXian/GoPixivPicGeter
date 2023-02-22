package model

import "gorm.io/gorm"

type Illust struct {
	Model
	Title           string
	Author          string
	IllustID        uint `gorm:"primaryKey;autoIncrement:false"`
	AuthorID        uint
	UploadTimestamp int64
	Tags            []Tag `gorm:"foreignKey:IllustID"`
	PageCount       int
	R18             bool
	IsSaved         bool
	URLPath         string
}

type Tag struct {
	Model
	IllustID uint
	TagName  string `gorm:"primaryKey"`
}

func (i Illust) TableName() string {
	return "pixiv_illusts"
}

func (t Tag) TableName() string {
	return "pixiv_tags"
}

func (i Illust) Create(db *gorm.DB) error {
	return db.Create(&i).Error
}

func (i Illust) GetIllust(db *gorm.DB) (Illust, error) {
	var ii Illust
	if err := db.Preload("Tags").Find(&ii, i.IllustID).Error; err != nil {
		return ii, err
	}
	return ii, nil
}

func (i Illust) Update(db *gorm.DB) error {
	return db.Updates(&i).Error
}

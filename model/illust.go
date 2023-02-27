package model

import (
	"errors"
	"math/rand"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

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
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&i).Error
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

func (i Illust) GetRandom(db *gorm.DB) (Illust, error) {
	var c int64
	var ii Illust
	err := db.Table(i.TableName()).Count(&c).Error
	if c == 0 {
		return ii, errors.New("no records")
	}
	if err := db.Preload("Tags").Offset(rand.Intn(int(c))).First(&ii).Error; err != nil {
		return ii, err
	}
	return ii, err
}

func (i Illust) UpdatePicIsSaved(db *gorm.DB) error {
	i.IsSaved = true
	return i.Update(db)
}

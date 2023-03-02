package model

import (
	"errors"
	"math/rand"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Illust struct {
	Model
	Title      string
	Type       string
	Author     User `gorm:"foreignKey:AuthorID"`
	AuthorID   uint
	IllustID   uint `gorm:"primaryKey;autoIncrement:false"`
	UploadTime string
	Tags       []Tag `gorm:"foreignKey:IllustID"`
	PageCount  int
	R18        bool
	IsSaved    bool
}

type User struct {
	AuthorID uint `gorm:"primaryKey;autoIncrement:false"`
	Name     string
	Account  string
}
type Tag struct {
	Model
	IllustID  uint
	Name      string `gorm:"primaryKey"`
	TransName string
}

func (i Illust) TableName() string {
	return "pixiv_illusts"
}

func (t Tag) TableName() string {
	return "pixiv_tags"
}

func (u User) TableName() string {
	return "pixiv_authors"
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

package main

import (
	"log"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

func NewGorm(dbPath string) (db *gorm.DB) {
	db, err := gorm.Open("sqlite3", dbPath)
	if err != nil {
		panic(err)
	}
	db.SingularTable(true)
	db.BlockGlobalUpdate(true)
	db.DB().SetConnMaxLifetime(time.Second * 300)
	db.DB().SetMaxIdleConns(100)
	db.DB().SetMaxOpenConns(5000)
	db.LogMode(true)

	if !db.HasTable(&Doc{}) {
		if err := db.CreateTable(&Doc{}).Error; err != nil {
			log.Fatalf("failed to create table:%v", err)
		}
	}

	return
}

const tableDoc = "doc"

type Doc struct {
	ID          int `gorm:"primaryKey;autoincrement"`
	Repo        string
	Branch      string
	Description string
	UploadedAt  int64
	Content     string
}

func CreateDoc(db *gorm.DB, doc *Doc) error {
	doc.UploadedAt = time.Now().Unix()
	return db.Save(doc).Error
}

func GetByRepoBranch(db *gorm.DB, repo, branch string) (*Doc, error) {
	v := new(Doc)
	err := db.Where("repo=? and branch=?", repo, branch).Find(v).Error
	return v, err
}

func List(db *gorm.DB) ([]*Doc, error) {
	list := make([]*Doc, 0)
	err := db.Table(tableDoc).Select("id, repo, branch, description, uploaded_at").Order("uploaded_at DESC").Find(&list).Error
	if gorm.IsRecordNotFoundError(err) {
		return list, nil
	}
	return list, err
}

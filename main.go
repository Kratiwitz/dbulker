package main

import (
	"log"
)

type Manga struct {
	// Id            string     `bson:"_id" bulker:"_id" bulker_rdb:"id" primary:"true"`
	Name          string     `bulker:"name"`
	SecondaryName string     `bulker:"secondaryName" bulker_rdb:"secondary_name"`
	Author        string     `bulker:"author"`
	Description   string     `bulker:"description"`
	Categories    []Category `bulker:"categories" relation_name:"manga_category" table:"category"`
	ChapterList   []Chapter  `bulker:"chapterList" relation_name:"manga_chapter" table:"chapter"`
}

type Category struct {
	Id   string `bulker:"id" bulker_type:"primary" bulker_column:"INT NOT NULL AUTO_INCREMENT"`
	Name string `bulker:"name" bulker_type:"main" bulker_column:"TEXT NULL"`
}

type Chapter struct {
	Id  string `bulker:"id" bulker_type:"primary" bulker_column:"INT NOT NULL AUTO_INCREMENT"`
	Url string `bulker:"url" bulker_type:"main" bulker_column:"TEXT NULL"`
}

func main() {
	mongodb, err := NewMongoDBClient("mongodb://localhost:27017", "manga_tx")
	CheckError(err)

	mysqldb, err := NewMysqlClient("root:root@tcp(127.0.0.1:3306)", "manga_luck")
	CheckError(err)

	data, err := mongodb.GetAll("mangas", Manga{})
	CheckError(err)

	err = mysqldb.FillAutoNestedSingle("manga", data[0])
	CheckError(err)
}

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

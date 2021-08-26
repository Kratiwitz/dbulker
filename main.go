package main

import (
	"log"
)

type Manga struct {
	Id            string     `bson:"_id" bulker:"id" primary:""`
	Name          string     `bulker:"name"`
	SecondaryName string     `bulker:"secondary_name"`
	Author        string     `bulker:"author"`
	Description   string     `bulker:"description"`
	Categories    []Category `bulker:"category" relation_name:"manga_category"`
	ChapterList   []Chapter  `bulker:"chapter" relation_name:"manga_chapter"`
}

type Category string
type Chapter string

func main() {
	mongodb, err := NewMongoDBClient("mongodb://localhost:27017", "manga_tx")
	CheckError(err)

	mysqldb, err := NewMysqlClient("root:root@tcp(127.0.0.1:3306)", "manga_luck")
	CheckError(err)

	data, err := mongodb.GetAll("mangas", Manga{})
	CheckError(err)

	_, err = mysqldb.FillAutoNestedSingle("manga", data[0])

	CheckError(err)
}

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

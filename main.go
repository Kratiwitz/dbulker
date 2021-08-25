package main

import (
	"dbbulker/database"
	"log"
)

type Manga struct {
	Id            string     `bson:"_id" bulker:"id"`
	Name          string     `bulker:"name"`
	SecondaryName string     `bulker:"secondary_name"`
	Author        string     `bulker:"author"`
	Description   string     `bulker:"description"`
	Categories    []Category `bulker:"category"`
	ChapterList   []Chapter  `bulker:"chapter"`
}

type Category string
type Chapter string

func main() {
	mongodb, err := database.NewMongoDBClient("mongodb://localhost:27017", "manga_tx")
	CheckError(err)

	mysqldb, err := database.NewMysqlClient("root:root@tcp(127.0.0.1:3306)", "manga_luck")
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

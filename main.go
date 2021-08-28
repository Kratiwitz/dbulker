package main

import (
	"fmt"
	"log"
)

type Manga struct {
	Id            string     `bson:"_id" bulker:"_id" primary:"true"`
	Name          string     `bulker:"name"`
	SecondaryName string     `bulker:"secondary_name"`
	Author        string     `bulker:"author"`
	Description   string     `bulker:"description"`
	Categories    []Category `bulker:"categories" relation_name:"manga_category"`
	ChapterList   []Chapter  `bulker:"chapterList" relation_name:"manga_chapter"`
}

type Category struct {
	Id   string `bulker:"id" primary:"true"`
	Name string `bulker:"name" bulker_type:"main"`
}

type Chapter struct {
	Id  string `bulker:"id" primary:"true"`
	Url string `bulker:"url" bulker_type:"main"`
}

func main() {
	mongodb, err := NewMongoDBClient("mongodb://localhost:27017", "manga_tx")
	CheckError(err)

	// mysqldb, err := NewMysqlClient("root:root@tcp(127.0.0.1:3306)", "manga_luck")
	CheckError(err)

	data, err := mongodb.GetAll("mangas", Manga{})
	CheckError(err)

	fmt.Println(data[0].(Manga).Categories[0].Name)
	// _, err = mysqldb.FillAutoNestedSingle("manga", data[0])
	// CheckError(err)
}

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

# DBulker [![GoDoc](https://godoc.org/github.com/kratiwitz/dbulker?status.svg)](http://godoc.org/github.com/kratiwitz/dbulker) [![Go Report Card](https://goreportcard.com/badge/github.com/kratiwitz/dbulker)](https://goreportcard.com/report/github.com/kratiwitz/dbulker)

Dbulker is a data transporter between MongoDB to Mysql (for now)

## Getting Package

```sh
go get github.com/kratiwitz/dbulker
```

## Prerequisites
- Go
- Mysql
- MongoDB

## Simple Usage
Create struct for your MongoDB object. `Id` field has a lot of tags. `bson` tag for MongoDB, `bulker` for matches field in `dbulker`, `bulker_rdb` for MySQL column.
```go
type Movie struct { 
	Id            string     `bson:"_id" bulker:"_id" bulker_rdb:"id"`
	Name          string     `bulker:"name"`
	Author        string     `bulker:"author"`
	Description   string     `bulker:"description"`
}
```

Create MongoDB client
```go
mongodb, err := dbulker.NewMongoDBClient("mongodb://localhost:27017", "movies")
```

Create Mysql client
```go
mysqldb, err := dbulker.NewMysqlClient("root:root@tcp(127.0.0.1:3306)", "movie")
```

Get Movies from Mongo
```go
data, err := mongodb.GetAll("movies", Movie{})
```

And write first data to Mysql
```go
mysqldb.FillAutoPlainSingle("movie", data[0])
```

Or write all data
```go
mysqldb.FillAutoPlainMultiple("movie", data)
```

## Relational
Let's assume that your data in MongoDB is as follows and you want to pass this data to Mysql in a related way.

```json
{
    "name": "test",
    "author": "test test",
    "categories": ["test", "test"],
    "sectionList": [
        "https://test.com/test/section-1",
        "https://test.com/test/section-2"
    ],
    "image": "test/test.png",
    "description": "test",
}
```

You must separate `sections List` and `categories` into different tables. You can use that.

```go
type Movie struct {
	Name          string     `bulker:"name"`
	Author        string     `bulker:"author"`
	Description   string     `bulker:"description"`
	Categories    []Category `bulker:"categories" relation_name:"movie_category" table:"category"`
	SectionList   []Section  `bulker:"sectionList" relation_name:"movie_section" table:"section"`
}

type Category struct {
	Id   string `bulker:"id" bulker_type:"primary" bulker_column:"INT NOT NULL AUTO_INCREMENT"`
	Name string `bulker:"name" bulker_type:"main" bulker_column:"TEXT NULL"`
}

type Section struct {
	Id  string `bulker:"id" bulker_type:"primary" bulker_column:"INT NOT NULL AUTO_INCREMENT"`
	Url string `bulker:"url" bulker_type:"main" bulker_column:"TEXT NULL"`
}

func main() {
	mongodb, err := NewMongoDBClient("mongodb://localhost:27017", "movie_db")
	CheckError(err)

	mysqldb, err := NewMysqlClient("root:root@tcp(127.0.0.1:3306)", "movie_db")
	CheckError(err)

	data, err := mongodb.GetAll("movies", Movie{})
	CheckError(err)

	err = mysqldb.FillAutoNestedSingle("movie", data[0])
	CheckError(err)
}

func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
```

Your database result will look like this;

Movie Table

| id | name | author | description |
| -- | ---- | ------ | ----------- |
| 0  | test | test   | test        |

Category Table

| id | name |
| -- | ---- |
| 0  | test |
| 1  | test |

Section Table

| id | url |
| -- | ---- |
| 0  | test |
| 1  | test |

Junction Table `movie_category`

| movie_id | category_id |
| -------- | ----------- |
| 1        | 0           |
| 1        | 1           |

Junction Table `movie_section`

| movie_id | section_id |
| -------- | ---------- |
| 1        | 0          |
| 1        | 1          |


## Tags
### `bulker`
`bulker` tag is a must for dbulker understand what is name of your field on MongoDB

Usage

```go
Name string `bulker:"name"`
```

### `bulker_rbd`
`bulker_rbd` represents your field in your Table of Mysql

Usage 
```go
SecondaryName string `bulker_rdb:"secondary_name"`
```

### `relation_name`
`relation_name` declare a junction table name for your relation of data

Usage 
```go
Categories []Category `relation_name:"movie_category"`
```

### `table`
`table` declare a different table name for your nested data

Usage 
```go
Categories []Category `table:"category"`
```

### `bulker_type`
`bulker_type` declare selected column is `PRIMARY` or `main`. If select `PRIMARY` the column will `id` in Mysql, if select `main` column is main field of data.

Usage 
```go
Name string `bulker:"name" bulker_type:"main"`
```

### `bulker_column`
`bulker_column` directly set of your Mysql field.

Usage 
```go
Id  string `bulker:"id" bulker_column:"INT NOT NULL AUTO_INCREMENT"`
```

### `bulker_unique`
`bulker_unique` reduce for repeated values.

Usage 
```go
Name string `bulker:"name" bulker_type:"main" bulker_unique:"true" bulker_column:"TEXT NULL"`
```

## License

The MIT License (MIT). See [License File](LICENSE) for more information.
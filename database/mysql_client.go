package database

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type MysqlDB struct {
	ConnectionString string
	Database         string
	client           *sql.DB
}

// NewMysqlClient create MysqlDB client, connect to db and return MysqlDB
func NewMysqlClient(connectionString, database string) (*MysqlDB, error) {
	mysqlDb := &MysqlDB{
		ConnectionString: connectionString,
		Database:         database,
	}

	err := mysqlDb.Connect()

	return mysqlDb, err
}

// Connect open connection to mysql
func (mysqldb *MysqlDB) Connect() error {
	connectionString := fmt.Sprintf("%s/%s", mysqldb.ConnectionString, mysqldb.Database)
	db, err := sql.Open("mysql", connectionString)

	if err != nil {
		return err
	}

	mysqldb.client = db

	return nil
}

// FillAutoPlainMultiple writes multiple given data to
// given table and returns an error if any
// Ignore any nested child.
func (mysqldb *MysqlDB) FillAutoPlainMultiple(targetTable string, data []interface{}) error {
	sql := "INSERT INTO " + targetTable + " ("

	dataType := reflect.TypeOf(data[0])
	numField := dataType.NumField()
	dataIndexes := []int{}

	for i := 0; i < numField; i++ {
		kind := dataType.Field(i).Type.Kind()

		if kind == reflect.Array || kind == reflect.Slice {
			if sql[len(sql)-1:] == "," {
				sql = sql[:len(sql)-1]
			}
			continue
		}

		tag := dataType.Field(i).Tag.Get("bulker")

		if numField-1 != i {
			tag += ","
		}

		dataIndexes = append(dataIndexes, i)

		sql += tag
	}

	sql += ") VALUES "

	dataLength := len(data)

	for index, v := range data {
		sql += "("
		t := reflect.ValueOf(v)

		for i, valueIndex := range dataIndexes {
			value := strings.ReplaceAll(t.Field(valueIndex).String(), "\"", "\\\"")
			value = "\"" + value + "\""

			if len(dataIndexes)-1 != i {
				value += ","
			}

			sql += value
		}

		if dataLength-1 == index {
			sql += ")"
		} else {
			sql += "),"
		}
	}

	_, err := mysqldb.client.Exec(sql)

	if err != nil {
		return err
	}

	return nil
}

// FillAutoPlainSingle, writes given data to the given table
// and returns the last inserted id and error if any.
// Ignore any nested child.
func (mysqldb *MysqlDB) FillAutoPlainSingle(targetTable string, data interface{}) (int64, error) {
	sql := "INSERT INTO " + targetTable + " ("

	t := reflect.TypeOf(data)
	numField := t.NumField()
	dataIndexes := []int{}

	for i := 0; i < numField; i++ {
		kind := t.Field(i).Type.Kind()

		if kind == reflect.Array || kind == reflect.Slice {
			if sql[len(sql)-1:] == "," {
				sql = sql[:len(sql)-1]
			}
			continue
		}

		tag := t.Field(i).Tag.Get("bulker")

		if numField-1 != i {
			tag += ","
		}

		dataIndexes = append(dataIndexes, i)

		sql += tag
	}

	sql += ") VALUES ("

	values := reflect.ValueOf(data)

	for i, valueIndex := range dataIndexes {
		value := strings.ReplaceAll(values.Field(valueIndex).String(), "\"", "\\\"")
		value = "\"" + value + "\""

		if len(dataIndexes)-1 != i {
			value += ","
		}

		sql += value
	}

	sql += ")"

	res, err := mysqldb.client.Exec(sql)

	if err != nil {
		return 0, err
	}

	lastInsertedId, err := res.LastInsertId()

	if err != nil {
		return 0, err
	}

	return lastInsertedId, nil
}

// FillAutoNestedSingle, writes given data to the given table
// and returns the last inserted id and error if any.
// Create automatically relations.
func (mysqldb *MysqlDB) FillAutoNestedSingle(targetTable string, data interface{}) (int64, error) {
	sql := "INSERT INTO " + targetTable + " ("

	t := reflect.TypeOf(data)
	numField := t.NumField()
	dataIndexes := []int{}

	for i := 0; i < numField; i++ {
		kind := t.Field(i).Type.Kind()

		if kind == reflect.Array || kind == reflect.Slice {
			if sql[len(sql)-1:] == "," {
				sql = sql[:len(sql)-1]
			}
			continue
		}

		tag := t.Field(i).Tag.Get("bulker")

		if numField-1 != i {
			tag += ","
		}

		dataIndexes = append(dataIndexes, i)

		sql += tag
	}

	sql += ") VALUES ("

	values := reflect.ValueOf(data)

	for i, valueIndex := range dataIndexes {
		value := strings.ReplaceAll(values.Field(valueIndex).String(), "\"", "\\\"")
		value = "\"" + value + "\""

		if len(dataIndexes)-1 != i {
			value += ","
		}

		sql += value
	}

	sql += ")"

	res, err := mysqldb.client.Exec(sql)

	if err != nil {
		return 0, err
	}

	lastInsertedId, err := res.LastInsertId()

	if err != nil {
		return 0, err
	}

	return lastInsertedId, nil
}

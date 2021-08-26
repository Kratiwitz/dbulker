package main

import (
	"database/sql"
	"fmt"
	"reflect"

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
		if checkNested(dataType, i) {
			removeLastCommaFromSqlIfAny(&sql)
			continue
		}

		addFieldTagToSql(&sql, &dataIndexes, dataType, i, numField)
	}

	sql += ") VALUES "

	dataLength := len(data)

	for index, v := range data {
		sql += LeftParanthesisToken

		writeValuesToSql(&sql, &v, dataIndexes)

		if dataLength-1 == index {
			sql += RightParanthesisToken
		} else {
			sql += RightParanthesisToken + CommaToken
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

	dataType := reflect.TypeOf(data)
	numField := dataType.NumField()
	dataIndexes := []int{}

	for i := 0; i < numField; i++ {
		if checkNested(dataType, i) {
			removeLastCommaFromSqlIfAny(&sql)
			continue
		}

		addFieldTagToSql(&sql, &dataIndexes, dataType, i, numField)
	}

	sql += RightParanthesisToken + " " + ValuesToken + " " + LeftParanthesisToken
	writeValuesToSql(&sql, &data, dataIndexes)
	sql += RightParanthesisToken

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
// create table for nested objects if any
// and returns the last inserted id and error if any.
// Create automatically relations.
func (mysqldb *MysqlDB) FillAutoNestedSingle(targetTable string, data interface{}) (int64, error) {
	sql := "INSERT INTO " + targetTable + " ("

	dataType := reflect.TypeOf(data)
	numField := dataType.NumField()
	dataIndexes := []int{}

	for i := 0; i < numField; i++ {
		if checkNested(dataType, i) {
			removeLastCommaFromSqlIfAny(&sql)
			continue
		}

		addFieldTagToSql(&sql, &dataIndexes, dataType, i, numField)
	}

	sql += RightParanthesisToken + " " + ValuesToken + " " + LeftParanthesisToken

	writeValuesToSql(&sql, &data, dataIndexes)

	sql += RightParanthesisToken

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

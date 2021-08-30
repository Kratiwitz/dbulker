package main

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type MysqlDB struct {
	ConnectionString string
	Database         string
	client           *sql.DB
}

const (
	InsertTemplate      = "INSERT INTO %s (%s) VALUES %s"
	CreateTableTemplate = "CREATE TABLE IF NOT EXISTS %s (%s)"
	PrimaryKeyTemplate  = "PRIMARY KEY (`%s`)"
)

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
	var columns string
	var values string

	dataType := reflect.TypeOf(data[0])
	numField := dataType.NumField()
	dataIndexes := []int{}

	for i := 0; i < numField; i++ {
		if checkNested(dataType, i) {
			removeLastCommaFromSqlIfAny(&columns)
			continue
		}

		addFieldTagToSql(&columns, &dataIndexes, dataType, i, numField)
	}
	dataLength := len(data)

	for index, v := range data {
		values += LeftParanthesisToken

		writeValuesToSql(&values, &v, dataIndexes)

		if dataLength-1 == index {
			values += RightParanthesisToken
		} else {
			values += RightParanthesisToken + CommaToken
		}
	}

	_, err := mysqldb.client.Exec(fmt.Sprintf(InsertTemplate, targetTable, columns, values))

	if err != nil {
		return err
	}

	return nil
}

// FillAutoPlainSingle, writes given data to the given table
// and returns the last inserted id and error if any.
// Ignore any nested child.
func (mysqldb *MysqlDB) FillAutoPlainSingle(targetTable string, data interface{}) (int64, error) {
	var columns string
	var values string

	dataType := reflect.TypeOf(data)
	numField := dataType.NumField()
	dataIndexes := []int{}

	for i := 0; i < numField; i++ {
		if checkNested(dataType, i) {
			removeLastCommaFromSqlIfAny(&columns)
			continue
		}

		addFieldTagToSql(&columns, &dataIndexes, dataType, i, numField)
	}

	writeValuesToSql(&values, &data, dataIndexes)
	values = LeftParanthesisToken + values + RightParanthesisToken

	res, err := mysqldb.client.Exec(fmt.Sprintf(InsertTemplate, targetTable, columns, values))

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
func (mysqldb *MysqlDB) FillAutoNestedSingle(targetTable string, data interface{}) error {
	var columns string
	var values string
	var haveNested bool

	dataType := reflect.TypeOf(data)
	numField := dataType.NumField()
	dataIndexes := []int{}

	for i := 0; i < numField; i++ {
		if checkNested(dataType, i) {
			haveNested = true
			removeLastCommaFromSqlIfAny(&columns)
			continue
		}

		addFieldTagToSql(&columns, &dataIndexes, dataType, i, numField)
	}

	writeValuesToSql(&values, &data, dataIndexes)
	values = LeftParanthesisToken + values + RightParanthesisToken

	res, err := mysqldb.client.Exec(fmt.Sprintf(InsertTemplate, targetTable, columns, values))

	if err != nil {
		return err
	}

	mainInsertedId, err := res.LastInsertId()

	if err != nil {
		return err
	}

	if !haveNested {
		return nil
	}

	type JunctionTableConfiguration struct {
		TableName   string
		SelfName    string
		MainId      int64
		InsertedIDs []int64
	}

	junctions := []JunctionTableConfiguration{}

	for i := 0; i < numField; i++ {
		if checkNested(dataType, i) {
			field := dataType.Field(i)
			fieldT := field.Type.Elem()
			subFieldNum := field.Type.Elem().NumField()
			fields := make([]string, 0)
			tableName := field.Tag.Get(TagTable)
			subDataIndexes := []int{}

			var subColumns []string

			for j := 0; j < subFieldNum; j++ {
				subField := fieldT.Field(j)
				columnName := subField.Tag.Get(TagBulkerRDB)
				columnSetting := subField.Tag.Get(TagBulkerColumn)
				columnType := subField.Tag.Get(TagBulkerType)

				if columnName == "" {
					columnName = subField.Tag.Get(TagBulker)
				}

				if columnType == "primary" {
					fields = append(fields, fmt.Sprintf(PrimaryKeyTemplate, columnName))
				} else {
					subColumns = append(subColumns, columnName)
					subDataIndexes = append(subDataIndexes, j)
				}

				fields = append(fields, fmt.Sprintf("`%s` %s", columnName, columnSetting))
			}

			createSql := fmt.Sprintf(CreateTableTemplate, tableName, strings.Join(fields, ","))
			mysqldb.client.Exec(createSql)

			mData := reflect.ValueOf(data).Field(i)
			dataLength := mData.Len()
			var insertedIds []int64

			for j := 0; j < dataLength; j++ {
				subValues := LeftParanthesisToken
				subData := mData.Index(j).Interface()

				writeValuesToSql(&subValues, &subData, subDataIndexes)

				subValues += RightParanthesisToken

				insertSql := fmt.Sprintf(InsertTemplate, tableName, strings.Join(subColumns, ","), subValues)

				inRes, err := mysqldb.client.Exec(insertSql)

				if err != nil {
					continue
				}

				lastInserted, err := inRes.LastInsertId()

				if err != nil {
					continue
				}

				insertedIds = append(insertedIds, lastInserted)
			}

			junctions = append(junctions, JunctionTableConfiguration{
				TableName:   field.Tag.Get(TagRelationName),
				SelfName:    tableName,
				MainId:      mainInsertedId,
				InsertedIDs: insertedIds,
			})
		}
	}

	for _, junction := range junctions {
		columns := fmt.Sprintf("`%s_id` INT, `%s_id` INT", targetTable, junction.SelfName)
		createSql := fmt.Sprintf(CreateTableTemplate, junction.TableName, columns)
		_, err := mysqldb.client.Exec(createSql)

		if err != nil {
			return err
		}

		for _, id := range junction.InsertedIDs {
			columns = fmt.Sprintf("`%s_id`, `%s_id`", targetTable, junction.SelfName)
			values := fmt.Sprintf("(%s, %s)", strconv.Itoa(int(junction.MainId)), strconv.Itoa(int(id)))
			insertSql := fmt.Sprintf(InsertTemplate, junction.TableName, columns, values)
			_, err := mysqldb.client.Exec(insertSql)

			if err != nil {
				return err
			}
		}
	}

	return nil
}

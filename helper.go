package main

import (
	"reflect"
	"strings"
)

func slugify(str string) string {
	str = strings.ReplaceAll(str, " ", "_")
	str = strings.ToLower(str)

	return str
}

func checkNested(dataType reflect.Type, index int) bool {
	kind := dataType.Field(index).Type.Kind()

	return kind == reflect.Array || kind == reflect.Slice || len(dataType.Field(index).Tag.Get(TagRelationName)) > 0
}

func removeLastCommaFromSqlIfAny(sql *string) {
	if (*sql)[len((*sql))-1:] == CommaToken {
		(*sql) = (*sql)[:len((*sql))-1]
	}
}

func addFieldTagToSql(sql *string, dataIndexes *[]int, dataType reflect.Type, index int, numField int) {
	tag := dataType.Field(index).Tag.Get(TagBulker)

	if numField-1 != index {
		tag += CommaToken
	}

	(*dataIndexes) = append((*dataIndexes), index)

	(*sql) += tag
}

func writeValuesToSql(sql *string, data *interface{}, dataIndexes []int) {
	values := reflect.ValueOf((*data))

	for i, valueIndex := range dataIndexes {
		value := strings.ReplaceAll(values.Field(valueIndex).String(), "\"", "\\\"")
		value = "\"" + value + "\""

		if len(dataIndexes)-1 != i {
			value += CommaToken
		}

		(*sql) += value
	}
}

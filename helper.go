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
	tag := dataType.Field(index).Tag.Get(TagBulkerRDB)

	if tag == "" {
		tag = dataType.Field(index).Tag.Get(TagBulker)
	}

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

func getStructFieldByTag(tag string, tagName string, t reflect.Type) (reflect.StructField, int) {
	for i := 0; i < t.NumField(); i++ {
		tf := t.Field(i)
		if strings.EqualFold(tf.Tag.Get(tagName), tag) {
			return t.Field(i), i
		}
	}

	return reflect.StructField{}, -1
}

func checkHasRelationTag(t reflect.StructField) bool {
	return len(t.Tag.Get(TagRelationName)) > 0
}

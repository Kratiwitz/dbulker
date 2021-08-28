package main

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type MongoDB struct {
	ConnectionString string
	Database         string
	client           *mongo.Client
	ctx              context.Context
	cancelFunc       context.CancelFunc
}

// NewMongoDBClient create MongoDB client, connect to db and return MongoDB
func NewMongoDBClient(connectionString, database string) (*MongoDB, error) {
	mongoDb := &MongoDB{
		ConnectionString: connectionString,
		Database:         database,
	}

	err := mongoDb.Connect()

	return mongoDb, err
}

// Connect open connection to mongo
func (mongodb *MongoDB) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongodb.ConnectionString))

	if err != nil {
		return err
	}

	err = client.Ping(ctx, readpref.Primary())

	if err != nil {
		return err
	}

	mongodb.client = client
	mongodb.ctx = ctx
	mongodb.cancelFunc = cancel

	return nil
}

// GetAll returns given collection data
func (mongodb *MongoDB) GetAll(collection string, custom interface{}) ([]interface{}, error) {
	cur, err := mongodb.client.Database(mongodb.Database).Collection(collection).Find(context.Background(), bson.D{})

	if err != nil {
		return nil, err
	}

	defer cur.Close(context.Background())

	var data []interface{}

	for cur.Next(context.Background()) {
		var element bson.M
		err = cur.Decode(&element)

		if err != nil {
			return data, err
		}

		customType := reflect.TypeOf(custom)
		customPtr := reflect.New(customType)
		customPtrVal := reflect.Value(customPtr).Elem()

		for k := range element {
			customFieldType, existsCheck := getStructFieldByTag(k, TagBulker, customType)
			f := customPtrVal.FieldByName(customFieldType.Name)

			if checkHasRelationTag(customFieldType) {
				sre := customFieldType.Type

				if sre.Kind() != reflect.Array && sre.Kind() != reflect.Slice {
					return data, fmt.Errorf("relational struct field must be array or slice %v", sre)
				}

				sree := sre.Elem()

				if sree.Kind() != reflect.Struct {
					return data, fmt.Errorf("relational array must contains struct field %v", sre)
				}

				mainField, existCheck := getStructFieldByTag("main", TagBulkerType, sree)

				if existCheck != -1 {
					kind := reflect.TypeOf(element[k]).Kind()
					if kind == reflect.Array || kind == reflect.Slice {
						earr := element[k].(primitive.A)
						for _, e := range earr {
							sreeVPtr := reflect.New(sree)
							sreeV := reflect.Value(sreeVPtr).Elem()
							field := sreeV.FieldByName(mainField.Name)
							field.Set(reflect.ValueOf(e))
							f.Set(reflect.Append(f, sreeV))
						}
					} else {
						f.Set(reflect.ValueOf(element[k]))
					}
				}
			} else if existsCheck != -1 {
				if reflect.TypeOf(element[k]) == reflect.TypeOf(primitive.ObjectID{}) {
					id := element[k].(primitive.ObjectID).String()
					f.Set(reflect.ValueOf(id))
				} else {
					// TODO: check if primitive array for element's item
					f.Set(reflect.ValueOf(element[k]))
				}
			}
		}

		data = append(data, customPtrVal.Interface())

		// typeOf := reflect.TypeOf(custom)

		// fieldCount := typeOf.NumField()

		// for i := 0; i < fieldCount; i++ {
		// 	kind := typeOf.Field(i).Type.Kind()

		// 	if kind == reflect.Array || kind == reflect.Slice {
		// 		childFieldCount := typeOf.Field(i).Type.Elem().NumField()
		// 		mainFieldIsActive := false
		// 		mainFieldIndex := 0

		// 		for j := 0; j < childFieldCount; j++ {
		// 			childTag := typeOf.Field(i).Type.Elem().Field(j).Tag

		// 			if childTag.Get("bulker_type") == "main" {
		// 				mainFieldIsActive = true
		// 				mainFieldIndex = j
		// 			}
		// 		}

		// 		if mainFieldIsActive {
		// 			dummyStruct := reflect.StructField{}
		// 			field := typeOf.Elem().Field(i).
		// 			field = dummyStruct
		// 		}
		// 	}
		// }

		// temp := reflect.New(reflect.TypeOf(custom))
		// err := cur.Decode(inter)
		// data = append(data, temp.Elem().Interface())
	}

	return data, nil
}

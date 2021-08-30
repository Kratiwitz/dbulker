package dbulker

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
			sre := customFieldType.Type

			if checkHasRelationTag(customFieldType) {
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
				elemT := reflect.TypeOf(element[k])

				if elemT == reflect.TypeOf(primitive.ObjectID{}) {
					id := element[k].(primitive.ObjectID).Hex()
					f.SetString(id)
				} else if elemT != reflect.TypeOf(primitive.A{}) {
					f.Set(reflect.ValueOf(element[k]))
				} else {
					earr := element[k].(primitive.A)
					for _, e := range earr {
						sree := sre.Elem()
						sreeVPtr := reflect.New(sree)
						sreeV := reflect.Value(sreeVPtr).Elem()
						sreeV.Set(reflect.ValueOf(e).Convert(sree))
						f.Set(reflect.Append(f, sreeV))
					}
				}
			}
		}

		data = append(data, customPtrVal.Interface())
	}

	return data, nil
}

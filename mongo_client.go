package main

import (
	"context"
	"reflect"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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
		temp := reflect.New(reflect.TypeOf(custom))
		err := cur.Decode(temp.Interface())

		if err != nil {
			return nil, err
		}

		data = append(data, temp.Elem().Interface())
	}

	return data, nil
}

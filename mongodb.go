package cache

import (
	"context"
	"reflect"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/patrickmn/go-cache"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoItem struct {
	Key       string `bson:"_id"`
	ExpiredAt int64  `bson:"expired_at"`
	Value     string `bson:"value"`
}
type MongoDBStore struct {
	client            *mongo.Client
	DefaultExpiration time.Duration
	databaseName      string
	entity            string
	logger            Logger
}

type MongoDBStoreOptions struct {
	DatabaseURI       string
	DatabaseName      string
	Entity            string
	DefaultExpiration time.Duration
	DefaultCacheItems map[string]cache.Item
	CleanupInterval   time.Duration
	Logger            Logger
}

func NewMongoDBStore(opt MongoDBStoreOptions) *MongoDBStore {
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(opt.DatabaseURI))
	if err != nil {
		panic(err)
	}

	var store = &MongoDBStore{
		client:            client,
		DefaultExpiration: opt.DefaultExpiration,
		databaseName:      opt.DatabaseName,
		entity:            opt.Entity,
		logger:            opt.Logger,
	}

	if store.entity == "" {
		store.entity = "caches"
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		store.Logger().Printf("Connect to %s [ERROR] %v\n", err)
	}

	return store
}

func (c *MongoDBStore) getCollection() *mongo.Collection {
	collection := c.client.Database(c.databaseName).Collection(c.entity)
	return collection
}

func (c *MongoDBStore) Get(key string, value interface{}) error {
	if !isPtr(value) {
		c.Logger().Printf("%s: Get key = %s value = %v error %v\n", c.Type(), key, value, ErrMustBePointer)
		return ErrMustBePointer
	}

	var content = mongoItem{}
	var ctx = context.TODO()
	var query = bson.M{"_id": key}
	if err := c.getCollection().FindOne(ctx, query).Decode(&content); err != nil {
		c.Logger().Printf("%s: Get key = %s [ERROR] %v\n", c.Type(), key, err)
		return err
	}

	if content.ExpiredAt <= time.Now().Unix() {
		if _, err := c.getCollection().DeleteOne(ctx, query); err != nil {
			c.Logger().Printf("%s: Delete key = %s [ERROR] %v\n", c.Type(), key, err)
			return err
		}
	}

	var err = jsoniter.UnmarshalFromString(content.Value, value)
	if err != nil {
		c.Logger().Printf("%s: Decode key = %s [ERROR] %v\n", c.Type(), key, err)
		return err
	}

	return nil
}

func (c *MongoDBStore) Set(key string, value interface{}, expiration ...time.Duration) error {
	var v = reflect.ValueOf(value)
	if v.Kind() != reflect.Ptr {
		c.Logger().Printf("%s: Set key = %s value = %v [ERROR] %v\n", c.Type(), key, value, ErrMustBePointer)
		return ErrMustBePointer
	}

	var exp = c.DefaultExpiration
	if len(expiration) > 0 {
		exp = expiration[0]
	}

	bytes, err := jsoniter.Marshal(value)
	if err != nil {
		c.Logger().Printf("%s: Encode key = %s value = %v [ERROR] %v\n", c.Type(), key, v.Interface(), err)
		return err
	}

	var content = mongoItem{
		Key:   key,
		Value: string(bytes),
	}

	if exp > 0 {
		content.ExpiredAt = time.Now().Add(exp).Unix()
	}

	var ctx = context.TODO()
	var query = bson.M{"_id": key}
	var update = bson.M{
		"$set": content,
	}
	result, err := c.getCollection().UpdateOne(ctx, query, &update)
	if err != nil {
		c.Logger().Printf("%s: UpdateOne key = %s value = %v [ERROR] %v\n", c.Type(), key, content, err)
		return err
	}

	c.Logger().Printf("%s: UpdateOne key = %s result = %v\n", c.Type(), key, result)

	if result.MatchedCount == 0 {
		result, err := c.getCollection().InsertOne(ctx, &content)
		if err != nil {
			c.Logger().Printf("%s: InsertOne key = %s value = %v [ERROR] %v\n", c.Type(), key, content, err)
			return err
		}

		c.Logger().Printf("%s: InsertOne key = %s result = %v\n", c.Type(), key, result.InsertedID)

	}

	return nil
}

func (c *MongoDBStore) Delete(key string) error {
	var ctx = context.TODO()
	var query = bson.M{"_id": key}
	if _, err := c.getCollection().DeleteOne(ctx, query); err != nil {
		c.Logger().Printf("%s: DeleteOne key = %s [ERROR] %v\n", c.Type(), key, err)
		return err
	}
	return nil
}

func (c *MongoDBStore) Type() string {
	return "mongodb"
}

func (c *MongoDBStore) Logger() Logger {
	if c.logger != nil {
		return c.logger
	}
	return DefaultLogger
}

package cache

import (
	"context"
	"log"
	"reflect"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/vmihailenco/msgpack/v5"
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
}

type MongoDBStoreOptions struct {
	DatabaseURI       string
	DatabaseName      string
	Entity            string
	DefaultExpiration time.Duration
	DefaultCacheItems map[string]cache.Item
	CleanupInterval   time.Duration
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
	}

	if store.entity == "" {
		store.entity = "caches"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("Connect to %s %v\n", opt.DatabaseURI, err)
	}

	return store
}

func (c *MongoDBStore) getCollection() *mongo.Collection {
	collection := c.client.Database(c.databaseName).Collection(c.entity, &options.CollectionOptions{})
	return collection
}

func (c *MongoDBStore) Get(key string, value interface{}) error {
	if !isPtr(value) {
		return ErrMustBePointer
	}

	var content = mongoItem{}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var query = bson.M{"_id": key}
	if err := c.getCollection().FindOne(ctx, query).Decode(&content); err != nil {
		return err
	}

	if content.ExpiredAt <= time.Now().Unix() {
		if _, err := c.getCollection().DeleteOne(ctx, query); err != nil {
			return err
		}
	}

	var err = msgpack.Unmarshal([]byte(content.Value), value)
	if err != nil {
		return err
	}

	return nil
}

func (c *MongoDBStore) Set(key string, value interface{}, expiration ...time.Duration) error {
	var v = reflect.ValueOf(value)
	if v.Kind() != reflect.Ptr {
		return ErrMustBePointer
	}

	var exp = c.DefaultExpiration
	if len(expiration) > 0 {
		exp = expiration[0]
	}

	bytes, err := msgpack.Marshal(value)
	if err != nil {
		return err
	}

	var content = mongoItem{
		Key:   key,
		Value: string(bytes),
	}

	if exp > 0 {
		content.ExpiredAt = time.Now().Add(exp).Unix()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var query = bson.M{"_id": key}
	var update = bson.M{
		"$set": content,
	}
	result, err := c.getCollection().UpdateOne(ctx, query, &update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		_, err := c.getCollection().InsertOne(ctx, &content)
		if err != nil {
			return err
		}

	}

	return nil
}

func (c *MongoDBStore) Delete(key string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	var query = bson.M{"_id": key}
	if _, err := c.getCollection().DeleteOne(ctx, query); err != nil {
		return err
	}
	return nil
}

func (c *MongoDBStore) Type() string {
	return "mongodb"
}

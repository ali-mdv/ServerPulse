package mongo

import (
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var Client *mongo.Client

func Init(host string, port int, username string, password string) {
	uri := fmt.Sprintf(
		"mongodb://%s:%s@%s:%d/?authSource=admin",
		username,
		password,
		host,
		port,
	)

	mongoClient, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	Client = mongoClient
	log.Println("✅ Connected to MongoDB!")
}

func GetDatabase(DBName string) *mongo.Database {
	return Client.Database(DBName)
}

func GetCollection(DBName string, collectionName string) *mongo.Collection {
	return Client.Database(DBName).Collection(collectionName)
}

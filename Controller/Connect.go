package Controller

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongo() *mongo.Client {
	// 建立連線設定
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	client, err := mongo.Connect(context.TODO(), clientOptions)

	var ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)

	//ping the database
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	return client
}

//Client instance

var DB *mongo.Client = ConnectMongo()

//取得資料庫table

func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	collection := client.Database("MemberDB").Collection(collectionName)
	return collection
}

var MemberInfoCollection = GetCollection(DB, "MemberInfo")

var MemberCollection = GetCollection(DB, "Member")

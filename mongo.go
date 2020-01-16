package main

import (
    "context"
    "fmt"
    "os"
    "strings"

    // "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var dbName string

func connectClient() *mongo.Client {
    mongoURI := os.Getenv("MONGODB_URI")
    if mongoURI == "" {
        mongoURI = "mongodb://localhost:27017"
        dbName = "bot"
    } else {
        splitURI := strings.Split(mongoURI, "/")
        dbName = splitURI[len(splitURI) - 1]
    }

    mongoURI = mongoURI + "?retryWrites=false"

    clientOptions := options.Client().ApplyURI(mongoURI)

    // Connect to MongoDB
    client, err := mongo.Connect(context.TODO(), clientOptions)
    if err != nil {
        fmt.Println(err)
    }

    // Check the connection
    err = client.Ping(context.TODO(), nil)
    if err != nil {
        fmt.Println(err)
    }

    fmt.Println("Connected to db: " + dbName)
    return client
}

// TODO: Move save, delete, update, etc. functionality here

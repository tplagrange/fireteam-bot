package api

import (
    "context"
    "fmt"
    "os"

    "github.com/gin-gonic/gin"

    // "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectDB() {


    mongoURI := os.Getenv("MONGODB_URI")
    if mongoURI == "" {
        mongoURI = "mongodb://localhost:27017"
    }

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

    fmt.Println("Connected to MongoDB.")
}

func GetLoadout(c *gin.Context) {
    fmt.Println("In the backend!")
    // If the user does not have an associated bungie token, respond unauthorized
}

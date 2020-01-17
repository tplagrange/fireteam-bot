package main

import (
    "context"
    "fmt"
    "os"
    "strings"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var dbName    string
var userTable string

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

    // Set table names
    userTable = "users"

    fmt.Println("Connected to db: " + dbName)

    return client
}

// TODO: Move save, delete, update, etc. functionality here

func deleteUser(id string) {
    collection := db.Database(dbName).Collection(userTable)

    // Delete any existing entries for this user
    filter := bson.D{{ "discordid", id}}
    deleteResult, err := collection.DeleteOne(context.TODO(), filter)
    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Println(deleteResult)
    }
}

// Sets and returns the most recently played character id
func updateActiveCharacter(user User) string {
    // If user has no active membership, must update
    if (user.ActiveMembership == "-1") {
        // getActiveMembership()
        return "-1"
    }

    activeCharacter := getActiveCharacter(user.ActiveMembership)

    collection := db.Database(dbName).Collection(userTable)

    filter := bson.M{"discordid": bson.M{"$eq": user.DiscordID}}
    update := bson.M{"$set": bson.M{"activecharacter": activeCharacter}}

    // Call the driver's UpdateOne() method and pass filter and update to it
    _, err := collection.UpdateOne(
        context.Background(),
        filter,
        update,
    )
    if ( err != nil ) {
        fmt.Println(err)
    }

    return activeCharacter
}

func createUser(user User) {
    collection := db.Database(dbName).Collection(userTable)
    insertResult, err := collection.InsertOne(context.TODO(), user)
    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Println(insertResult.InsertedID)
    }
}
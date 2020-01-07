package main

import "go.mongodb.org/mongo-driver/bson/primitive"

type User struct {
    Id           primitive.ObjectID `bson:"_id, omitempty"`
    DiscordID    string `bson:"string field" json:"string field"`
    MembershipID string `bson:"string field" json:"string field"`
    AccessToken  string `bson:"string field" json:"string field"`
    RefreshToken string `bson:"string field" json:"string field"`
}

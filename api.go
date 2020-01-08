package main

import (
    "context"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "os"
    "net/http"
    "net/url"
    "strings"

    "github.com/gin-gonic/gin"
    // "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/bson"
)

type TokenResponse struct {
    Access_token    string
    Token_type      string
    Expires_in      int
    Refresh_token   string
    Refresh_expires_in  int
    Membership_id   string
}

// Handle the redirect URL from Bungie's OAUTH 2.0 Mechanism
func bungieCallback(c *gin.Context) {
    code := c.Query("code")
    state := c.Query("state")

    // Now use the code to receive an access token
    client := &http.Client{}
    data := url.Values{}
    data.Set("grant_type", "authorization_code")
    data.Set("code", code)
    req, _ := http.NewRequest("POST", "https://www.bungie.net/platform/app/oauth/token/", strings.NewReader(data.Encode()))
    req.Header.Add("Authorization", "Basic " + base64.StdEncoding.EncodeToString([]byte(os.Getenv("CLIENT_ID") + ":" + os.Getenv("CLIENT_SECRET"))))
    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    resp, _ := client.Do(req)

    // Assess GetToken Response Code
    if resp.StatusCode == http.StatusOK {
        var tokenResponse TokenResponse
        // This could potentialy be changed to use unmarshalling to save memory
        err := json.NewDecoder(resp.Body).Decode(&tokenResponse)
        // err := json.Unmarshal(resp.Body, &tokenResponse)
        resp.Body.Close()
        if err != nil {
            fmt.Println(err)
        }

        collection := db.Database(dbName).Collection("users")

        // Delete any existing entries for this user
        filter := bson.D{{ "discordid", state}}
        deleteResult, err := collection.DeleteOne(context.TODO(), filter)
        if err != nil {
            fmt.Println(err)
        } else {
            fmt.Println(deleteResult)
        }

        // Collect the available destiny membership id(s) as an array
        req, _ = http.NewRequest("GET", "https://www.bungie.net/platform/User/GetBungieAccount/" + tokenResponse.Membership_id + "/254/", nil)
        req.Header.Add("X-API-Key", os.Getenv("API_KEY"))
        resp, _ = client.Do(req)

        // Assess GetBungieAccount Response Code
        if resp.StatusCode == http.StatusOK {
            destinyMemberships := make([]Membership, 0)

            // Determine which Destiny membership IDs are associated with the Bungie account
            var accountResponse interface{}
            err = json.NewDecoder(resp.Body).Decode(&accountResponse)
            accountMap  := accountResponse.(map[string]interface{})
            responseMap := accountMap["Response"].(map[string]interface{})
            destinyMembershipsArray := responseMap["destinyMemberships"].([]interface{})
            for _, u := range destinyMembershipsArray {
                valuesMap := u.(map[string]interface{})
                tmpMembership := Membership{valuesMap["membershipType"].(float64), valuesMap["membershipId"].(string)}
                destinyMemberships = append(destinyMemberships, tmpMembership)
            }

            // Insert new user entry
            newUser := User{state, destinyMemberships, "-1", tokenResponse.Access_token, tokenResponse.Refresh_token}
            insertResult, err := collection.InsertOne(context.TODO(), newUser)
            if err != nil {
                fmt.Println(err)
            } else {
                fmt.Println(insertResult.InsertedID)
            }
        } else {
            // Error in GetBungieAccount
            fmt.Println(resp.StatusCode)
        }

    } else {
        // Error in GetTokenResponse
        fmt.Println(resp.StatusCode)
    }
}

// Direct the discord user to bungie's OAUTH 2.0 Mechanism
func bungieAuth(c *gin.Context) {
    discordID := c.Param("id")

    bungieAuthURL := "https://www.bungie.net/en/OAuth/Authorize?client_id=" +
                     os.Getenv("CLIENT_ID") +
                     "&response_type=code" +
                     "&state=" + discordID

    c.Redirect(http.StatusMovedPermanently, bungieAuthURL)
}

// Call this function before running any privileged calls
func validate(id string) int {
    collection := db.Database(dbName).Collection("users")
    filter := bson.D{{ "discordid", id}}

    // Check if the api retrieved a discord id
    if (id == "") {
        return 500
    }

    // Check if there is a db entry for the discord id
    var result User
    err := collection.FindOne(context.TODO(), filter).Decode(&result)
    if err != nil {
        fmt.Println(err)
        return 401
    }

    return 200
}

// Return a json object containing the guardian's loadout
func getLoadout(c *gin.Context) {
    discordID  := c.Param("id")
    filter     := bson.D{{ "discordid", discordID}}
    collection := db.Database(dbName).Collection("users")

    var result User
    err := collection.FindOne(context.TODO(), filter).Decode(&result)
    if err != nil {
        fmt.Println(err)
    }

    if (validate(discordID) != 200) {
        c.String(401, "User must register")
        return
    }

    c.String(200, "Found user: " + result.DiscordID)
}

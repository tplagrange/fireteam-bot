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
            destinyMembershipIDs := make([]int, 1)

            // Determine which Destiny membership IDs are associated with the Bungie account
            var accountResponse interface{}
            fmt.Println(accountResponse)
            err = json.NewDecoder(resp.Body).Decode(&accountResponse)
            accountMap  := accountResponse.(map[string]interface{})
            fmt.Println(accountMap)
            responseMap := accountMap["Response"].(map[string]interface{})
            fmt.Println(responseMap)
            destinyMembershipsArray := responseMap["destinyMemberships"].([]interface{})
            fmt.Println(destinyMembershipsArray)
            for _, u := range destinyMembershipsArray {
                valuesMap := u.(map[string]interface{})
                fmt.Println(valuesMap)
                value, ok := valuesMap["membershipId"].(int)
                fmt.Printf("%T\n", value)
                if ok {
                    destinyMembershipIDs = append(destinyMembershipIDs, value)
                } else {
                    fmt.Println("membershipId could not be cast to an int")
                }
            }

            // Insert new user entry
            newUser := User{state, destinyMembershipIDs, tokenResponse.Membership_id, tokenResponse.Access_token, tokenResponse.Refresh_token}
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

    if result.DiscordID == "" {
        c.String(403, "User does not exist")
        return
    }

    c.String(200, "Found user: " + result.DiscordID)
}

package main

import (
    "bytes"
    "context"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "os"
    "net/http"
    "net/url"
    "strconv"
    "strings"
    "sync"
    "time"

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

type ItemSetActionRequest struct {
    ItemIds        []int
    CharacterId    int
    MembershipType int
}

type SafeSlice struct {
    s   []string
    mux sync.Mutex
}

type SafeMap struct {
    m    map[string]int
    mux  sync.Mutex
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

        deleteUser(state)

        // collection := db.Database(dbName).Collection("users")

        // // Delete any existing entries for this user
        // filter := bson.D{{ "discordid", state}}
        // deleteResult, err := collection.DeleteOne(context.TODO(), filter)
        // if err != nil {
        //     fmt.Println(err)
        // } else {
        //     fmt.Println(deleteResult)
        // }

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
            resp.Body.Close()

            accountMap  := accountResponse.(map[string]interface{})
            responseMap := accountMap["Response"].(map[string]interface{})
            destinyMembershipsArray := responseMap["destinyMemberships"].([]interface{})

            activeMembership := "-1"
            for _, u := range destinyMembershipsArray {
                valuesMap := u.(map[string]interface{})


                //////
                ///
                /// For now, we assume PC is the active membership
                activeMembershipType := valuesMap["membershipType"].(float64)
                if ( activeMembershipType == 3 ) {
                    activeMembership = valuesMap["membershipId"].(string)
                    fmt.Println( "Active Membership: " + valuesMap["displayName"].(string) )
                }
                // Replace with getActiveMembership() implementation
                ///
                //////


                tmpMembership := Membership{activeMembershipType, valuesMap["membershipId"].(string)}
                destinyMemberships = append(destinyMemberships, tmpMembership)
            }

            // Empty User Values
            loadouts   := make([]Loadout, 0)

            // Insert new user entry
            newUser := User{loadouts, destinyMemberships, state, activeMembership, "-1", tokenResponse.Access_token, tokenResponse.Refresh_token}
            createUser(newUser)
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
    discordID := c.Query("id")

    fmt.Println(discordID)

    bungieAuthURL := "https://www.bungie.net/en/OAuth/Authorize?client_id=" +
                     os.Getenv("CLIENT_ID") +
                     "&response_type=code" +
                     "&state=" + discordID

    fmt.Println(bungieAuthURL)

    c.Redirect(http.StatusMovedPermanently, bungieAuthURL)
}

func refreshToken(user User) string {
    client := &http.Client{}
    data := url.Values{}
    data.Set("grant_type", "refresh_token")
    data.Set("refresh_token", user.RefreshToken)
    req, _ := http.NewRequest("POST", "https://www.bungie.net/platform/app/oauth/token/", strings.NewReader(data.Encode()))
    req.Header.Add("Authorization", "Basic " + base64.StdEncoding.EncodeToString([]byte(os.Getenv("CLIENT_ID") + ":" + os.Getenv("CLIENT_SECRET"))))
    req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
    resp, _ := client.Do(req)

    defer resp.Body.Close()
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
        filter := bson.M{"discordid": bson.M{"$eq": user.DiscordID}}
        update := bson.M{"$set": bson.M{"accesstoken": tokenResponse.Access_token, "refreshtoken": tokenResponse.Refresh_token}}

        // Call the driver's UpdateOne() method and pass filter and update to it
        _, err = collection.UpdateOne( context.Background(), filter, update )
        if ( err != nil ) {
            fmt.Println(err)
        }

        return tokenResponse.Access_token
    } else {
        return user.AccessToken
    }
}

// Call this function before running any privileged calls
// TODO: Include a check for active token that refreshes the token
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

    // Check if there is an active membership id
    if (result.ActiveMembership == "-1") {
        return 300
    }

    return 200
}

// Returns a lits of shaders (by name) that all members of the party have collected
func getPartyShaders(c *gin.Context) {
    discordID   := c.Query("id")

    result := findUser(discordID)

    switch returnCode := validate(discordID); returnCode {
    case 200:
        ///////// Start Success Condition

        // Get the membership IDs of the party members
        client := &http.Client{}
        reqURL := "https://www.bungie.net/platform/Destiny2/3/Profile/" +
                  result.ActiveMembership +
                  "/?components=1000"
        req, _ := http.NewRequest("GET", reqURL, nil)
        req.Header.Add("X-API-Key", os.Getenv("API_KEY"))

        resp, _ := client.Do(req)

        if resp.StatusCode == http.StatusOK {
        } else {
            c.String(resp.StatusCode, "Error getting profile information")
            return
        }

        var jsonResponse interface{}
        err := json.NewDecoder(resp.Body).Decode(&jsonResponse)
        if err != nil {
            fmt.Println(err)
        }
        resp.Body.Close()

        partyMIDs := make([]string, 0)

        profileMap, ok := jsonResponse.(map[string]interface{})["Response"].(map[string]interface{})["profileTransitoryData"].(map[string]interface{})
        if !ok {
            members := profileMap["data"].(map[string]interface{})["partyMembers"].([]interface{})
            for _, u := range members {
                valuesMap := u.(map[string]interface{})
                partyMIDs = append(partyMIDs, valuesMap["membershipId"].(string))
            }
        } else {
            partyMIDs = append(partyMIDs, result.ActiveMembership)
        }

        // Now we need to get the active character id for every membership ID
        apiQueries := SafeSlice{s: make([]string, 0)}

        var wg sync.WaitGroup
        for _, u := range partyMIDs {
            wg.Add(1)
            go func(wait *sync.WaitGroup) {
                defer wait.Done()
                cid := getActiveCharacter(u)
                apiQueries.mux.Lock()
                apiQueries.s = append(apiQueries.s, u + "/Character/" + cid)
                apiQueries.mux.Unlock()
            }(&wg)
        }
        wg.Wait()


        // Now we have every character ID in the party, we need to get shader information for every character
        shaderHashes := SafeMap{m: make(map[string]int)}
        for _, query := range apiQueries.s {
            wg.Add(1)
            go func(q string, wait *sync.WaitGroup) {
                defer wait.Done()

                shaderURL := "https://www.bungie.net/platform/Destiny2/3/Profile/" +
                          q +
                          "/Collectibles/" +
                          "2063273188/" +
                          "?components=800"
                shaderReq, _ := http.NewRequest("GET", shaderURL, nil)
                shaderReq.Header.Add("X-API-Key", os.Getenv("API_KEY"))
                shaderResp, _ := client.Do(shaderReq)

                if shaderResp.StatusCode == http.StatusOK {
                } else {
                    c.String(shaderResp.StatusCode, "Error getting shader information")
                    return
                }

                var shaderJSON interface{}
                err := json.NewDecoder(shaderResp.Body).Decode(&shaderJSON)
                shaderResp.Body.Close()

                hashData := shaderJSON.(map[string]interface{})["Response"].(map[string]interface{})["collectibles"].(map[string]interface{})["data"].(map[string]interface{})["collectibles"].(map[string]interface{})

                for hash, value := range hashData {
                    // We will track the counts of shader hashes present using a hash
                    // If the hash value int is equal to the number of characters in the party, then everyone has the shader
                    state := value.(map[string]interface{})["state"].(float64)
                    if state != 0 {
                        continue
                    }
                    shaderHashes.mux.Lock()
                    count, ok := shaderHashes.m[hash]
                    shaderHashes.mux.Unlock()
                    if ok {
                        shaderHashes.mux.Lock()
                        shaderHashes.m[hash] = count + 1
                        shaderHashes.mux.Unlock()
                    } else {
                        shaderHashes.mux.Lock()
                        shaderHashes.m[hash] = 1
                        shaderHashes.mux.Unlock()
                    }

                    if ( err != nil ) {
                        fmt.Println(err)
                    }
                }
            }(query, &wg)
        }
        wg.Wait()

        commonHashes  := make([]string, 0)
        numCharacters := len(apiQueries.s)
        for hash, count := range shaderHashes.m {
            if count == numCharacters {
                commonHashes = append(commonHashes, hash)
            }
        }

        shaders := make([]Shader, 0)
        for _, hash := range commonHashes {
            shaders = append(shaders, matchCollectibleHash(hash))
        }

        c.JSON(200, shaders)
        ///////// End Success Condition
    case 300:
        c.String(300, "Please select a membership ID to continue request")
    case 401:
        c.String(401, "User must register")
    default:
        c.String(500, "Unexpected error")
    }
}

// Return a json object containing the guardian's loadout
func getCurrentLoadout(c *gin.Context) {
    discordID   := c.Query("id")
    loadoutName := c.Query("name")

    filter      := bson.D{{ "discordid", discordID}}
    collection  := db.Database(dbName).Collection("users")

    var result User
    err := collection.FindOne(context.TODO(), filter).Decode(&result)
    if err != nil {
        fmt.Println(err)
    }

    switch returnCode := validate(discordID); returnCode {
    // Success Condition
    case 200:
        activeCharacter := updateActiveCharacter(result)

        client := &http.Client{}
        reqURL := "https://www.bungie.net/platform/Destiny2/3/Profile/" +
                  result.ActiveMembership +
                  "/Character/" +
                  activeCharacter +
                  "/?components=205"
        req, _ := http.NewRequest("GET", reqURL, nil)
        req.Header.Add("X-API-Key", os.Getenv("API_KEY"))

        fmt.Println(reqURL)

        resp, _ := client.Do(req)

        defer resp.Body.Close()

        if resp.StatusCode == http.StatusOK {
            // Store Inventory Data for Character

            // If there is already a loadout by that name, update that loadout
            var loadout Loadout
            newLoadout   := true
            loadoutIndex := -1
            for i, u := range result.Loadouts {
                if (u.Name == loadoutName) {
                    newLoadout = false
                    loadoutIndex = i
                }
            }

            loadout = Loadout{make([]Item, 0), loadoutName}

            var jsonResponse interface{}
            err = json.NewDecoder(resp.Body).Decode(&jsonResponse)

            items  := jsonResponse.(map[string]interface{})["Response"].(map[string]interface{})["equipment"].(map[string]interface{})["data"].(map[string]interface{})["items"].([]interface{})

            for _, u := range items {
                valuesMap := u.(map[string]interface{})
                loadout.Items = append(loadout.Items, Item{valuesMap["itemInstanceId"].(string)})
            }

            if (newLoadout) {
                result.Loadouts = append(result.Loadouts, loadout)
            } else {
                result.Loadouts[loadoutIndex] = loadout
            }

            filter := bson.M{"discordid": bson.M{"$eq": result.DiscordID}}
            update := bson.M{"$set": bson.M{"loadouts": result.Loadouts}}

            // Call the driver's UpdateOne() method and pass filter and update to it
            _, err = collection.UpdateOne(
                context.Background(),
                filter,
                update,
            )
            if ( err != nil ) {
                fmt.Println(err)
            }

        }

    case 300:
        c.String(300, "Please select a membership ID to continue request")
    case 401:
        c.String(401, "User must register")
    default:
        c.String(500, "Unexpected error")
    }
}

// Sets a saved loadout for the use
func setLoadout(c *gin.Context) {
    discordID   := c.Query("id")
    loadoutName := c.Param("name")

    filter      := bson.D{{ "discordid", discordID}}
    collection  := db.Database(dbName).Collection("users")

    var user User
    err := collection.FindOne(context.TODO(), filter).Decode(&user)
    if err != nil {
        fmt.Println(err)

    }

    switch returnCode := validate(discordID); returnCode {
    // Success Condition
    case 200:
        var items []int
        activeCharacter := updateActiveCharacter(user)

        for _, l := range user.Loadouts {
            if ( l.Name == loadoutName ) {
                for _, i := range l.Items {
                    itemID, _ := strconv.Atoi(i.Id)
                    items = append(items, itemID)
                }
            }
        }

        ///
        /// Assuming this is on PC
        ///
        mid := 3 // Membership type set to 3; for PC
        cid, _ := strconv.Atoi(activeCharacter)

        loadout := ItemSetActionRequest{
            ItemIds: items,
            CharacterId: cid,
            MembershipType: mid,
        }
        loadoutJSON, err := json.Marshal(loadout)
        if err != nil {
            fmt.Println(err)
        }

        client := &http.Client{}
        reqURL := "http://www.bungie.net/Platform/Destiny2/Actions/Items/EquipItems/"
        req, _ := http.NewRequest("POST", reqURL, bytes.NewBuffer(loadoutJSON))
        req.Header.Add("X-API-Key", os.Getenv("API_KEY"))
        req.Header.Add("Authorization", "Bearer " + user.AccessToken)
        req.Header.Add("Content-Type", "application/json")

        fmt.Println(reqURL)

        resp, _ := client.Do(req)

        switch r := resp.StatusCode; r {
        case 200:
            fmt.Println("Set Loadout to: " + loadoutName)
        case 401:
            // Token is wrong or expired; retry after refreshing
            refreshToken(user)
            reqURL = " http://www.bungie.net/Platform/Destiny2/Actions/Items/EquipItems/"
            req, _ = http.NewRequest("POST", reqURL, bytes.NewBuffer(loadoutJSON))
            req.Header.Add("X-API-Key", os.Getenv("API_KEY"))
            req.Header.Add("Authorization", "Bearer " + user.AccessToken)
            req.Header.Add("Content-Type", "application/json")
            resp, _ = client.Do(req)

            if ( resp.StatusCode == 200) {
                c.String(200, "Set Loadout to: " + loadoutName)
            } else {
                c.String(500, "Error refreshing token")
            }
        default:
            fmt.Println("Error setting loadout to: " + loadoutName)
        }
    default:
        c.String(500, "Error setting loadout")
    }
}

func getActiveMembership() {
    // maybe just grab the most recently active account?
    return
}

// Simply gets and returns a string containing the most recently played character id
func getActiveCharacter(mid string) string {
    var profileResponse interface{}

    // Make GET request to Profile endpoint
    client := &http.Client{}
    reqURL := "https://www.bungie.net/platform/Destiny2/3/Profile/" +
              mid +
              "/?components=200"
    req, _ := http.NewRequest("GET", reqURL, nil)
    req.Header.Add("X-API-Key", os.Getenv("API_KEY"))
    resp, err := client.Do(req)
    if ( err != nil) {
        fmt.Println(err)
    }
    // Parse response json for character ids
    err = json.NewDecoder(resp.Body).Decode(&profileResponse)
    if ( err != nil ) {
        fmt.Println(err)
    }
    resp.Body.Close()

    // Get relevant json data
    responseJSON  := profileResponse.(map[string]interface{})
    responseMap   := responseJSON["Response"].(map[string]interface{})
    characterMap  := responseMap["characters"].(map[string]interface{})["data"].(map[string]interface{})

    activeCharacter := "-1"
    latestDate := time.Time{}

    for k, v := range characterMap {
        dateString := v.(map[string]interface{})["dateLastPlayed"].(string) // e.g. "2020-01-09T06:11:35Z"
        date, _    := time.Parse(
            time.RFC3339,
            dateString)
        if (date.After(latestDate)) {
            activeCharacter = k
            latestDate = date
        }
    }

    return activeCharacter
}

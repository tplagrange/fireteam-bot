package main

import (
	"fmt"
	"encoding/json"
	"net/http"
	"os"
)

func getManifest() map[string]interface{} {
    client := &http.Client{}
    req, _ := http.NewRequest("GET", "https://www.bungie.net/Platform/Destiny2/Manifest/", nil)
    req.Header.Add("X-API-Key", os.Getenv("API_KEY"))
    resp, _ := client.Do(req)

    if resp.StatusCode != http.StatusOK {
            fmt.Println("Error getting manifest")
            return make(map[string]interface{})
    }

    var jsonResponse interface{}
    err := json.NewDecoder(resp.Body).Decode(&jsonResponse)
    if err != nil {
        fmt.Println(err)
    }
    resp.Body.Close()

    return jsonResponse.(map[string]interface{})["Response"].(map[string]interface{})
}

// Returns a json object containing information about the matched shader
func matchShaderHash(hash string) string {
	// Check against the db to find a match for the hash

	// If no match is found, update the db and try again
	// If still not match, return an error
	return "-1"
}

func getShaderHashes() {
	manifest := getManifest()

	resource := manifest["jsonWorldComponentContentPaths"].(map[string]interface{})["en"].(map[string]interface{})["DestinyCollectibleDefinition"].(string)
	url := "https://www.bungie.net" + resource 

    client := &http.Client{}
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Add("X-API-Key", os.Getenv("API_KEY"))
    resp, _ := client.Do(req)

    if resp.StatusCode != http.StatusOK {
            fmt.Println("Error getting shader hash data")
    }

    var jsonResponse interface{}
    err := json.NewDecoder(resp.Body).Decode(&jsonResponse)
    if err != nil {
        fmt.Println(err)
    }
    resp.Body.Close()

	updateShaderHashes(jsonResponse)
}
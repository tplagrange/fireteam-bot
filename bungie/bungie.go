package bungie

import (
    "os"
)

authURL := "https://www.bungie.net/en/OAuth/Authorize?client_id=30633&response_type=code"

func AuthenticateUser(client *Client) {
    // Add Bungie Token to Header
    setHeaders(client)
    // Get Access Token for User
    token =
    // Set Access Token for User
    client.SetAuthToken(token)
}

func setHeaders(client *Client) {
    client.SetHeader("X-API-KEY", os.Getenv("BUNGIE_TOKEN"))
}

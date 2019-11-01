package bungie

import (
    "os"

    "github.com/go-resty/resty/v2"
)

func AuthenticateUser(client *Client) {
    setHeaders(client)
}

func setHeaders(client *Client) {
    client.SetHeader("X-API-KEY", os.Getenv("BUNGIE_TOKEN"))
}

Key: X-API-KEY
Value: {paste your API key here}
Description: Destiny

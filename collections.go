package main

type Membership struct {
    MembershipType  float64
    MembershipID    string
}

type User struct {
    DiscordID         string
    Membership        []Membership
    ActiveMembership  string
    ActiveCharacter   string
    AccessToken       string
    RefreshToken      string
}

// type Manifests struct {
//     DestinyInventoryBucketDefinition    string
// }

type Item struct {
    name    string
    id      string
    type    string
}

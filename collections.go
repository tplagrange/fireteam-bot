package main

type Loadout struct {
    Items   []Item
    Name    string
    id      uint32
}

type Membership struct {
    MembershipID    float64
    MembershipType  string
}

type User struct {
    Loadouts          []Loadout
    Membership        []Membership
    DiscordID         string
    ActiveMembership  string
    ActiveCharacter   string
    AccessToken       string
    RefreshToken      string
}

// type Manifests struct {
//     DestinyInventoryBucketDefinition    string
// }

type Item struct {
    Name    string
    Id      string
    Type    string
}

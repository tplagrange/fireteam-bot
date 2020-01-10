package main

type Item struct {
    Id      string
}

type Loadout struct {
    Items   []Item
    Name    string
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

package main

type Membership struct {
    MembershipType  float64
    MembershipID    string
}

type User struct {
    DiscordID         string
    Membership        []Membership
    ActiveMembership  string
    AccessToken       string
    RefreshToken      string
}

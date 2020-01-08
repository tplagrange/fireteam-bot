package main

type Membership struct {
    MembershipType  int
    MembershipID    string
}

type User struct {
    DiscordID         string
    Membership        []Membership
    ActiveMembership  string
    AccessToken       string
    RefreshToken      string
}

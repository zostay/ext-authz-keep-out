package main

import (
	"flag"
)

var (
	realm, user, pass string
)

func init() {
	flag.StringVar(&realm, "realm", "Invitation Only", "set the name of the Basic realm to present")
	flag.StringVar(&user, "user", "demo", "set the name of the user to accept")
	flag.StringVar(&pass, "pass", "demo", "set the password of the user to accept")
}

func main() {
	flag.Parse()
	ko := NewKeepOut(realm, user, pass)
	ko.Run("0.0.0.0:8080", nil)
}

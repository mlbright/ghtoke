package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"flag"
	"github.com/google/go-github/github"
	"golang.org/x/crypto/ssh/terminal"
	"math/rand"
	"net/url"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {

	note := flag.String("note", fmt.Sprintf("ghtoken: %s", randStr(6)), "A note to remind you what the OAuth token is for.")
	scopes := flag.String("scopes", "", "A list of scopes that this authorization is in.")
	flag.Parse()

	r := bufio.NewReader(os.Stdin)
	fmt.Print("GitHub Username: ")
	username, _ := r.ReadString('\n')

	fmt.Print("GitHub Password: ")
	bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
	password := string(bytePassword)

	tp := github.BasicAuthTransport{
		Username: strings.TrimSpace(username),
		Password: strings.TrimSpace(password),
	}

	client := github.NewClient(tp.Client())

	if os.Getenv("GITHUB_API") != "" {
		url, _ := url.Parse(os.Getenv("GITHUB_API"))
		client.BaseURL = url
	}

	ctx := context.Background()

	authRequest := &github.AuthorizationRequest{}
	if *scopes != "" {
		for _, scope := range strings.Split(*scopes, ",") {
			fmt.Sprintf("scope: %s", scope)
			authRequest.Scopes = append(authRequest.Scopes, github.Scope(scope))
		}
	}
	authRequest.Note = note

	authorization, _, err := client.Authorizations.Create(ctx, authRequest)

	// Is this a two-factor auth error? If so, prompt for OTP and try again.
	if _, ok := err.(*github.TwoFactorAuthError); ok {
		fmt.Print("\nGitHub OTP: ")
		otp, _ := r.ReadString('\n')
		tp.OTP = strings.TrimSpace(otp)
		authorization, _, err = client.Authorizations.Create(ctx, authRequest)
	}

	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return
	}

	fmt.Printf("\nnew token: %v\n", authorization.GetToken())
}

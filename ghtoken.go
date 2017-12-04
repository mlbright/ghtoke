// Copyright 2015 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The basicauth command demonstrates using the github.BasicAuthTransport,
// including handling two-factor authentication. This won't currently work for
// accounts that use SMS to receive one-time passwords.
package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/google/go-github/github"
	"golang.org/x/crypto/ssh/terminal"
		"net/url"
)

func main() {
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
	note := "Get me a token!"
	authRequest.Note = &note

	authorization, _, err := client.Authorizations.Create(ctx,authRequest) 

	// Is this a two-factor auth error? If so, prompt for OTP and try again.
	if _, ok := err.(*github.TwoFactorAuthError); ok {
		fmt.Print("\nGitHub OTP: ")
		otp, _ := r.ReadString('\n')
		tp.OTP = strings.TrimSpace(otp)
		authorization, _, err = client.Authorizations.Create(ctx,authRequest) 
	}

	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return
	}

	fmt.Printf("\nnew token: %v\n", authorization.GetToken())
}

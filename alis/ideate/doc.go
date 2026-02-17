/*
Package ideate provides a client for interacting with the Alis Ideate API.
This library provides type definitions and client stubs to easily integrate Alis Ideate features into your Go applications.

# Installation

	go get github.com/alis-exchange/ideate-go

# Usage

Here is a simple example of how to create a client and make a request to add a note to an idea using a collection token.

	package main

	import (
		"context"
		"log"

		"google.golang.org/grpc/metadata"

		"github.com/alis-exchange/ideate-go/alis/ideate"
	)

	func main() {
		ctx := context.Background()

		// 1. Establish a new client
		client, err := ideate.NewClient(ctx)
		if err != nil {
			log.Fatalf("failed to create client: %v", err)
		}

		// 2. Prepare the context with authentication
		// TODO: Replace with your actual user access token.
		// See the "Security Requirements" section below for details on obtaining a token.
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer <USER_ACCESS_TOKEN>")

		// 3. Define the target (e.g., using a Collection Token generated in Ideate)
		token := "<COLLECTION_TOKEN>"

		// 4. Make a request
		// In this example, we are adding a note to the stream identified by the token.
		_, err = client.AddNote(ctx, &ideate.AddNoteRequest{
			Content: "Hello, world!",
			StreamTarget: &ideate.AddNoteRequest_Token{
				Token: token,
			},
		})
		if err != nil {
			log.Fatalf("failed to add note: %v", err)
		}

		log.Println("Successfully added note.")
	}

# Security Requirements

## OAuth Client Registration

To ensure secure API interactions, you must register a new application with Alis:

 1. Log in to the Alis Identity Management System (https://identity.alisx.com/apps).
 2. Click "New app".
 3. Complete the registration steps and securely store your Client ID and Client Secret.
 4. Configure the Redirect URI to handle the OAuth callback.

## OAuth Flow

Authenticate users and obtain an access token using the standard OAuth 2.0 Authorization Code flow:

 1. Authorize: Redirect the user to the authorization endpoint:

    https://identity.alisx.com/authorize?client_id=<CLIENT_ID>&redirect_uri=<REDIRECT_URI>

 2. Grant Access: The user logs in and approves your application.

 3. Callback: The user is redirected to your <REDIRECT_URI> with an "?code=..." parameter.

 4. Exchange: Swap this authorization code for an Access Token and Refresh Token.

Include the access token in the "Authorization" header of your gRPC calls as shown in the Usage example.

# Storing Tokens

It is your responsibility to handle the access and refresh tokens in the way you want. We recommend two patterns for storing tokens, described below.

## Pattern 1: Cookie Storage

This pattern is ideal for web applications where you want to manage user sessions securely. The flow involves redirecting the user to Alis Identity for authentication, handling the callback to exchange an authorization code for tokens, and storing those tokens in secure, HTTP-only cookies.

### The Flow

 1. Initiate Sign-In: The user is redirected to the Alis Identity authorize endpoint.
 2. User Authentication: The user logs in and grants permission to your application.
 3. Authorization Callback: The user is redirected back to your application with a temporary authorization code.
 4. Token Exchange: Your server exchanges the code for an Access Token and a Refresh Token.
 5. Secure Storage: The tokens are stored as cookies in the user's browser.
 6. Token Refresh: When the Access Token expires, the Refresh Token is used to obtain a new one without requiring the user to log in again.

### Implementation Example

The following example demonstrates how to implement these endpoints using Go's standard library.

	package main

	import (
		"bytes"
		"encoding/json"
		"fmt"
		"net/http"
		"time"
	)

	const (
		AlisIdentityHost         = "https://identity.alisx.com"
		IdeateClientID           = "<YOUR_CLIENT_ID>"
		IdeateClientSecret       = "<YOUR_CLIENT_SECRET>"
		IdeateAccessTokenCookie  = "ideate_access_token"
		IdeateRefreshTokenCookie = "ideate_refresh_token"
	)

	func main() {
		// Endpoint to start the OAuth flow
		http.HandleFunc("/auth/ideate/signin", IdeateSignin)
		// Endpoint for the OAuth callback
		http.HandleFunc("/auth/ideate/callback", IdeateCallback)
		// Endpoint to refresh the access token
		http.HandleFunc("/auth/ideate/refresh", IdeateRefresh)

		// Start your server...
	}

	// IdeateSignin redirects the user to Alis Identity.
	// Trigger this endpoint when a user clicks "Sign in with Alis".
	func IdeateSignin(w http.ResponseWriter, r *http.Request) {
		// Construct the redirect URL for the OAuth flow
		redirectURI := fmt.Sprintf("https://%s/auth/ideate/callback", r.Host)
		authorizeURL := fmt.Sprintf("%s/authorize?client_id=%s&redirect_uri=%s",
			AlisIdentityHost, IdeateClientID, redirectURI)

		// Redirect the user to Alis Identity
		http.Redirect(w, r, authorizeURL, http.StatusTemporaryRedirect)
	}

	// IdeateCallback handles the redirection from Alis Identity after the user authenticates.
	// It receives an authorization code and exchanges it for access and refresh tokens.
	func IdeateCallback(w http.ResponseWriter, r *http.Request) {
		// Extract the authorization code from the query parameters
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "missing code query param", http.StatusBadRequest)
			return
		}

		// Exchange the code for tokens using the identity service
		tokens, err := getIdeateTokens(r, "authorization_code", code)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Store the received tokens in secure cookies
		tokens.WriteAsCookies(w)

		// Redirect the user to the application's main page
		http.Redirect(w, r, "/ideate", http.StatusTemporaryRedirect)
	}

	// IdeateRefresh uses the refresh token stored in the cookie to obtain a new access token.
	// Trigger this from your client when an Ideate API call returns an unauthenticated error.
	func IdeateRefresh(w http.ResponseWriter, r *http.Request) {
		// Retrieve the refresh token from the cookie
		cookie, err := r.Cookie(IdeateRefreshTokenCookie)
		if err != nil {
			http.Error(w, "missing refresh token cookie", http.StatusUnauthorized)
			return
		}

		// Request new tokens using the refresh token
		tokens, err := getIdeateTokens(r, "refresh_token", cookie.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Update the cookies with the new tokens
		tokens.WriteAsCookies(w)
		w.WriteHeader(http.StatusOK)
	}

	type IdeateTokens struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	// WriteAsCookies helper to save tokens into HTTP-only cookies.
	func (t *IdeateTokens) WriteAsCookies(w http.ResponseWriter) {
		http.SetCookie(w, &http.Cookie{
			Name:     IdeateAccessTokenCookie,
			Value:    t.AccessToken,
			Path:     "/",
			HttpOnly: true, // Recommended for security
			MaxAge:   int((7 * 24 * time.Hour).Seconds()),
		})
		http.SetCookie(w, &http.Cookie{
			Name:     IdeateRefreshTokenCookie,
			Value:    t.RefreshToken,
			Path:     "/",
			HttpOnly: true, // Recommended for security
			MaxAge:   int((7 * 24 * time.Hour).Seconds()),
		})
	}

	// getIdeateTokens makes a POST request to the Alis Identity token endpoint.
	func getIdeateTokens(r *http.Request, grantType string, grant string) (*IdeateTokens, error) {
		type Body struct {
			GrantType    string `json:"grant_type"`
			Code         string `json:"code,omitempty"`
			RefreshToken string `json:"refresh_token,omitempty"`
			ClientID     string `json:"client_id"`
			ClientSecret string `json:"client_secret"`
			RedirectURI  string `json:"redirect_uri,omitempty"`
		}

		body := &Body{
			GrantType:    grantType,
			ClientID:     IdeateClientID,
			ClientSecret: IdeateClientSecret,
		}

		if grantType == "authorization_code" {
			body.Code = grant
			body.RedirectURI = fmt.Sprintf("https://%s/auth/ideate/callback", r.Host)
		} else {
			body.RefreshToken = grant
		}

		jsonBody, _ := json.Marshal(body)
		resp, err := http.Post(AlisIdentityHost+"/token", "application/json", bytes.NewBuffer(jsonBody))
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("identity server error: %s", resp.Status)
		}

		var tokens IdeateTokens
		if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
			return nil, err
		}
		return &tokens, nil
	}

### When to trigger these endpoints

  - /signin: Link your "Login" button to this endpoint.
  - /callback: This is your registered redirect URI. It handles the logic after user approval.
  - /refresh: Call this from your frontend or client-side logic whenever a request to Ideate fails with an authentication error (e.g., gRPC code Unauthenticated).

## Pattern 2: Alis Build Identity Connectors

### Prerequisites
This pattern requires your application to be built on the Alis Build platform, and requires the usage of the [Users Management Block](https://console.alisx.com/build/blocks/users/documentation).

...
*/
package ideate

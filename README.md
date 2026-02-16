# ideate-go

[![Go Reference](https://pkg.go.dev/badge/github.com/alis-exchange/ideate-go/alis/ideate.svg)](https://pkg.go.dev/github.com/alis-exchange/ideate-go/alis/ideate)
[![License](https://img.shields.io/github/license/alis-exchange/ideate-go)](LICENSE)

Public Go client for interacting with the **Alis Ideate** API. This library provides type definitions and client stubs to easily integrate Alis Ideate features into your Go applications.

## 🚀 Installation

```bash
go get github.com/alis-exchange/ideate-go
```

## 📚 Usage

Here is a simple example of how to create a client and make a request to add a note to an idea using a collection token.

```go
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
```

## 🔐 Security Requirements

### OAuth Client Registration

To ensure secure API interactions, you must register a new application with Alis:

1.  Log in to the [Alis Identity Management System](https://identity.alisx.com/apps).
2.  Click **New app**.
3.  Complete the registration steps and securely store your **Client ID** and **Client Secret**.
4.  Configure the **Redirect URI** to handle the OAuth callback.

### OAuth Flow

Authenticate users and obtain an access token using the standard OAuth 2.0 Authorization Code flow:

1.  **Authorize**: Redirect the user to the authorization endpoint:
    ```
    https://identity.alisx.com/authorize?client_id=<CLIENT_ID>&redirect_uri=<REDIRECT_URI>
    ```
2.  **Grant Access**: The user logs in and approves your application.
3.  **Callback**: The user is redirected to your `<REDIRECT_URI>` with an `?code=...` parameter.
4.  **Exchange**: Swap this authorization code for an **Access Token** and **Refresh Token**.

Include the access token in the `Authorization` header of your gRPC calls as shown in the Usage example.

## 🤝 Contributing

Contributions are welcome! Please feel free to open issues or submit pull requests.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

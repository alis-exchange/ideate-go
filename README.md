# ideate-go

Public type definitions and client stubs to interact with Alis Ideate

## 🚀 Installation

```bash
go get github.com/ideate/ideate-go
```

## Usage

```go
package main

import (
    "context"
    "log"

    "github.com/ideate/ideate-go/alis/ideate"
)

func main() {
    ctx := context.Background()

    // Establish a new client
    client, err := ideate.NewClient(ctx)
    if err != nil {
        log.Fatalf("failed to create client: %v", err)
    }

    ctx := context.Background()
    // TODO: add valid access token to outgoing requests. See the security requirements section for more information
    ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "<USER_ACCESS_TOKEN>")


    // The collection token
    // Generated in Ideate
    token := "<COLLECTION_TOKEN>"

    // Make a request
    response, err := client.AddNote(ctx, &ideate.AddNoteRequest{
        Content: "Hello, world!",
        StreamTarget: &ideate.AddNoteRequest_Token{Token: token}
    })
    if err != nil {
        log.Fatalf("failed to add note: %v", err)
    }

}
```

## Security requirements

### OAuth Client Registration

To ensure that API calls are made securely, we require that you register a new application.

1. Open the [Alis Identity Management System](https://identity.alisx.com/apps) and sign in.
2. Click on "New app"
3. Go through the various steps to register your application, ensuring you copy the ClientID and ClientSecret.

Once completed, you will be able to trigger the OAuth flow directly from your application.

Make sure to setup the redirect URI of your application to handle the callback from the OAuth flow.

### OAuth Flow

1. Trigger the OAuth flow by redirecting the user to the `https://identity.alisx.com/authorize?client_id=<CLIENT_ID>&redirect_uri=<REDIRECT_URI>` endpoint, where `<CLIENT_ID>` is the ClientID of your application and `<REDIRECT_URI>` is your application's endpoint that will handle the callback.
2. The user will be prompted to log in and grant access to your application.
3. The user will be redirected back to your application with an authorization code, at the `<REDIRECT_URI>` endpoint.
4. Exchange the authorization code for an access token and refresh token.

In order for successful calls to be made to the Ideate API, you will need to provide the access token in the `Authorization` header of your requests.

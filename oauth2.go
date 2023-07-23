package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const missingClientSecretsMessage = `
Please configure OAuth 2.0
To make this sample run, you need to populate the client_secrets.json file
found at:
   %v
with information from the {{ Google Cloud Console }}
{{ https://cloud.google.com/console }}
For more information about the client_secrets.json file format, please visit:
https://developers.google.com/api-client-library/python/guide/aaa_client_secrets
`

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(scope string, userCredentials string) *http.Client {
	ctx := context.Background()

	filePath := filepath.Join(".", userCredentials)
	b, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Unable to read client secret file '%s': %s", filePath, err.Error())
	}

	// If modifying the scope, delete your previously saved credentials
	// at ~/.credentials/youtube-go.json
	config, err := google.ConfigFromJSON(b, scope, "https://www.googleapis.com/auth/calendar.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file '%s' to config: %s", filePath, err.Error())
	}

	// Use a redirect URI like this for a web app. The redirect URI must be a
	// valid one for your OAuth2 credentials.
	//config.RedirectURL = "http://localhost"
	// Use the following redirect URI if launchWebServer=false in oauth2.go
	config.RedirectURL = "urn:ietf:wg:oauth:2.0:oob"

	var tok *oauth2.Token
	if update {
		tok = updateToken(config)
		saveToken("token.json", tok)
	} else {
		tok, err = tokenFromFile("token.json")
		if err != nil {
			log.Fatalf("cannot get token '%s'. Program require manually run for credentials file with -update argument", "credentials.json")
		}
	}

	tok, isUpdated := refreshToken(config, tok)
	if isUpdated {
		saveToken("token.json", tok)
	}

	if !tok.Valid() {
		log.Fatalf("token '%s' is expired. Program require manually run for credentials file with -update argument", "credentials.json")
	}

	return config.Client(ctx, tok)
}

func refreshToken(config *oauth2.Config, tok *oauth2.Token) (*oauth2.Token, bool) {
	ctx := context.Background()

	// will be expired so let's update it
	if tok.Expiry.Add(time.Duration(-5) * time.Minute).Before(time.Now()) {
		src := config.TokenSource(ctx, tok)
		newToken, err := src.Token() // this actually goes and renews the tokens
		if err != nil {
			log.Fatal(err)
		}

		return newToken, true
	}

	return tok, false
}

func updateToken(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	tok, err := getTokenFromPrompt(config, authURL)
	if err != nil {
		log.Fatal(err)
	}

	return tok
}

// Exchange the authorization code for an access token
func exchangeToken(config *oauth2.Config, code string) (*oauth2.Token, error) {
	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token %v", err)
	}
	return tok, nil
}

// getTokenFromPrompt uses Config to request a Token and prompts the user
// to enter the token on the command line. It returns the retrieved Token.
func getTokenFromPrompt(config *oauth2.Config, authURL string) (*oauth2.Token, error) {
	var code string
	fmt.Printf("Go to the following link in your browser. After completing "+
		"the authorization flow, enter the authorization code on the command "+
		"line: \n%v\n", authURL)

	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}
	fmt.Println(authURL)
	return exchangeToken(config, code)
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	if update {
		fmt.Println("trying to save token")
		fmt.Printf("Saving credential file to: %s\n", file)
	}
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0660)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

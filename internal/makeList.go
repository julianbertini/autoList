package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

// IDsep is what separates id values
const IDsep string = "-"

// CategorySep delineates betweeen categories
const CategorySep string = "###"

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.

	// DEBUGGING file path
	// tokFile := "../configs/token.json"
	tokFile := "configs/token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func getHeaders(respValues [][]interface{}) *map[string]map[string][2]int {

	var coords [2]int
	var key string
	var valueStr string
	headersMap := make(map[string]map[string][2]int)

	for i, row := range respValues {

		if len(row) > 0 {
			rowStr := row[0].(string)
			if rowStr == CategorySep {
				for col, value := range row {
					valueStr = value.(string)

					if valueStr == CategorySep {
						valueStr = row[col+1].(string)
						key = strings.Split(valueStr, IDsep)[1]
						headers := make(map[string][2]int)
						headersMap[key] = headers
					} else {
						coords[0] = i
						coords[1] = col
						headersMap[key][valueStr] = coords
					}
				}
			}
		}
	}

	return &headersMap
}

func main() {

	// DEBUGGING file path
	// b, err := ioutil.ReadFile("../configs/credentials.json")
	b, err := ioutil.ReadFile("configs/credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(config)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	spreadsheetID := "1sLGhnXgDLig0HTJE4_uQ6OXXc4iQO4U5QdqWGfyyBck"
	readRange := "Recipes!A:Z"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()

	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	// need to identify indices of headers

	headersMap := *getHeaders(resp.Values)
	fmt.Println(headersMap)

	// if len(resp.Values) == 0 {
	// 	fmt.Println("No data found.")
	// } else {
	// 	fmt.Println("Name, Major:")
	// 	for _, row := range resp.Values {
	// 		// Print columns A and E, which correspond to indices 0 and 4.
	// 		fmt.Printf("%s; %s\n", row[1], row[6])
	// 	}
	// }
}

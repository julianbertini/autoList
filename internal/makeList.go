package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/sheets/v4"
)

//QuantitySep denotes separation between measurement quant. and type
const QuantitySep string = "/"

//MeasurementSep denotes the end of ingredient and start of measurement
const MeasurementSep string = ";"

//IngredientSep denotes the end of ingredient and start of measurement
const IngredientSep string = ","

// Ingredients name of ingredients column
const Ingredients string = "Ingredients"

// IDprefix ID- which predeces every ID number
const IDprefix string = "ID-"

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
	tokFile := "../configs/token.json"
	// tokFile := "configs/token.json"
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

type unitConversions struct {
	Tsp  map[string]float64 `json:"tsp"`
	Tbsp map[string]float64 `json:"tbsp"`
	Lb   map[string]float64 `json:"lb"`
	Oz   map[string]float64 `json:"oz"`
	Floz map[string]float64 `json:"floz"`
	Cups map[string]float64 `json:"cups"`
}

func loadUnitConversions() *unitConversions {
	// DEBUGGING
	f, err := os.Open("../configs/measurements.json")
	// f, err := os.Open("/configs/measurements.json")
	if err != nil {
		log.Fatalf("Unable to open measurements.json file. Error: %v", err)
	}

	defer f.Close()

	unitConversions := &unitConversions{}
	err = json.NewDecoder(f).Decode(unitConversions)
	if err != nil {
		log.Fatalf("Unable to enconde measurements.json file into struct. Error: %v", err)
	}
	return unitConversions
}

func convertUnit(fromQuant string, fromUnit string, toUnit string, units *unitConversions) float64 {

	var newQuant float64
	floatQuant, _ := strconv.ParseFloat(fromQuant, 32)

	switch fromUnit {
	case "tsp":
		if _, ok := units.Tsp[toUnit]; !ok {
			log.Fatalf("Unit conversion not found from " + fromUnit + " to " + toUnit)
		}
		newQuant = floatQuant * units.Tsp[toUnit]
	case "tbsp":
		if _, ok := units.Tbsp[toUnit]; !ok {
			log.Fatalf("Unit conversion not found from " + fromUnit + " to " + toUnit)
		}
		newQuant = floatQuant * units.Tbsp[toUnit]
	case "lb":
		if _, ok := units.Lb[toUnit]; !ok {
			log.Fatalf("Unit conversion not found from " + fromUnit + " to " + toUnit)
		}
		newQuant = floatQuant * units.Lb[toUnit]
	case "oz":
		if _, ok := units.Oz[toUnit]; !ok {
			log.Fatalf("Unit conversion not found from " + fromUnit + " to " + toUnit)
		}
		newQuant = floatQuant * units.Oz[toUnit]
	case "floz":
		if _, ok := units.Floz[toUnit]; !ok {
			log.Fatalf("Unit conversion not found from " + fromUnit + " to " + toUnit)
		}
		newQuant = floatQuant * units.Floz[toUnit]
	case "cups":
		if _, ok := units.Cups[toUnit]; !ok {
			log.Fatalf("Unit conversion not found from " + fromUnit + " to " + toUnit)
		}
		newQuant = floatQuant * units.Cups[toUnit]
	default:
		log.Fatalf("From unit: " + fromUnit + " is not currently supported.")
	}

	return newQuant

}

func addIngredientsToList(ingredients []string, units *unitConversions, groceryMap map[string][]string) {

	var fromUnit string
	var toUnit string
	var fromQuantFloat float64
	// ["Onion;1/cups"]
	for _, ingredient := range ingredients {
		newIngredientTokens := strings.Split(ingredient, MeasurementSep)            // ["Onion", "1/cups"]
		newIngredientName := newIngredientTokens[0]                                 // "Onion"
		newIngredientMeasurment := newIngredientTokens[1]                           // "1/cups"
		newMeasurementTokens := strings.Split(newIngredientMeasurment, QuantitySep) // ["1", "cups"]
		fromQuant := newMeasurementTokens[0]                                        // "1"
		fromQuantFloat, _ = strconv.ParseFloat(fromQuant, 64)
		fromUnit = ""
		toUnit = ""

		existingMeasurementTokens, ok := groceryMap[newIngredientName]

		if len(newMeasurementTokens) > 1 && len(newMeasurementTokens[1]) > 0 {
			fromUnit = newMeasurementTokens[1]
		}
		if len(existingMeasurementTokens) > 1 && len(existingMeasurementTokens[1]) > 0 {
			toUnit = existingMeasurementTokens[1] // "cups"
		}

		if ok { // 1: Check if the ingredient is already in groveryMap. It is.
			if len(fromUnit) > 0 && len(toUnit) > 0 && fromUnit != toUnit {
				fromQuantFloat = convertUnit(fromQuant, fromUnit, toUnit, units)
			}

			//Convert the new ingredient into the unit of existing ingredient

			// Convert existing ingredient amount to float
			toQuant, _ := strconv.ParseFloat(groceryMap[newIngredientName][0], 64)
			// Add existing ingredient amount to new amount, and convert to str.
			newQuant := strconv.FormatFloat(fromQuantFloat+toQuant, 'f', 2, 64)
			// Add new ingredient quantity to existing ingredient in list
			groceryMap[newIngredientName][0] = newQuant

		} else { // 2: Check if the ingredient is already in groceryMap. It is not.
			// Add the ingredient and amount to the map
			measurement := []string{fromQuant, fromUnit}
			groceryMap[newIngredientName] = measurement
		}
	}
}

func getIngredients(respValues [][]interface{}, recipeID string, headersMap map[string]map[string][2]int) []string {

	var ingredients string
	var category string
	var coords [2]int

	category = strings.Split(recipeID, IDsep)[1]
	recipeNum, _ := strconv.Atoi(strings.Split(recipeID, IDsep)[0])

	coords = headersMap[category][Ingredients]
	coords[0] = coords[0] + recipeNum

	ingredients = respValues[coords[0]][coords[1]].(string)

	return strings.Split(ingredients, IngredientSep)
}

func addRecipeToList(respValues [][]interface{}, recipeMap map[string]string, recipeID string) {

}

func getHeaders(respValues [][]interface{}) map[string]map[string][2]int {

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

	return headersMap
}

func main() {

	// DEBUGGING file path
	b, err := ioutil.ReadFile("../configs/credentials.json")
	// b, err := ioutil.ReadFile("configs/credentials.json")
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
	headersMap := getHeaders(resp.Values)
	// fmt.Println(headersMap)

	groceryMap := make(map[string][]string)

	ingredients := getIngredients(resp.Values, "1-M", headersMap)
	units := loadUnitConversions()

	addIngredientsToList(ingredients, units, groceryMap)
	fmt.Println(groceryMap)

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

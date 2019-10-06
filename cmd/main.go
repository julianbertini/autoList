// Copyright (c) 2019 Julian Bertini
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/julianbertini/autoList/internal/recipe"
	"github.com/julianbertini/autoList/internal/sheetswrapper"
)

// Define a flag type of []string to store the command-line recipe IDs
type recipeList []string

// Define the variable that the command-line arguments will be stored into
var rl recipeList

// String is the method to format the flag's value, part of the flag.Value interface.
// The String method's output will be used in diagnostics.
func (rl *recipeList) String() string {
	return fmt.Sprint(*rl)
}

// Set is the method to set the flag value, part of the flag.Value interface.
// Set's argument is a string to be parsed to set the flag.
// It's a comma-separated list, so we split it.
func (rl *recipeList) Set(value string) error {

	if len(*rl) > 0 {
		return errors.New("interval flag already set")
	}

	for _, id := range strings.Split(value, ",") {
		*rl = append(*rl, id)
	}
	return nil
}

// Define the flag as a custom Var flag. This function gets called before main
// implicitly by Go. It's used to initialize things, in general.
func init() {

	// Tie the command-line flag to the rl variable and
	// set a usage message.
	flag.Var(&rl, "rl", "comma-separated list of IDs for recipes")
}

func main() {

	// ################# BEGIN: Executes Once

	// Initialize Google Sheets connection
	spreadsheetID := "1sLGhnXgDLig0HTJE4_uQ6OXXc4iQO4U5QdqWGfyyBck"
	readRange := "Recipes!A:Z"
	service := sheetswrapper.GetService()
	resp, err := service.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()

	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	// Initialize unit converter, headersMap, and groceryMap
	units := recipe.LoadUnitConversions()
	headersMap := recipe.GetHeaders(resp.Values)
	groceryMap := make(map[string][]string)

	// Tells the flags package to parge command-line flags as defined above
	flag.Parse()

	// ################# END

	// Generates ingredient list from provided IDs

	for _, id := range rl {
		// Get ingredients for specified recipe ID
		ingredients := recipe.GetIngredients(resp.Values, id, headersMap)

		// Add ingredients to in-memory grocery list
		if len(ingredients) > 0 {
			recipe.AddIngredientsToList(ingredients, units, groceryMap)
		}
	}

	if len(groceryMap) > 0 {
		// Save in-memory grocery list to .txt file
		recipe.SaveListToFile("testGroceryList.txt", groceryMap)
	}

}

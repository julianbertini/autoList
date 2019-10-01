// Copyright (c) 2019 Julian Bertini
package main

import (
	"log"

	"github.com/julianbertini/autoList/internal/recipe"
	"github.com/julianbertini/autoList/internal/sheetswrapper"
)

func main() {

	spreadsheetID := "1sLGhnXgDLig0HTJE4_uQ6OXXc4iQO4U5QdqWGfyyBck"
	readRange := "Recipes!A:Z"
	service := sheetswrapper.GetService()
	resp, err := service.Spreadsheets.Values.Get(spreadsheetID, readRange).Do()

	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet: %v", err)
	}

	// need to identify indices of headers
	headersMap := recipe.GetHeaders(resp.Values)
	// fmt.Println(headersMap)

	groceryMap := make(map[string][]string)

	ingredients := recipe.GetIngredients(resp.Values, "1-M", headersMap)
	units := recipe.LoadUnitConversions()

	recipe.AddIngredientsToList(ingredients, units, groceryMap)

	recipe.SaveListToFile("groceryList.txt", groceryMap)

}

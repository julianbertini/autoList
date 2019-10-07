// Copyright (c) 2019 Julian Bertini

// TODO:
//		- What happens when ingredient does not specify a measurement (missing amount and unit)?
//		- Show selected recipe names on grocery list printout
//		- Make ingredient processing case-agnostic (so upper/lower case doesn't make ingredients different)
//		- When unit conversion cannot be made, add ingredient to list in diff. units

package recipe

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
)

//QuantitySep denotes separation between measurement quant. and type
const QuantitySep string = "/"

//MeasurementSep denotes the end of ingredient and start of measurement
const MeasurementSep string = ";"

//IngredientSep denotes the end of ingredient and start of measurement
const IngredientSep string = ","

// Ingredients name of ingredients column
const Ingredients string = "Ingredients"

// Recipe Name: name of Recipe Name column
const RecipeName string = "Recipe Name"

// IDprefix ID- which predeces every ID number
const IDprefix string = "ID-"

// IDsep is what separates id values
const IDsep string = "-"

// CategorySep delineates betweeen categories
const CategorySep string = "###"

type unitConversions struct {
	Tsp  map[string]float64 `json:"tsp"`
	Tbsp map[string]float64 `json:"tbsp"`
	Lb   map[string]float64 `json:"lb"`
	Oz   map[string]float64 `json:"oz"`
	Floz map[string]float64 `json:"floz"`
	Cups map[string]float64 `json:"cups"`
}

func writeHeader(f *os.File) {
	header := "\n<<<<< Grocery List >>>>>\n"
	header += "------------------------\n"
	f.WriteString(header)
}

func writeRecipeNames(f *os.File, recipeNames []string) {
	recipes := "\n<<<<< Recipe Names >>>>>\n"
	recipes += "------------------------\n"
	for _, name := range recipeNames {
		recipes += " - " + name + "\n"
	}
	f.WriteString(recipes)
}

func SaveListToFile(path string, groceryMap map[string][]string, recipeNames []string) {
	fmt.Printf("Saving grocery list to file: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalf("Unable to open or create file to save grocery list: %v", err)
	}

	defer f.Close()

	writeHeader(f)
	// Here we write contents of groceryMap
	for ingredient, measurement := range groceryMap {

		ingredientLine := " # " + ingredient + " --> " + "["
		for i, measurementToken := range measurement {
			if i > 0 && i%2 == 0 {
				ingredientLine += ", " + measurementToken
			} else {
				ingredientLine += " " + measurementToken
			}
		}
		ingredientLine += "]" + "\n"

		_, err = f.WriteString(ingredientLine)
		if err != nil {
			log.Fatalf("Error writing ingredients to .txt file: %v", err)
		}
	}
	writeRecipeNames(f, recipeNames)

	fmt.Printf("Done!\n")
}

func LoadUnitConversions() *unitConversions {
	// DEBUGGING
	// f, err := os.Open("../configs/measurements.json")
	f, err := os.Open("../configs/measurements.json")
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

func AddIngredientsToList(ingredients []string, units *unitConversions, groceryMap map[string][]string) {

	var fromUnit string
	var foundUnitIdx int = 0
	var toPossibleUnits []string
	var convertErr error

	// ["Onion;1/cups"]
	for _, ingredient := range ingredients {
		newIngredientTokens := strings.Split(ingredient, MeasurementSep) // ["Onion", "1/cups"]

		if len(newIngredientTokens) != 2 {
			log.Fatalf("\nUnexpected symbol encountered while parsing ingredient: %s.\n\t Expected ';' symbol to separate ingredient and measurement.", ingredient)
		}

		newIngredientName := strings.TrimSpace(newIngredientTokens[0])              // "Onion"
		newIngredientMeasurment := strings.TrimSpace(newIngredientTokens[1])        // "1/cups"
		newMeasurementTokens := strings.Split(newIngredientMeasurment, QuantitySep) // ["1", "cups"]
		fromQuant := strings.TrimSpace(newMeasurementTokens[0])                     // "1"
		fromQuantFloat, err := strconv.ParseFloat(fromQuant, 64)

		if err != nil {
			log.Fatalf("\nError:\n* Unexpected token encountered while parsing ingredient: %s.\n* Expected numeric quantity only, or amount and units separated by '/' symbol.\n* Message: %v", ingredient, err)
		}

		fromUnit = ""

		existingMeasurementTokens, ok := groceryMap[newIngredientName]

		if len(newMeasurementTokens) > 1 && len(newMeasurementTokens[1]) > 0 {
			fromUnit = strings.TrimSpace(newMeasurementTokens[1])
		}
		if len(existingMeasurementTokens) > 1 && len(existingMeasurementTokens[1]) > 0 {
			for i := range existingMeasurementTokens {
				if (i+1)%2 == 0 { // only even elements correspond to units, others are quantities
					toPossibleUnits = append(toPossibleUnits, existingMeasurementTokens[i]) // "[cups, lb, ...]"
				}
			}
		}

		if ok { // 1: Check if the ingredient is already in groveryMap. It is.

			// If there are from units and possible to units to convert to or add on to
			if (len(fromUnit) > 0 && len(toPossibleUnits) > 0) || (len(fromUnit) == 0 && len(toPossibleUnits) == 0) {

				if len(fromUnit) > 0 && len(toPossibleUnits) > 0 {
					foundUnit := false
					for i, unit := range toPossibleUnits {
						if fromUnit == unit {
							foundUnit = true
							foundUnitIdx = i
							break
						}
					}
					if !foundUnit {
						fromQuantFloat, convertErr = convertUnit(fromQuant, fromUnit, toPossibleUnits, units)
						if convertErr != nil {
							fmt.Printf("Warning: %v\n", convertErr)
							groceryMap[newIngredientName] = append(groceryMap[newIngredientName], fromQuant, fromUnit)
							return
						}
					}
				}
				// Convert existing ingredient amount to float
				toQuant, _ := strconv.ParseFloat(groceryMap[newIngredientName][foundUnitIdx], 64)
				// Add existing ingredient amount to new amount, and convert to str.
				newQuant := strconv.FormatInt(int64(math.Round(fromQuantFloat+toQuant)), 10)
				// Add new ingredient quantity to existing ingredient in list
				groceryMap[newIngredientName][foundUnitIdx] = newQuant

			} else {
				groceryMap[newIngredientName] = append(groceryMap[newIngredientName], fromQuant, fromUnit)
			}

		} else { // 2: Check if the ingredient is already in groceryMap. It is not.
			// Add the ingredient and amount to the map
			measurement := []string{fromQuant, fromUnit}
			groceryMap[newIngredientName] = measurement
		}
	}
}

func GetIngredients(respValues [][]interface{}, recipeID string, headersMap map[string]map[string][2]int) []string {

	var ingredients string = ""
	var category string
	var coords [2]int

	recipeNum, err := strconv.Atoi(strings.Split(recipeID, IDsep)[0])

	if err != nil {
		log.Fatalf("\nError: Recipe ID provided is not valid. \n\t ID must be in the form of [integer]-[letter].")
	}

	category = strings.Split(recipeID, IDsep)[1]

	coords = headersMap[category][Ingredients]
	coords[0] = coords[0] + recipeNum

	if len(respValues) <= coords[0] || len(respValues[coords[0]]) <= coords[1] {
		fmt.Printf("Warning: recipe with ID %s not found or missing ingredients.\n", recipeID)
		return strings.Split(ingredients, "")
	}

	ingredients = respValues[coords[0]][coords[1]].(string)

	return strings.Split(ingredients, IngredientSep)
}

func GetRecipeNames(respValues [][]interface{}, rl []string, headers map[string]map[string][2]int) []string {
	var recipeNames []string
	var category string
	var coords [2]int

	for _, id := range rl {

		recipeNum, err := strconv.Atoi(strings.Split(id, IDsep)[0])

		if err != nil {
			log.Fatalf("\nError: Recipe ID provided is not valid. \n\t ID must be in the form of [integer]-[letter].")
		}

		category = strings.Split(id, IDsep)[1]

		coords = headers[category][RecipeName]
		coords[0] = coords[0] + recipeNum

		if len(respValues) <= coords[0] || len(respValues[coords[0]]) <= coords[1] {
			fmt.Printf("Warning: recipe with ID %s not found or missing Recipe Name.\n", id)
		} else {
			recipeNames = append(recipeNames, respValues[coords[0]][coords[1]].(string))
		}

	}

	return recipeNames
}

func GetHeaders(respValues [][]interface{}) map[string]map[string][2]int {

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

func convertUnit(fromQuant string, fromUnit string, toUnits []string, units *unitConversions) (float64, error) {

	var newQuant float64 = 0
	var err error = nil
	floatQuant, _ := strconv.ParseFloat(fromQuant, 32)

	switch fromUnit {
	case "tsp":
		for _, toUnit := range toUnits {
			if _, ok := units.Tsp[toUnit]; ok {
				newQuant = floatQuant * units.Tsp[toUnit]
			}
		}
		if newQuant == 0 {
			err = errors.New("Unit conversion not found from " + fromUnit + "to" + strings.Join(toUnits, ",") + ".")
		}
	case "tbsp":
		for _, toUnit := range toUnits {
			if _, ok := units.Tbsp[toUnit]; ok {
				newQuant = floatQuant * units.Tbsp[toUnit]
			}
		}
		if newQuant == 0 {
			err = errors.New("Unit conversion not found from " + fromUnit + "to" + strings.Join(toUnits, ",") + ".")
		}
	case "lb":
		for _, toUnit := range toUnits {
			if _, ok := units.Lb[toUnit]; ok {
				newQuant = floatQuant * units.Lb[toUnit]
			}
		}
		if newQuant == 0 {
			err = errors.New("Unit conversion not found from " + fromUnit + "to" + strings.Join(toUnits, ",") + ".")
		}
	case "oz":
		for _, toUnit := range toUnits {
			if _, ok := units.Oz[toUnit]; ok {
				newQuant = floatQuant * units.Oz[toUnit]
			}
		}
		if newQuant == 0 {
			err = errors.New("Unit conversion not found from " + fromUnit + "to" + strings.Join(toUnits, ",") + ".")
		}
	case "floz":
		for _, toUnit := range toUnits {
			if _, ok := units.Floz[toUnit]; ok {
				newQuant = floatQuant * units.Floz[toUnit]
			}
		}
		if newQuant == 0 {
			err = errors.New("Unit conversion not found from " + fromUnit + "to" + strings.Join(toUnits, ",") + ".")
		}
	case "cups":
		for _, toUnit := range toUnits {
			if _, ok := units.Cups[toUnit]; ok {
				newQuant = floatQuant * units.Cups[toUnit]
			}
		}
		if newQuant == 0 {
			err = errors.New("Unit conversion not found from " + fromUnit + "to" + strings.Join(toUnits, ",") + ".")
		}
	default:
		err = errors.New("Unit conversion from " + fromUnit + " not supported.")
	}

	return newQuant, err
}

package helpers

import (
	"math/big"

	"go.vocdoni.io/dvote/api"
)

// ExtractResults extracts the choices and results from an election. It returns nil if there is an issue processing the data.
func ExtractResults(election *api.Election, censusTokenDecimals uint32) (choices []string, results []*big.Int) {
	if election == nil || election.Metadata == nil || election.Results == nil {
		return nil, nil // Return nil if the main structures are nil
	}

	apiQuestions := election.Metadata.Questions
	apiResults := election.Results
	if len(apiQuestions) == 0 || len(apiQuestions[0].Choices) == 0 ||
		len(apiResults) == 0 || len(apiResults[0]) < len(apiQuestions[0].Choices) {
		return nil, nil
	}

	for _, question := range apiQuestions[0].Choices {
		t, ok := question.Title["default"]
		if !ok {
			continue // Skip if there's no default title
		}
		// check for the index in the results array
		if len(apiResults[0]) <= int(question.Value) {
			continue
		}
		bigIntResult := apiResults[0][question.Value].MathBigInt()
		if censusTokenDecimals > 0 {
			// Scale the result down based on the number of decimals
			scalingFactor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(censusTokenDecimals)), nil)
			bigIntResult = new(big.Int).Div(bigIntResult, scalingFactor)
		}
		choices = append(choices, t)
		results = append(results, bigIntResult)
	}
	return choices, results
}

// CalculateTurnout computes the turnout percentage from two big.Int strings.
// If the strings are not valid numbers, it returns zero.
func CalculateTurnout(totalWeightStr, castedWeightStr string) *big.Int {
	totalWeight := new(big.Int)
	castedWeight := new(big.Int)

	_, ok := totalWeight.SetString(totalWeightStr, 10)
	if !ok {
		return big.NewInt(0)
	}

	_, ok = castedWeight.SetString(castedWeightStr, 10)
	if !ok {
		return big.NewInt(0)
	}

	// Multiply castedWeight by 100 to preserve integer properties during division
	castedWeightMul := new(big.Int).Mul(castedWeight, big.NewInt(100))

	// Compute the turnout percentage as an integer if the total weight is not zero
	if totalWeight.Cmp(big.NewInt(0)) == 0 {
		return big.NewInt(0)
	}
	turnoutPercentage := new(big.Int).Div(castedWeightMul, totalWeight)

	return turnoutPercentage
}

// bigIntsToStrings converts a slice of *big.Int to a slice of their string representations.
// It safely handles nil pointers within the input slice.
func BigIntsToStrings(bigInts []*big.Int) []string {
	strings := make([]string, len(bigInts))
	for i, bigInt := range bigInts {
		if bigInt == nil {
			strings[i] = "nil" // Represent nil pointers as "nil" in the output
		} else {
			strings[i] = bigInt.String()
		}
	}
	return strings
}

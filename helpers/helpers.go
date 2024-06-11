package helpers

import (
	"encoding/json"
	"math/big"

	"go.vocdoni.io/dvote/api"
	"go.vocdoni.io/dvote/log"
)

// ExtractResults extracts the choices and results from an election. It returns nil if there is an issue processing the data.
func ExtractResults(election *api.Election, censusTokenDecimals uint32) (choices []string, results []*big.Int) {
	if election == nil || election.Metadata == nil || election.Results == nil {
		return nil, nil // Return nil if the main structures are nil
	}
	metadata := UnpackMetadata(election.Metadata)

	apiQuestions := metadata.Questions
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
			bigIntResult = TruncateDecimals(bigIntResult, censusTokenDecimals)
		}
		choices = append(choices, t)
		results = append(results, bigIntResult)
	}
	return choices, results
}

// CalculateTurnout computes the turnout percentage from two big.Int strings.
// If the strings are not valid numbers, it returns zero.
func CalculateTurnout(totalWeightStr, castedWeightStr string) float32 {
	totalWeight := new(big.Int)
	castedWeight := new(big.Int)

	_, ok := totalWeight.SetString(totalWeightStr, 10)
	if !ok {
		return 0
	}

	_, ok = castedWeight.SetString(castedWeightStr, 10)
	if !ok {
		return 0
	}

	// Multiply castedWeight by 100 to preserve integer properties during division
	castedWeightFloat, _ := new(big.Int).Mul(castedWeight, big.NewInt(100)).Float64()

	// Compute the turnout percentage as an integer if the total weight is not zero
	if totalWeight.Cmp(big.NewInt(0)) == 0 {
		return 0
	}
	totalWeightFloat, _ := totalWeight.Float64()
	turnoutPercentage := castedWeightFloat / totalWeightFloat

	return float32(turnoutPercentage)
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

// TruncateDecimals takes a big.Int representing a fixed-point number and truncates it
// to a whole number by removing the specified number of decimal places.
func TruncateDecimals(num *big.Int, numberOfDecimals uint32) *big.Int {
	if numberOfDecimals == 0 || num == nil {
		return num
	}
	// Create a big.Int from 10
	ten := big.NewInt(10)

	// Calculate 10^numberOfDecimals
	divisor := new(big.Int).Exp(ten, big.NewInt(int64(numberOfDecimals)), nil)

	// Divide num by 10^numberOfDecimals to truncate the decimal part
	result := new(big.Int).Div(num, divisor)

	return result
}

// UnpackMetadata is a helper function to unpack the metadata of an election
// to the api.ElectionDescription type.
func UnpackMetadata(metadata any) *api.ElectionDescription {
	emptyDesc := api.ElectionDescription{
		Title: map[string]string{"default": "unknown"},
		Questions: []api.Question{{
			Choices: []api.ChoiceMetadata{},
		}},
	}
	// Note that we need to marshal and then unmarsla because the metadata object can be of any type.
	// We expect to find the ElectionDescription object inside the metadata, but it could be anything else.
	// So the safer approach is to marshal and unmarshal it.
	data, err := json.Marshal(metadata)
	if err != nil {
		log.Warnw("failed to marshal metadata", "error", err)
		return &emptyDesc
	}
	desc := &api.ElectionDescription{}
	if err := json.Unmarshal(data, desc); err != nil {
		log.Warnw("failed to unmarshal metadata", "error", err)
		return &emptyDesc
	}
	return desc
}

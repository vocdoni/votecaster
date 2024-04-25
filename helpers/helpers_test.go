package helpers

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.vocdoni.io/dvote/api"
	"go.vocdoni.io/dvote/types"
)

func TestExtractResults(t *testing.T) {
	testCases := []struct {
		name            string
		election        *api.Election
		tokenDecimals   uint32
		expectedChoices []string
		expectedResults []*big.Int
	}{
		{
			name: "Correct results and choices with four items",
			election: &api.Election{
				Metadata: &api.ElectionMetadata{
					Questions: []api.Question{
						{
							Choices: []api.ChoiceMetadata{
								{Title: map[string]string{"default": "Choice 1"}, Value: 0},
								{Title: map[string]string{"default": "Choice 2"}, Value: 1},
								{Title: map[string]string{"default": "Choice 3"}, Value: 2},
								{Title: map[string]string{"default": "Choice 4"}, Value: 3},
							},
						},
					},
				},
				ElectionSummary: api.ElectionSummary{
					Results: [][]*types.BigInt{
						{
							new(types.BigInt).SetUint64(100),
							new(types.BigInt).SetUint64(200),
							new(types.BigInt).SetUint64(300),
							new(types.BigInt).SetUint64(400),
						},
					},
				},
			},
			tokenDecimals:   0,
			expectedChoices: []string{"Choice 1", "Choice 2", "Choice 3", "Choice 4"},
			expectedResults: []*big.Int{big.NewInt(100), big.NewInt(200), big.NewInt(300), big.NewInt(400)},
		},
		{
			name: "Results and choices with different sizes",
			election: &api.Election{
				Metadata: &api.ElectionMetadata{
					Questions: []api.Question{
						{
							Choices: []api.ChoiceMetadata{
								{Title: map[string]string{"default": "Choice 1"}, Value: 0},
								{Title: map[string]string{"default": "Choice 2"}, Value: 1},
							},
						},
					},
				},
				ElectionSummary: api.ElectionSummary{
					Results: [][]*types.BigInt{
						{
							new(types.BigInt).SetUint64(100),
						},
					},
				},
			},
			tokenDecimals:   0,
			expectedChoices: nil,
			expectedResults: nil,
		},
		{
			name: "Choices with unordered values",
			election: &api.Election{
				Metadata: &api.ElectionMetadata{
					Questions: []api.Question{
						{
							Choices: []api.ChoiceMetadata{
								{Title: map[string]string{"default": "Choice 3"}, Value: 2},
								{Title: map[string]string{"default": "Choice 1"}, Value: 0},
								{Title: map[string]string{"default": "Choice 2"}, Value: 1},
							},
						},
					},
				},
				ElectionSummary: api.ElectionSummary{
					Results: [][]*types.BigInt{
						{
							new(types.BigInt).SetUint64(100),
							new(types.BigInt).SetUint64(200),
							new(types.BigInt).SetUint64(300),
						},
					},
				},
			},

			tokenDecimals:   0,
			expectedChoices: []string{"Choice 3", "Choice 1", "Choice 2"},
			expectedResults: []*big.Int{big.NewInt(300), big.NewInt(100), big.NewInt(200)},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			choices, results := ExtractResults(tc.election, tc.tokenDecimals)
			assert.Equal(t, tc.expectedChoices, choices)
			assert.Equal(t, tc.expectedResults, results)
		})
	}
}

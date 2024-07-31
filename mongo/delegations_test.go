package mongo

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func Test_solveNestedDelegations(t *testing.T) {
	var (
		one   = primitive.NewObjectID()
		two   = primitive.NewObjectID()
		three = primitive.NewObjectID()
	)
	setA := []Delegation{
		{
			ID:         one,
			From:       1,
			To:         2,
			CommuniyID: "a",
		},
		{
			ID:         two,
			From:       2,
			To:         3,
			CommuniyID: "a",
		},
		{
			ID:         three,
			From:       3,
			To:         4,
			CommuniyID: "a",
		},
	}
	expectedA := []Delegation{
		{
			ID:         one,
			From:       1,
			To:         4,
			CommuniyID: "a",
		},
		{
			ID:         two,
			From:       2,
			To:         4,
			CommuniyID: "a",
		},
		{
			ID:         three,
			From:       3,
			To:         4,
			CommuniyID: "a",
		},
	}

	resultsA := solveNestedDelegations(setA, nil)
	if len(resultsA) != len(expectedA) {
		t.Errorf("expected len %d, got %d", len(expectedA), len(resultsA))
	}

	for _, expected := range expectedA {
		found := false
		for _, result := range resultsA {
			if expected.From == result.From &&
				expected.To == result.To &&
				expected.CommuniyID == result.CommuniyID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %v to be in result", expected)
		}
	}

	filterB := []Delegation{
		{
			ID:         one,
			From:       1,
			To:         2,
			CommuniyID: "a",
		},
	}

	expectedB := []Delegation{
		{
			ID:         one,
			From:       1,
			To:         4,
			CommuniyID: "a",
		},
	}

	resultsB := solveNestedDelegations(setA, filterB)
	if len(resultsB) != len(expectedB) {
		t.Errorf("expected len %d, got %d", len(expectedB), len(resultsB))
	}

	for _, expected := range expectedB {
		found := false
		for _, result := range resultsB {
			if expected.From == result.From &&
				expected.To == result.To &&
				expected.CommuniyID == result.CommuniyID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected %v to be in result", expected)
		}
	}
}

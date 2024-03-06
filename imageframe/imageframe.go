package imageframe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"go.vocdoni.io/dvote/api"
)

const (
	BackgroundAfterVote    = "aftervote.png"
	BackgroundAlreadyVoted = "alreadyvoted.png"
	BackgroundNotElegible  = "notelegible.png"

	BackgroundsDir = "images/"
	BaseURL        = "https://img.frame.vote"
)

var backgroundFrames map[string][]byte

func init() {
	loadImage := func(name string) []byte {
		imgFile, err := os.Open(path.Join(BackgroundsDir, name))
		if err != nil {
			log.Fatalf("failed to load image %s: %v", name, err)
		}
		defer imgFile.Close()
		b, err := io.ReadAll(imgFile)
		if err != nil {
			log.Fatalf("failed to read image %s: %v", name, err)
		}
		return b
	}
	backgroundFrames = make(map[string][]byte)
	backgroundFrames[BackgroundAfterVote] = loadImage(BackgroundAfterVote)
	backgroundFrames[BackgroundAlreadyVoted] = loadImage(BackgroundAlreadyVoted)
	backgroundFrames[BackgroundNotElegible] = loadImage(BackgroundNotElegible)
}

// ImageRequest is a general struct for making requests to the API.
// It includes all possible fields that can be sent to the API.
type ImageRequest struct {
	Type          string   `json:"type"`
	Error         string   `json:"error,omitempty"`
	Info          []string `json:"info,omitempty"`
	Question      string   `json:"question,omitempty"`
	Choices       []string `json:"choices,omitempty"`
	Results       []string `json:"results,omitempty"`
	VoteCount     int      `json:"voteCount,omitempty"`
	MaxCensusSize int      `json:"maxCensusSize,omitempty"`
}

// ErrorImage creates an image representing an error message.
func ErrorImage(errorMessage string) ([]byte, error) {
	requestData := ImageRequest{
		Type:  "error",
		Error: errorMessage,
	}
	return makeRequest(requestData)
}

// InfoImage creates an image displaying an informational message.
func InfoImage(infoLines []string) ([]byte, error) {
	requestData := ImageRequest{
		Type: "info",
		Info: infoLines,
	}
	return makeRequest(requestData)
}

// QuestionImage creates an image representing a question with choices.
func QuestionImage(election *api.Election) ([]byte, error) {
	title := election.Metadata.Questions[0].Title["default"]
	var choices []string
	for _, option := range election.Metadata.Questions[0].Choices {
		choices = append(choices, option.Title["default"])
	}

	requestData := ImageRequest{
		Type:     "question",
		Question: title,
		Choices:  choices,
	}
	return makeRequest(requestData)
}

// ResultsImage creates an image showing the results of a poll.
func ResultsImage(election *api.Election) ([]byte, error) {
	if election == nil || election.Metadata == nil || len(election.Metadata.Questions) == 0 {
		return nil, fmt.Errorf("election has no questions")
	}

	title := election.Metadata.Questions[0].Title["default"]
	choices := []string{}
	results := []string{}
	for _, option := range election.Metadata.Questions[0].Choices {
		choices = append(choices, option.Title["default"])
		results = append(results, election.Results[0][option.Value].MathBigInt().String())
	}

	requestData := ImageRequest{
		Type:          "results",
		Question:      title,
		Choices:       choices,
		Results:       results,
		VoteCount:     int(election.VoteCount),
		MaxCensusSize: int(election.Census.MaxCensusSize),
	}

	return makeRequest(requestData)
}

// AfterVoteImage creates a static image to be displayed after a vote has been cast.
func AfterVoteImage() []byte {
	return backgroundFrames[BackgroundAfterVote]
}

// AlreadyVotedImage creates a static image to be displayed when a user has already voted.
func AlreadyVotedImage() []byte {
	return backgroundFrames[BackgroundAlreadyVoted]
}

// NotElegibleImage creates a static image to be displayed when a user is not elegible to vote.
func NotElegibleImage() []byte {
	return backgroundFrames[BackgroundNotElegible]
}

// makeRequest handles the communication with the API.
func makeRequest(data ImageRequest) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	response, err := http.Post(fmt.Sprintf("%s/image", BaseURL), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", response.StatusCode)
	}

	return io.ReadAll(response.Body)
}

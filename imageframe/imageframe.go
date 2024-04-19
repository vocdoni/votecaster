package imageframe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"os"
	"path"
	"sync/atomic"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"go.vocdoni.io/dvote/api"
	"go.vocdoni.io/dvote/log"
)

const (
	BackgroundAfterVote    = "aftervote.png"
	BackgroundAlreadyVoted = "alreadyvoted.png"
	BackgroundNotElegible  = "notelegible.png"
	BackgroundNotFound     = "notfound.png"

	BackgroundNotificationsAccepted = "notifications-accepted.png"
	BackgroundNotificationsDenied   = "notifications-denied.png"
	BackgroundNotifications         = "notifications.png"
	BackgroundNotificationsError    = "notifications-error.png"

	BackgroundsDir    = "images/"
	ImageGeneratorURL = "https://img.frame.vote"

	TimeoutImageGeneration = 15 * time.Second
)

const (
	imageType = iota
	imageTypeQuestion
	imageTypeResults
)

var (
	backgroundFrames           map[string][]byte
	imagesLRU                  *lru.Cache[string, []byte]
	hitsCounter, missesCounter atomic.Int64
)

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
	backgroundFrames[BackgroundNotFound] = loadImage(BackgroundNotFound)
	backgroundFrames[BackgroundNotificationsAccepted] = loadImage(BackgroundNotificationsAccepted)
	backgroundFrames[BackgroundNotificationsDenied] = loadImage(BackgroundNotificationsDenied)
	backgroundFrames[BackgroundNotifications] = loadImage(BackgroundNotifications)
	backgroundFrames[BackgroundNotificationsError] = loadImage(BackgroundNotificationsError)

	var err error
	imagesLRU, err = lru.New[string, []byte](2048)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for range time.Tick(60 * time.Second) {
			log.Infow("image cache stats", "hits", hitsCounter.Load(), "misses", missesCounter.Load(), "size", imagesLRU.Len())
		}
	}()
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
func ErrorImage(errorMessage string) (string, error) {
	requestData := ImageRequest{
		Type:  "error",
		Error: errorMessage,
	}
	imgCacheKey := oneTimeImageCacheKey()
	go func() {
		png, err := makeRequest(requestData)
		if err != nil {
			log.Errorw(fmt.Errorf("failed to create image: %w", err), "error image")
			return
		}
		AddImageToCacheWithID(imgCacheKey, png)
	}()
	time.Sleep(2 * time.Second)
	return imgCacheKey, nil
}

// InfoImage creates an image displaying an informational message.
// Returns the image id that can be fetch using FromCache(id).
func InfoImage(infoLines []string) (string, error) {
	requestData := ImageRequest{
		Type: "info",
		Info: infoLines,
	}
	imgCacheKey := oneTimeImageCacheKey()
	go func() {
		png, err := makeRequest(requestData)
		if err != nil {
			log.Errorw(fmt.Errorf("failed to create image: %w", err), "info image")
			return
		}
		AddImageToCacheWithID(imgCacheKey, png)
	}()
	time.Sleep(2 * time.Second)
	return imgCacheKey, nil
}

// QuestionImage creates an image representing a question with choices.
func QuestionImage(election *api.Election) (string, error) {
	if election == nil || election.Metadata == nil || len(election.Metadata.Questions) == 0 {
		return "", fmt.Errorf("election has no questions")
	}
	// Check if the image is already in the cache
	if id := electionImageCacheKey(election, imageTypeQuestion); id != "" {
		return id, nil
	}

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
	go func() {
		png, err := makeRequest(requestData)
		if err != nil {
			log.Warnw("failed to create image", "error", err)
			return
		}
		cacheElectionImage(png, election, imageTypeQuestion)
	}()
	// Add some time to allow the image to be generated
	time.Sleep(2 * time.Second)
	return generateElectionCacheKey(election, imageTypeQuestion), nil
}

// ResultsImage creates an image showing the results of a poll.
func ResultsImage(election *api.Election, censusTokenDecimals uint32) (string, error) {
	if election == nil || election.Metadata == nil || len(election.Metadata.Questions) == 0 {
		return "", fmt.Errorf("election has no questions")
	}
	// Check if the image is already in the cache
	if id := electionImageCacheKey(election, imageTypeResults); id != "" {
		return id, nil
	}

	title := election.Metadata.Questions[0].Title["default"]

	choices := []string{}
	results := []string{}
	for _, option := range election.Metadata.Questions[0].Choices {
		choices = append(choices, option.Title["default"])
		value := ""
		if censusTokenDecimals > 0 {
			resultsValueFloat := new(big.Float).Quo(
				new(big.Float).SetInt(election.Results[0][option.Value].MathBigInt()),
				new(big.Float).SetInt(big.NewInt(int64(math.Pow(10, float64(censusTokenDecimals))))),
			)
			value = fmt.Sprintf("%.2f", resultsValueFloat)
		} else {
			value = election.Results[0][option.Value].MathBigInt().String()
		}
		results = append(results, value)
	}

	requestData := ImageRequest{
		Type:          "results",
		Question:      title,
		Choices:       choices,
		Results:       results,
		VoteCount:     int(election.VoteCount),
		MaxCensusSize: int(election.Census.MaxCensusSize),
	}

	go func() {
		png, err := makeRequest(requestData)
		if err != nil {
			log.Warnw("failed to create image", "error", err)
			return
		}
		cacheElectionImage(png, election, imageTypeResults)
	}()
	time.Sleep(2 * time.Second)
	return generateElectionCacheKey(election, imageTypeResults), nil
}

// AfterVoteImage creates a static image to be displayed after a vote has been cast.
func AfterVoteImage() string {
	return AddImageToCache(backgroundFrames[BackgroundAfterVote])
}

// AlreadyVotedImage creates a static image to be displayed when a user has already voted.
func AlreadyVotedImage() string {
	return AddImageToCache(backgroundFrames[BackgroundAlreadyVoted])
}

// NotElegibleImage creates a static image to be displayed when a user is not elegible to vote.
func NotElegibleImage() string {
	return AddImageToCache(backgroundFrames[BackgroundNotElegible])
}

// NotFoundImage creates a static image to be displayed when an election is not found.
func NotFoundImage() string {
	return AddImageToCache(backgroundFrames[BackgroundNotFound])
}

// NotificationsAcceptedImage creates a static image to be displayed when notifications are accepted.
func NotificationsAcceptedImage() string {
	return AddImageToCache(backgroundFrames[BackgroundNotificationsAccepted])
}

// NotificationsDeniedImage creates a static image to be displayed when notifications are denied.
func NotificationsDeniedImage() string {
	return AddImageToCache(backgroundFrames[BackgroundNotificationsDenied])
}

// NotificationsImage creates a static image to be displayed when notifications are requested.
func NotificationsImage() string {
	return AddImageToCache(backgroundFrames[BackgroundNotifications])
}

// NotificationsErrorImage creates a static image to be displayed when there is an error with notifications.
func NotificationsErrorImage() string {
	return AddImageToCache(backgroundFrames[BackgroundNotificationsError])
}

// makeRequest handles the communication with the API.
func makeRequest(data ImageRequest) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	startTime := time.Now()
	response, err := http.Post(fmt.Sprintf("%s/image", ImageGeneratorURL), "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	log.Debugw("image request", "type", data.Type, "elapsed (s)", time.Since(startTime).Seconds())
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", response.StatusCode)
	}

	return io.ReadAll(response.Body)
}

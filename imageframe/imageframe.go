package imageframe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"sync/atomic"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/vocdoni/vote-frame/helpers"
	"github.com/vocdoni/vote-frame/mongo"
	"go.vocdoni.io/dvote/api"
	"go.vocdoni.io/dvote/log"
)

const (
	BackgroundNotificationsAccepted = "notifications-accepted.png"
	BackgroundNotificationsDenied   = "notifications-denied.png"
	BackgroundNotifications         = "notifications.png"
	BackgroundNotificationsError    = "notifications-error.png"
	BackgroundNotificationsManage   = "notifications-manage.png"

	BackgroundsDir    = "images/"
	ImageGeneratorURL = "https://img.frame.vote"

	TimeoutImageGeneration = 15 * time.Second
)

const (
	imageType = iota
	imageTypeQuestion
	imageTypeResults
	imageTypePreview
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
	backgroundFrames[BackgroundNotificationsAccepted] = loadImage(BackgroundNotificationsAccepted)
	backgroundFrames[BackgroundNotificationsDenied] = loadImage(BackgroundNotificationsDenied)
	backgroundFrames[BackgroundNotifications] = loadImage(BackgroundNotifications)
	backgroundFrames[BackgroundNotificationsError] = loadImage(BackgroundNotificationsError)
	backgroundFrames[BackgroundNotificationsManage] = loadImage(BackgroundNotificationsManage)

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
	Title         string   `json:"title,omitempty"`
	Type          string   `json:"type"`
	Error         string   `json:"error,omitempty"`
	Info          []string `json:"info,omitempty"`
	Question      string   `json:"question,omitempty"`
	Choices       []string `json:"choices,omitempty"`
	Results       []string `json:"results,omitempty"`
	VoteCount     uint64   `json:"voteCount"`
	Participation float32  `json:"participation"`
	Turnout       float32  `json:"turnout"`
	Ends          string   `json:"ends,omitempty"`
	Ended         bool     `json:"ended,omitempty"`
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
	time.Sleep(1 * time.Second)
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
	time.Sleep(1 * time.Second)
	return imgCacheKey, nil
}

// QuestionImage creates an image representing a question with choices.
func QuestionImage(election *api.Election) (string, error) {
	if election == nil || election.Metadata == nil {
		return "", fmt.Errorf("election has no metadata")
	}
	metadata := helpers.UnpackMetadata(election.Metadata)
	// Check if the image is already in the cache
	if id := electionImageCacheKey(election, imageTypeQuestion); id != "" {
		return id, nil
	}

	title := metadata.Questions[0].Title["default"]
	var choices []string
	for _, option := range metadata.Questions[0].Choices {
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
	time.Sleep(1 * time.Second)
	return generateElectionCacheKey(election, imageTypeQuestion), nil
}

// Preview creates an image representing a question preview.
func Preview(election *api.Election) (string, error) {
	if election == nil || election.Metadata == nil {
		return "", fmt.Errorf("election has no metadata")
	}
	metadata := helpers.UnpackMetadata(election.Metadata)
	// Check if the image is already in the cache
	if id := electionImageCacheKey(election, imageTypePreview); id != "" {
		return id, nil
	}

	title := metadata.Questions[0].Title["default"]
	ends := election.EndDate.Format("2006-01-02 15:04:05")
	ended := election.EndDate.Before(time.Now())

	requestData := ImageRequest{
		Type:     "preview",
		Ends:     ends,
		Question: title,
		Ended:    ended,
	}

	go func() {
		png, err := makeRequest(requestData)
		if err != nil {
			log.Warnw("failed to create image", "error", err)
			return
		}
		cacheElectionImage(png, election, imageTypePreview)
	}()
	// Add some time to allow the image to be generated
	time.Sleep(2 * time.Second)
	return generateElectionCacheKey(election, imageTypePreview), nil
}

// ResultsImage creates an image showing the results of a poll.
// It returns the image id that can be fetch using FromCache(id).
// The totalWeightStr is the total weight of the census, if empty Turnout is not calculated.
// The electiondb is the election data from the database, if nil the participation is not calculated.
func ResultsImage(election *api.Election, electiondb *mongo.Election, totalWeightStr string) (string, error) {
	if election == nil || election.Metadata == nil {
		return "", fmt.Errorf("election has no metadata")
	}
	metadata := helpers.UnpackMetadata(election.Metadata)
	// Check if the image is already in the cache
	if id := electionImageCacheKey(election, imageTypeResults); id != "" {
		return id, nil
	}

	participation := float32(0)
	weightTurnout := float32(0)

	if electiondb != nil {
		if electiondb.FarcasterUserCount > 0 {
			participation = (float32(electiondb.CastedVotes) * 100) / float32(electiondb.FarcasterUserCount)
		}
		weightTurnout = helpers.CalculateTurnout(totalWeightStr, electiondb.CastedWeight)
	}

	title := metadata.Questions[0].Title["default"]
	choices, results := helpers.ExtractResults(election, 0)

	requestData := ImageRequest{
		Type:          "results",
		Question:      title,
		Choices:       choices,
		Results:       helpers.BigIntsToStrings(results),
		VoteCount:     election.VoteCount,
		Participation: participation,
		Turnout:       weightTurnout,
	}
	log.Debugw("requesting results image",
		"type", requestData.Type,
		"question", requestData.Question,
		"choices", requestData.Choices,
		"results", requestData.Results,
		"voteCount", requestData.VoteCount,
		"participation", requestData.Participation,
		"turnout", requestData.Turnout)

	go func() {
		png, err := makeRequest(requestData)
		if err != nil {
			log.Warnw("failed to create image", "error", err)
			return
		}
		cacheElectionImage(png, election, imageTypeResults)
	}()
	time.Sleep(1 * time.Second)
	return generateElectionCacheKey(election, imageTypeResults), nil
}

// AfterVoteImage creates a static image to be displayed after a vote has been cast.
func AfterVoteImage() string {
	return emptyBodyImage("votecast")
}

// AlreadyVotedImage creates a static image to be displayed when a user has already voted.
func AlreadyVotedImage() string {
	return emptyBodyImage("alreadyvoted")
}

// NotElegibleImage creates a static image to be displayed when a user is not elegible to vote.
func NotElegibleImage() string {
	return emptyBodyImage("noteligible")
}

// AlreadyDelegated creates a static image to be displayed when a user has already delegated their vote.
func AlreadyDelegated() string {
	return notEligibleImage("You cannot vote as your voting power was delegated for this community poll")
}

// NotFoundImage creates a static image to be displayed when an election is not found.
func NotFoundImage() string {
	png, _ := ErrorImage("Election not found")
	return png
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

// NotificationsManageImage creates a static image to be displayed when the user tries to manage muted users.
func NotificationsManageImage() string {
	return AddImageToCache(backgroundFrames[BackgroundNotificationsManage])
}

// NotificationsErrorImage creates a static image to be displayed when there is an error with notifications.
func NotificationsErrorImage() string {
	return AddImageToCache(backgroundFrames[BackgroundNotificationsError])
}

// notEligibleImage creates images of methods that do not have body
func notEligibleImage(title string) string {
	requestData := ImageRequest{
		Type:  "noteligible",
		Title: title,
	}
	imgCacheKey := oneTimeImageCacheKey()
	go func() {
		png, err := makeRequest(requestData)
		if err != nil {
			log.Errorw(fmt.Errorf("failed to create noteligible image: %w", err), "error image")
			return
		}
		AddImageToCacheWithID(imgCacheKey, png)
	}()
	time.Sleep(2 * time.Second)
	return imgCacheKey
}

// emptyBodyImage creates images of methods that do not have body
func emptyBodyImage(t string) string {
	requestData := ImageRequest{
		Type: t,
	}
	imgCacheKey := oneTimeImageCacheKey()
	go func() {
		png, err := makeRequest(requestData)
		if err != nil {
			log.Errorw(fmt.Errorf("failed to create "+t+" image: %w", err), "error image")
			return
		}
		AddImageToCacheWithID(imgCacheKey, png)
	}()
	time.Sleep(2 * time.Second)
	return imgCacheKey
}

// makeRequest handles the communication with the API, with retries on failure.
func makeRequest(data ImageRequest) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	maxAttempts := 5
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		response, err := http.Post(fmt.Sprintf("%s/image", ImageGeneratorURL), "application/json", bytes.NewBuffer(jsonData))
		if err == nil && response.StatusCode == http.StatusOK {
			defer response.Body.Close()
			return io.ReadAll(response.Body)
		}

		if response != nil {
			response.Body.Close() // Ensure the response body is closed on each attempt.
		}

		if attempt < maxAttempts {
			sleepDuration := time.Duration(attempt*2) * time.Second // Exponential back-off strategy
			time.Sleep(sleepDuration)
			log.Debugw("retrying image request", "attempt", attempt, "sleepDuration", sleepDuration)
		} else {
			log.Debugw("image request failed after retries", "type", data.Type, "attempts", maxAttempts)
			break
		}
	}

	return nil, fmt.Errorf("image generation API request failed")
}

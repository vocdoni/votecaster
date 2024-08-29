package imageframe

import (
	"fmt"
	"time"

	"github.com/zeebo/blake3"
	"go.vocdoni.io/dvote/api"
	"go.vocdoni.io/dvote/log"
	"go.vocdoni.io/dvote/util"
)

// generateElectionCacheKey returns a unique identifier cache key, for the election.
// The cache key is based on the electionID, voteCount and finalResults.
func generateElectionCacheKey(election *api.Election, imageType int) string {
	if election == nil {
		return ""
	}
	switch imageType {
	case imageTypeResults:
		return fmt.Sprintf("%s_%d-%d%d", election.ElectionID.String(), election.VoteCount, func() int {
			if election.FinalResults {
				return 1
			}
			return 0
		}(), imageType)
	case imageTypePreview:
		// set current date and hour to cache file name
		dhour := time.Now().Format("2006-01-02_15")
		return fmt.Sprintf("%s_%s_%d", election.ElectionID.String(), dhour, imageType)
	case imageTypeQuestion:
		return fmt.Sprintf("%s_%d", election.ElectionID.String(), imageType)
	default:
		log.Errorw(fmt.Errorf("unknown image type %d", imageType), "cacheElectionID")
		// fallback
		return fmt.Sprintf("%s_%d", election.ElectionID.String(), imageType)
	}
}

// cacheElectionImage adds an image to the LRU cache.
// Returns the cache key.
// If electionID is nil, the image is not associated with any election.
func cacheElectionImage(data []byte, election *api.Election, imageType int) string {
	id := generateElectionCacheKey(election, imageType)
	imagesLRU.Add(id, data)
	return id
}

// electionImageCacheKey checks if an election associated image exist in the LRU cache.
// If so it returns the cache key identifier, otherwise it returns an empty string.
func electionImageCacheKey(election *api.Election, imageType int) string {
	id := generateElectionCacheKey(election, imageType)
	_, ok := imagesLRU.Get(id)
	if !ok {
		missesCounter.Add(1)
		return ""
	}
	hitsCounter.Add(1)
	return id
}

// genericImageCacheKey returns a unique identifier cache key, for the image data.
func oneTimeImageCacheKey() string {
	return util.RandomHex(20)
}

// AddImageToCache adds an image to the LRU cache.
// Returns the cache key.
// The key is the hash of the image data.
func AddImageToCache(data []byte) string {
	id := fmt.Sprintf("%x", blake3.Sum256(data))
	imagesLRU.Add(id, data)
	return id
}

// AddImageToCacheWithID adds an image to the LRU cache with a specific ID.
func AddImageToCacheWithID(id string, data []byte) {
	imagesLRU.Add(id, data)
}

// FromCache tries to retrieve data from cache with a specified id.
// Returns nil if the image is not in the cache.
// If the data is not available, it retries until the timeout is reached.
func FromCache(id string) []byte {
	if id == "" {
		return nil
	}

	// Using a ticker for retry interval
	ticker := time.NewTicker(time.Millisecond * 200)
	defer ticker.Stop()

	// Setting up a timeout
	timeoutChan := time.After(TimeoutImageGeneration)

	for {
		select {
		case <-timeoutChan:
			// Timeout reached, return nil
			return nil
		case <-ticker.C:
			// Attempt to get data from cache
			if data, ok := imagesLRU.Get(id); ok {
				return data
			}
		}
	}
}

// IsInCache checks if an image is in the LRU cache.
func IsInCache(id string) bool {
	return imagesLRU.Contains(id)
}

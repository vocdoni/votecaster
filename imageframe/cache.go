package imageframe

import (
	"fmt"
	"time"

	"github.com/zeebo/blake3"
	"go.vocdoni.io/dvote/api"
)

// cacheElectionID returns a unique identifier cache key, for the election.
// The cache key is based on the electionID, voteCount and finalResults.
func cacheElectionID(election *api.Election) string {
	if election == nil {
		return ""
	}
	return fmt.Sprintf("%s_%d-%d", election.ElectionID.String(), election.VoteCount, func() int {
		if election.FinalResults {
			return 1
		}
		return 0
	}())
}

// addElectionImageToCache adds an image to the LRU cache.
// Returns the cache key.
// If electionID is nil, the image is not associated with any election.
func addElectionImageToCache(data []byte, election *api.Election) string {
	id := cacheElectionID(election)
	imagesLRU.Add(id, data)
	return id
}

// electionImageInCache checks if an election associated image exist in the LRU cache.
// If so it returns the cache key identifier, otherwise it returns an empty string.
func electionImageInCache(election *api.Election) string {
	id := cacheElectionID(election)
	_, ok := imagesLRU.Get(id)
	if !ok {
		missesCounter.Add(1)
		return ""
	}
	hitsCounter.Add(1)
	return id
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

// FromCache gets an image from the LRU cache.
// Returns nil if the image is not in the cache.
// Keeps retrying for a maximum time of 8 seconds before returning nil.
func FromCache(id string) []byte {
	if id == "" {
		return nil
	}
	// Retry for a maximyum time of 8 seconds if the image is not in the cache
	startTime := time.Now()
	for {
		if time.Since(startTime) > 8*time.Second {
			return nil
		}
		data, ok := imagesLRU.Get(id)
		if ok {
			return data
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// IsInCache checks if an image is in the LRU cache.
func IsInCache(id string) bool {
	return imagesLRU.Contains(id)
}

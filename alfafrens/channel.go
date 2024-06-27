package alfafrens

import (
	"sync"

	"go.vocdoni.io/dvote/types"
)

// DefaultUpdateRetries is the default number of retries to fetch the subscribers
// of a cached channel.
const DefaultUpdateRetries = 5

// CachedChannel is a helper struct to keep track of the subscribers of a channel
// and avoid fetching the same data multiple times to bypass the rate limits.
type CachedChannel struct {
	addr        types.HexBytes
	subscribers map[uint64]bool
	mtx         sync.Mutex
}

// NewCachedChannel creates a new CachedChannel instance based on the given
// channel address.
func NewCachedChannel(addr string) *CachedChannel {
	return &CachedChannel{
		addr:        types.HexBytes(addr),
		subscribers: make(map[uint64]bool),
	}
}

// Update fetches the subscribers of the channel and updates the internal state.
func (ch *CachedChannel) Update(retries int) error {
	var err error
	var fids []uint64
	for i := 0; i < retries; i++ {
		fids, err = ChannelFids(ch.addr)
		if err == nil {
			break
		}
	}
	if err != nil {
		return err
	}
	ch.mtx.Lock()
	defer ch.mtx.Unlock()
	for _, fid := range fids {
		ch.subscribers[fid] = true
	}
	return nil
}

// IsSubscribed returns true if the given fid is subscribed to the channel.
func (ch *CachedChannel) IsSubscribed(fid uint64) bool {
	ch.mtx.Lock()
	defer ch.mtx.Unlock()
	return ch.subscribers[fid]
}

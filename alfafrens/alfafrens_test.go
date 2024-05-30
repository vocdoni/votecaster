package alfafrens

import (
	"encoding/hex"
	"testing"

	qt "github.com/frankban/quicktest"
)

const (
	testChannelAddress = "946ca533da30bc5632ee594d8157fd2bb3d29356"
	testFid            = uint64(239694)
)

func TestChannelFids(t *testing.T) {
	q := qt.New(t)
	channel, err := hex.DecodeString(testChannelAddress)
	q.Assert(err, qt.IsNil)

	fids, err := ChannelFids(channel)
	q.Assert(err, qt.IsNil)
	t.Logf("FIDs: %d", len(fids))
}

func TestChannelByFid(t *testing.T) {
	q := qt.New(t)
	channel, err := ChannelByFid(testFid)
	q.Assert(err, qt.IsNil)
	q.Assert(len(channel) > 0, qt.Equals, true)
	q.Assert(channel.String(), qt.Equals, testChannelAddress)
}

package alfafrens

import (
	"encoding/hex"
	"testing"

	qt "github.com/frankban/quicktest"
)

const (
	testChannelAddress = "8ee30662780088d2d2bfd0ced68b80bd7d629865"
	testFid            = uint64(237855)
)

func TestChannelFids(t *testing.T) {
	q := qt.New(t)
	channel, err := hex.DecodeString(testChannelAddress)
	q.Assert(err, qt.IsNil)

	fids, err := ChannelFids(channel)
	q.Assert(err, qt.IsNil)
	t.Logf("FIDs: %v", fids)
}

func TestChannelByFid(t *testing.T) {
	q := qt.New(t)
	channel, err := ChannelByFid(testFid)
	q.Assert(err, qt.IsNil)
	q.Assert(len(channel) > 0, qt.Equals, true)
	q.Assert(channel.String(), qt.Equals, testChannelAddress)
}

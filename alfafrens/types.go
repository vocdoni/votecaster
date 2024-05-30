package alfafrens

import (
	"encoding/json"
	"strconv"

	"go.vocdoni.io/dvote/types"
)

const (
	channelSubscribersURL = "https://alfafrens.com/api/v0/getChannelSubscribersAndStakes?channelAddress=%s"
	userByFidInfo         = "https://alfafrens.com/api/v0/getUserByFid?fid=%d"
)

// ChannelResponse represents the structure of the JSON response
type ChannelResponse struct {
	ID                            types.HexBytes `json:"id"`
	NumberOfSubscribers           int            `json:"numberOfSubscribers"`
	NumberOfStakers               int            `json:"numberOfStakers"`
	TotalSubscriptionFlowRate     *types.BigInt  `json:"totalSubscriptionFlowRate"`
	TotalSubscriptionInflowAmount *types.BigInt  `json:"totalSubscriptionInflowAmount"`
	TotalClaimed                  *types.BigInt  `json:"totalClaimed"`
	Owner                         types.HexBytes `json:"owner"`
	CurrentStaked                 *types.BigInt  `json:"currentStaked"`
	Members                       []Member       `json:"members"`
	Title                         string         `json:"title"`
	Bio                           string         `json:"bio"`
	HasMore                       bool           `json:"hasMore"`
}

// Member represents a single member in the members list
type Member struct {
	ID                             types.HexBytes `json:"id"`
	LastUpdatedTimestamp           string         `json:"lastUpdatedTimestamp"`
	Subscriber                     Subscriber     `json:"subscriber"`
	Channel                        Channel        `json:"channel"`
	IsSubscribed                   bool           `json:"isSubscribed"`
	IsStaked                       bool           `json:"isStaked"`
	CurrentStaked                  *types.BigInt  `json:"currentStaked"`
	TotalSubscriptionOutflowRate   *types.BigInt  `json:"totalSubscriptionOutflowRate"`
	TotalSubscriptionOutflowAmount *types.BigInt  `json:"totalSubscriptionOutflowAmount"`
	FID                            uint64Str      `json:"fid"`
}

// Subscriber represents the subscriber details
type Subscriber struct {
	ID types.HexBytes `json:"id"`
}

// Channel represents the channel details
type Channel struct {
	ID    types.HexBytes `json:"id"`
	Owner types.HexBytes `json:"owner"`
}

// User represents the structure of the JSON response
type User struct {
	UserAddress    types.HexBytes `json:"userAddress"`
	FID            uint64Str      `json:"fid"`
	Handle         string         `json:"handle"`
	ChannelAddress types.HexBytes `json:"channeladdress"`
}

// Uint64Str is a custom type that operates as a uint64 but marshals/unmarshals as a string
type uint64Str uint64

// MarshalJSON implements the json.Marshaler interface for Uint64Str
func (u uint64Str) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatUint(uint64(u), 10))
}

// UnmarshalJSON implements the json.Unmarshaler interface for Uint64Str
func (u *uint64Str) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	val, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}
	*u = uint64Str(val)
	return nil
}

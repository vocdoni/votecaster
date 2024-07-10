package farcasterapi

import (
	"context"
	"fmt"
)

// MaxCastBytes is the maximum number of bytes that a cast can have.
const MaxCastBytes = 350

var (
	// ErrNoDataFound is returned when there is no data found.
	ErrNoDataFound = fmt.Errorf("no data found")
	// ErrNoNewCasts is returned when there are no new casts.
	ErrNoNewCasts = fmt.Errorf("no new casts")
	// ErrChannelNotFound is returned when the requested channel is not found.
	ErrChannelNotFound = fmt.Errorf("channel not found")
)

type API interface {
	// SetFarcasterUser sets the farcaster user with the given fid and signer
	// (UUID or privKey).
	SetFarcasterUser(fid uint64, signer string) error
	// FID returns the fid of the farcaster user set in the API
	FID() uint64
	// Stop stops the API
	Stop() error
	// LastMentions retrieves the last mentions from the given timestamp, it
	// returns the messages in a slice of APIMessage, the last timestamp and an
	// error if something goes wrong
	LastMentions(ctx context.Context, timestamp uint64) ([]*APIMessage, uint64, error)
	// GetCast retrieves the cast with the given fid and hash, it returns the
	// message in an APIMessage struct and an error if something goes wrong.
	GetCast(ctx context.Context, fid uint64, hash string) (*APIMessage, error)
	// Publish publishes a new cast with the given content, it returns an error
	// if something goes wrong. It receives a slice of fids to mention that will
	// be used to complete the content with the mentions. The fids must be
	// placed in the mentions slice in the same order they are in the content.
	// It also receives a slice of embedURLS that will be used to embed the
	// given URLs in the content.
	Publish(ctx context.Context, content string, mentionFids []uint64, embedURLS ...string) error
	// Reply replies to a cast of the given fid with the given hash and content,
	// it returns an error if something goes wrong
	Reply(ctx context.Context, targetMsg *APIMessage, content string, mentionFids []uint64, embedURLS ...string) error
	// UserDataByFID retrieves the Userdata of the user with the given fid, if
	// something goes wrong, it returns an error
	UserDataByFID(ctx context.Context, fid uint64) (*Userdata, error)
	// UserDataByVerificationAddress retrieves the Userdata of the user with the
	// given verification address, if something goes wrong, it returns an error
	UserDataByVerificationAddress(ctx context.Context, address []string) ([]*Userdata, error)
	// WebhookHandler handles the incoming webhooks from the farcaster API
	WebhookHandler(body []byte) error
	// SignersFromFID retrieves the signers (appkeys) of the user with the given fid
	SignersFromFID(fid uint64) ([]string, error)
	// UserFollowers method returns the FIDs of the followers of the user with
	// the given id. If something goes wrong, it returns an error.
	UserFollowers(ctx context.Context, fid uint64) ([]uint64, error)
	// Channel method returns the channel with the given id. If something goes
	// wrong, it returns an error.
	Channel(ctx context.Context, channelID string) (*Channel, error)
	// ChannelFIDs method returns the FIDs of the users that follow the channel
	// with the given id. If something goes wrong, it returns an error. It
	// return an ErrChannelNotFound error if the channel does not exist to be
	// handled by the caller.
	ChannelFIDs(ctx context.Context, channelID string, progress chan int) ([]uint64, error)
	// ChannelExists method returns a boolean indicating if the channel with the
	// given id exists. If something goes wrong checking the channel existence,
	// it returns an error.
	ChannelExists(ctx context.Context, channelID string) (bool, error)
	// FindChannel method returns the channels that matches with the handle (or
	// the part of it) provided. If something goes wrong, it returns an error.
	FindChannel(ctx context.Context, query string, adminFid uint64) ([]*Channel, error)
	// DirectMessage method sends a direct message to the user with the given
	// fid. If something goes wrong, it returns an error.
	DirectMessage(ctx context.Context, content string, to uint64) error
}

// ParentAPIMessage is a struct that represents the parent message of an
// APIMessage that does not includes the parent message itself, but only the
// fid of the author and hash as reference of the parent message.
type ParentAPIMessage struct {
	FID  uint64
	Hash string
}

// APIMessage is a struct that represents a message in the farcaster API.
type APIMessage struct {
	IsMention bool
	Content   string
	Author    uint64
	Hash      string
	Parent    *ParentAPIMessage
	Embeds    []string
}

// Userdata is a struct that represents the user data in the farcaster API.
type Userdata struct {
	FID                    uint64
	Username               string
	Displayname            string
	CustodyAddress         string
	VerificationsAddresses []string
	Signers                []string
	Avatar                 string
	Bio                    string
}

// Channel is a struct that represents a channel in the farcaster API.
type Channel struct {
	ID          string
	Name        string
	Description string
	Followers   int
	Image       string
	URL         string
}

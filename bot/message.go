package bot

import (
	"encoding/hex"
	"strings"
)

const (
	MESSAGE_TYPE_CAST_ADD = "MESSAGE_TYPE_CAST_ADD"
)

// CastAddBody represents the body of a cast add message.
type CastAddBody struct {
	Text      string `json:"text"`
	ParentURL string `json:"parentUrl"`
}

// MessageData represents the data of a message.
type MessageData struct {
	Type        string       `json:"type"`
	From        uint64       `json:"fid"`
	Timestamp   uint64       `json:"timestamp"`
	CastAddBody *CastAddBody `json:"castAddBody,omitempty"`
}

// Message represents a message from the API, it includes the data and the hash
// of the message.
type Message struct {
	Data    *MessageData `json:"data"`
	HexHash string       `json:"hash"`
}

// IsMention returns true if the message is a mention.
func (m *Message) IsMention() bool {
	return m.Data.Type == MESSAGE_TYPE_CAST_ADD && m.Data.CastAddBody != nil && m.Data.CastAddBody.Text != ""
}

// Mention returns the text of a cast.
func (m *Message) Mention() string {
	return m.Data.CastAddBody.Text
}

// Author returns the author of the message.
func (m *Message) Author() uint64 {
	return m.Data.From
}

// Timestamp returns the timestamp of the message.
func (m *Message) Hash() ([]byte, error) {
	hash := strings.TrimPrefix(m.HexHash, "0x")
	return hex.DecodeString(hash)
}

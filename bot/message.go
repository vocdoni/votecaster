package bot

import (
	"encoding/hex"
	"strings"
)

const (
	MESSAGE_TYPE_CAST_ADD = "MESSAGE_TYPE_CAST_ADD"
)

type CastAddBody struct {
	Text      string `json:"text"`
	ParentURL string `json:"parentUrl"`
}

type MessageData struct {
	Type        string       `json:"type"`
	From        uint64       `json:"fid"`
	Timestamp   uint64       `json:"timestamp"`
	CastAddBody *CastAddBody `json:"castAddBody,omitempty"`
}

type Message struct {
	Data    *MessageData `json:"data"`
	HexHash string       `json:"hash"`
}

func (m *Message) IsMention() bool {
	return m.Data.Type == MESSAGE_TYPE_CAST_ADD && m.Data.CastAddBody != nil && m.Data.CastAddBody.Text != ""
}

func (m *Message) Mention() string {
	return m.Data.CastAddBody.Text
}

func (m *Message) Author() uint64 {
	return m.Data.From
}

func (m *Message) Hash() ([]byte, error) {
	hash := strings.TrimPrefix(m.HexHash, "0x")
	return hex.DecodeString(hash)
}

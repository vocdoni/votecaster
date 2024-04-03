package hub

type hubCastEmbeds struct {
	Url string `json:"url"`
}

type hubCastAddBody struct {
	Text              string           `json:"text"`
	ParentURL         string           `json:"parentUrl"`
	Mentions          []uint64         `json:"mentions"`
	MentionsPositions []uint64         `json:"mentionsPositions"`
	Embeds            []*hubCastEmbeds `json:"embeds"`
}

type hubMessageData struct {
	Type        string          `json:"type"`
	From        uint64          `json:"fid"`
	Timestamp   uint64          `json:"timestamp"`
	CastAddBody *hubCastAddBody `json:"castAddBody,omitempty"`
}

type hubMessage struct {
	Data    *hubMessageData `json:"data"`
	HexHash string          `json:"hash"`
}

type hubMessageResponse struct {
	Messages []*hubMessage `json:"messages"`
}

type usernameProofs struct {
	Username       string `json:"name"`
	CustodyAddress string `json:"owner"`
	FID            uint64 `json:"fid"`
	Type           string `json:"type"`
	Timestamp      uint64 `json:"timestamp"`
}

type userdataResponse struct {
	Proofs []*usernameProofs `json:"proofs"`
}

type verification struct {
	Address string `json:"address"`
}

type verificationData struct {
	Type         string        `json:"type"`
	Verification *verification `json:"verificationAddEthAddressBody"`
	Signer       string        `json:"signer"`
}

type verificationMessage struct {
	Data *verificationData `json:"data"`
}

type verificationsResponse struct {
	Messages []*verificationMessage `json:"messages"`
}

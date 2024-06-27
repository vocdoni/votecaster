package hub

type hubCastEmbeds struct {
	Url string `json:"url"`
}

type hubParentCast struct {
	FID  uint64 `json:"fid"`
	Hash string `json:"hash"`
}

type hubCastAddBody struct {
	Text              string           `json:"text"`
	ParentURL         string           `json:"parentUrl"`
	Mentions          []uint64         `json:"mentions"`
	MentionsPositions []uint64         `json:"mentionsPositions"`
	Embeds            []*hubCastEmbeds `json:"embeds"`
	ParentCast        *hubParentCast   `json:"parentCastId"`
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

type hubReactionBody struct {
	Type string `json:"type"`
}

type hubReactionData struct {
	Type   string           `json:"type"`
	Author uint64           `json:"fid"`
	Body   *hubReactionBody `json:"reactionBody"`
}

type hubReaction struct {
	Data *hubReactionData `json:"data"`
	Hash string           `json:"hash"`
}

type hubReactionsResponse struct {
	Reactions []*hubReaction `json:"messages"`
}

type usernameProofs struct {
	Username       string `json:"name"`
	CustodyAddress string `json:"owner"`
	FID            uint64 `json:"fid"`
	Type           string `json:"type"`
	Timestamp      uint64 `json:"timestamp"`
}

type custodyAddressResponse struct {
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

type hubUserDataBody struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type hubUserData struct {
	Type      string           `json:"type"`
	FID       uint64           `json:"fid"`
	Timestamp uint64           `json:"timestamp"`
	Body      *hubUserDataBody `json:"userDataBody"`
}

type hubUserDataMessage struct {
	Data   *hubUserData `json:"data"`
	Hash   string       `json:"hash"`
	Signer string       `json:"signer"`
}

type hubUserdataResponse struct {
	Messages []*hubUserDataMessage `json:"messages"`
}

package neynar

type castEmbed struct {
	Url string `json:"url"`
}

type castPostRequest struct {
	Signer string       `json:"signer_uuid"`
	Text   string       `json:"text"`
	Parent string       `json:"parent"`
	Embeds []*castEmbed `json:"embeds"`
}

type userdataV1 struct {
	Fid                    uint64   `json:"fid"`
	Username               string   `json:"username"`
	CustodyAddress         string   `json:"custodyAddress"`
	VerificationsAddresses []string `json:"verifications"`
}

type userdataV1Result struct {
	User *userdataV1 `json:"user"`
}

type usersdataV1Result struct {
	Users      []*userdataV1 `json:"users"`
	NextCursor *cursor       `json:"next"`
}

type userdataV1Response struct {
	Result *userdataV1Result `json:"result"`
}

type UsersdataV1Response struct {
	Result *usersdataV1Result `json:"result"`
}

type verifiedAddressesV2 struct {
	EthAddresses []string `json:"eth_addresses"`
	SolAddresses []string `json:"sol_addresses"`
}

type userdataV2 struct {
	Object            string               `json:"object"`
	Fid               uint64               `json:"fid"`
	CustodyAddress    string               `json:"custody_address"`
	Username          string               `json:"username"`
	DisplayName       string               `json:"display_name"`
	PfpUrl            string               `json:"pfp_url"`
	Profile           userProfile          `json:"profile"`
	FollowerCount     int                  `json:"follower_count"`
	FollowingCount    int                  `json:"following_count"`
	Verifications     []string             `json:"verifications"`
	VerifiedAddresses *verifiedAddressesV2 `json:"verified_addresses"`
	ActiveStatus      string               `json:"active_status"`
}

type cursor struct {
	Cursor string `json:"cursor"`
}

type userdataV2Result struct {
	Users      []*userdataV2 `json:"users"`
	NextCursor *cursor       `json:"next"`
}

type parentCastAuthor struct {
	FID uint64 `json:"fid"`
}

type castWebhookData struct {
	Object       string            `json:"object"`
	Hash         string            `json:"hash"`
	Text         string            `json:"text"`
	Timestamp    string            `json:"timestamp"`
	Author       *userdataV2       `json:"author"`
	ParentURL    string            `json:"parent_url"`
	ParentHash   string            `json:"parentHash"`
	Embeds       []*castEmbed      `json:"embeds"`
	ParentAuthor *parentCastAuthor `json:"parentAuthor"`
}

type castsWebhookRequest struct {
	Type string           `json:"type"`
	Data *castWebhookData `json:"data"`
}

type castResponseV2 struct {
	Data *castWebhookData `json:"cast"`
}

// ---

type userProfile struct {
	Bio userBio `json:"bio"`
}

type userBio struct {
	Text string `json:"text"`
}

type warpcastChannel struct {
	ImageURL    string `json:"imageUrl"`
	Followers   int    `json:"followerCount"`
	Name        string `json:"name"`
	Description string `json:"description"`
	ID          string `json:"key"`
}

type warpcastChannelResult struct {
	Channel *warpcastChannel `json:"channel"`
}

type warpcastChannelResponse struct {
	Result *warpcastChannelResult `json:"result"`
}

// HUB API

type HubAPIResponse struct {
	Messages      []*HubMessage `json:"messages"`
	NextPageToken string        `json:"nextPageToken"`
}

type HubMessage struct {
	Data            *HubData `json:"data"`
	Hash            string   `json:"hash"`
	HashScheme      string   `json:"hashScheme"`
	Signature       string   `json:"signature"`
	SignatureScheme string   `json:"signatureScheme"`
	Signer          string   `json:"signer"`
}

type HubData struct {
	Type                          string                            `json:"type"`
	Fid                           int                               `json:"fid"`
	Timestamp                     int64                             `json:"timestamp"`
	Network                       string                            `json:"network"`
	VerificationAddAddressBody    *HubVerificationAddAddressBody    `json:"verificationAddAddressBody"`
	VerificationAddEthAddressBody *HubVerificationAddEthAddressBody `json:"verificationAddEthAddressBody"`
}

type HubVerificationAddAddressBody struct {
	Address          string `json:"address"`
	ClaimSignature   string `json:"claimSignature"`
	BlockHash        string `json:"blockHash"`
	VerificationType int    `json:"verificationType"`
	ChainId          int    `json:"chainId"`
	Protocol         string `json:"protocol"`
	EthSignature     string `json:"ethSignature"`
}

type HubVerificationAddEthAddressBody struct {
	Address          string `json:"address"`
	ClaimSignature   string `json:"claimSignature"`
	BlockHash        string `json:"blockHash"`
	VerificationType int    `json:"verificationType"`
	ChainId          int    `json:"chainId"`
	Protocol         string `json:"protocol"`
	EthSignature     string `json:"ethSignature"`
}

const (
	HUB_MESSAGE_TYPE_VERIFICATION = "MESSAGE_TYPE_VERIFICATION_ADD_ETH_ADDRESS"
)

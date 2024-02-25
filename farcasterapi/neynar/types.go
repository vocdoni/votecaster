package neynar

type NotificationAuthor struct {
	FID uint64 `json:"fid"`
}

type Notification struct {
	Hash      string             `json:"hash"`
	Author    NotificationAuthor `json:"author"`
	Type      string             `json:"type"`
	Text      string             `json:"text"`
	Timestamp string             `json:"timestamp"`
}

type NextNotificationCursor struct {
	Cursor string `json:"cursor"`
}

type NotificationsResult struct {
	Notifications []*Notification        `json:"notifications"`
	NextCursor    NextNotificationCursor `json:"next"`
}

type NotificationsResponse struct {
	Result *NotificationsResult `json:"result"`
}

type CastPostRequest struct {
	Signer string `json:"signer_uuid"`
	Text   string `json:"text"`
	Parent string `json:"parent"`
}

type UserdataV1 struct {
	FID                    uint64   `json:"fid"`
	Username               string   `json:"username"`
	CustodyAddress         string   `json:"custodyAddress"`
	VerificationsAddresses []string `json:"verifications"`
}

type UserdataV2 struct {
	Object            string            `json:"object"`
	Fid               uint64            `json:"fid"`
	CustodyAddress    string            `json:"custody_address"`
	Username          string            `json:"username"`
	DisplayName       string            `json:"display_name"`
	PfpUrl            string            `json:"pfp_url"`
	Profile           UserProfile       `json:"profile"`
	FollowerCount     int               `json:"follower_count"`
	FollowingCount    int               `json:"following_count"`
	Verifications     []string          `json:"verifications"`
	VerifiedAddresses VerifiedAddresses `json:"verified_addresses"`
	ActiveStatus      string            `json:"active_status"`
}

type UserdataResult struct {
	User *UserdataV1 `json:"user"`
}

type UserdataResponse struct {
	Result *UserdataResult `json:"result"`
}

// ---

type UserProfile struct {
	Bio UserBio `json:"bio"`
}

type UserBio struct {
	Text string `json:"text"`
}

type VerifiedAddresses struct {
	EthAddresses []string `json:"eth_addresses"`
	SolAddresses []string `json:"sol_addresses"`
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

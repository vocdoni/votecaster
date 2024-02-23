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
	FID                    uint64   `json:"fid"`
	Username               string   `json:"username"`
	CustodyAddress         string   `json:"custody_address"`
	VerificationsAddresses []string `json:"verifications"`
}

type UserdataResult struct {
	User *UserdataV1 `json:"user"`
}

type UserdataResponse struct {
	Result *UserdataResult `json:"result"`
}

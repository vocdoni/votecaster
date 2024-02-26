package bot

import "fmt"

var (
	ErrAPINotSet          = fmt.Errorf("api not set")
	ErrBotFIDNotSet       = fmt.Errorf("bot fid not set")
	ErrPrivateKeyNotSet   = fmt.Errorf("private key not set")
	ErrDecodingPrivateKey = fmt.Errorf("error decoding provided private key")
	ErrEndpointNotSet     = fmt.Errorf("endpoint not set")
	ErrNoNewCasts         = fmt.Errorf("no new casts")
)

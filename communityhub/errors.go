package communityhub

import "fmt"

var (
	ErrMissingDB           = fmt.Errorf("missing db")
	ErrDecodingCreationLog = fmt.Errorf("error decoding creation log from the community hub contract")
	ErrGettingCommunity    = fmt.Errorf("error getting community data from the community hub contract")
)

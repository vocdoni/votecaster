package communityhub

import "fmt"

var (
	// ErrMissingDB is returned when no database is provided during CommunituHub
	// initialization
	ErrMissingDB = fmt.Errorf("missing db")
	// ErrDecodingCommunity is returned when an error occurs while decoding a
	// community from the community hub contract
	ErrDecodingCommunity = fmt.Errorf("error decoding creation log from the community hub contract")
	// ErrGettingCommunity is returned when an error occurs while getting
	// community data from the community hub contract
	ErrGettingCommunity = fmt.Errorf("error getting community data from the community hub contract")
	// ErrCommunityNotFound is returned when the community is not found in the
	// contract
	ErrCommunityNotFound = fmt.Errorf("community not found")
)

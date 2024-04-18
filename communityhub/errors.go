package communityhub

import "fmt"

var (
	// ErrMissingDB is returned when no database is provided during CommunituHub
	// initialization
	ErrMissingDB = fmt.Errorf("missing db")
	// ErrDecodingCreationLog is returned when an error occurs while
	// decoding the creation log from the community hub contract
	ErrDecodingCreationLog = fmt.Errorf("error decoding creation log from the community hub contract")
	// ErrGettingCommunity is returned when an error occurs while getting
	// community data from the community hub contract
	ErrGettingCommunity = fmt.Errorf("error getting community data from the community hub contract")
)

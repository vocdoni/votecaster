package communityhub

import "fmt"

var (
	// ErrMissingDB is returned when no database is provided during CommunituHub
	// initialization
	ErrMissingDB = fmt.Errorf("missing db")
	// ErrClosedDB is returned when the database is closed
	ErrClosedDB = fmt.Errorf("db is closed")
	// ErrWeb3Client is returned when no web3 client is found in the provided
	// web3 pool
	ErrWeb3Client = fmt.Errorf("failed to get web3 client from the pool")
	// ErrInitContract is returned when the initialization of the contract fails
	// during CommunityHub initialization
	ErrInitContract = fmt.Errorf("failed to initialize contract")
	// ErrMissingContract is returned when the creation of a auth transactor
	// with the private key provided fails
	ErrCreatingSigner = fmt.Errorf("failed to create auth transactor signer")
	// ErrNoPrivKeyConfigured is returned when no private key is provided during
	// CommunityHub initialization and then a transaction is attempted to be
	// sent with a signer
	ErrNoPrivKeyConfigured = fmt.Errorf("no private key defined")
	// ErrDecodingCommunity is returned when an error occurs while decoding a
	// community from the community hub contract
	ErrDecodingCommunity = fmt.Errorf("error decoding creation log from the community hub contract")
	// ErrGettingCommunity is returned when an error occurs while getting
	// community data from the community hub contract
	ErrGettingCommunity = fmt.Errorf("error getting community data from the community hub contract")
	// ErrCommunityNotFound is returned when the community is not found in the
	// contract
	ErrCommunityNotFound = fmt.Errorf("community not found")
	// ErrDisabledCommunity is returned when the community is disabled in the
	// contract
	ErrDisabledCommunity = fmt.Errorf("community is disabled")
	// ErrGettingResults is returned when an error occurs while getting results
	// from the community hub contract
	ErrGettingResults = fmt.Errorf("error getting results from the community hub contract")
	// ErrSettingResults is returned when an error occurs while setting results
	// in the community hub contract
	ErrSettingResults = fmt.Errorf("error setting results in the community hub contract")
	// ErrUnknownCensusType is returned when an unknown census type is found
	// while encoding or decoding a community for the community hub contract
	ErrUnknownCensusType = fmt.Errorf("unknown census type")
	// ErrNoChannelProvided is returned when no channel is provided during the
	// creation of a community with a channel census type
	ErrNoChannelProvided = fmt.Errorf("no channel provided")
	// ErrBadCensusAddressees is returned when the census addressees are not
	// provided in the correct format or are empty
	ErrBadCensusAddressees = fmt.Errorf("bad community census addressees")
	// ErrAddCommunity is returned when an error occurs while adding a new
	// community from the community hub contract to the database
	ErrAddCommunity = fmt.Errorf("error adding community to the database")
	// ErrSendingTx is returned when an error occurs while sending the a
	// transaction to the community hub contract
	ErrSendingTx = fmt.Errorf("error estimating gas")
)

package communityhub

import "fmt"

var (
	// ErrMissingDB is returned when no database is provided during CommunituHub
	// initialization
	ErrMissingDB = fmt.Errorf("missing db")
	// ErrMissingCensus3 is returned when no census3 client is provided during
	// CommunityHub initialization
	ErrMissingCensus3 = fmt.Errorf("missing db")
	// ErrMissingContracts is returned when no contracts addresses and chain ids
	// are provided during CommunityHub initialization
	ErrMissingContracts = fmt.Errorf("missing contracts addresses and chain id")
	// ErrMissingChainAliases is returned when no chain aliases are provided
	// during CommunityHub initialization
	ErrMissingChainAliases = fmt.Errorf("missing chain aliases by chain id")
	// ErrNoValidContracts is returned when no valid contracts are provided
	// during CommunityHub initialization
	ErrNoValidContracts = fmt.Errorf("no valid contracts provided")
	// ErrContractNotFound is returned when the contract is not found in the
	// provided contracts list
	ErrContractNotFound = fmt.Errorf("contract not found for the provided chain id")
	// ErrClosedDB is returned when the database is closed
	ErrClosedDB = fmt.Errorf("db is closed")
	// ErrWeb3Client is returned when no web3 client is found in the provided
	// web3 pool
	ErrWeb3Client = fmt.Errorf("failed to get web3 client from the pool")
	// ErrInitContract is returned when the initialization of the contract fails
	// during CommunityHub initialization
	ErrInitContract = fmt.Errorf("failed to initialize contract")
	// ErrCreatingSigner is returned when the signer cannot be created
	ErrCreatingSigner = fmt.Errorf("failed to create auth transactor signer")
	// ErrInitializingPrivateKey is returned when the private key cannot be
	// initialized
	ErrInitializingPrivateKey = fmt.Errorf("failed to initialize private key")
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
	// ErrCommunityIDMismatch is returned when the community ID does not match
	// the expected community ID
	ErrCommunityIDMismatch = fmt.Errorf("community ID mismatch")
	// ErrSettingCommunity is returned when an error occurs while setting
	// community data in the community hub contract
	ErrSettingCommunity = fmt.Errorf("error setting community data in the community hub contract")
	// ErrInvalidCommunityData is returned when invalid community data is
	// provided
	ErrInvalidCommunityData = fmt.Errorf("invalid community data provided")
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
	// ErrNoUserRefProvided is returned when no user reference is provided
	// during the creation of a community with a followers census type
	ErrNoUserRefProvided = fmt.Errorf("no user reference provided")
	// ErrBadCensusAddressees is returned when the census addressees are not
	// provided in the correct format or are empty
	ErrBadCensusAddressees = fmt.Errorf("bad community census addressees")
	// ErrNoAdminCreator is returned when the provided admin list does not
	// contain the creator of the community
	ErrNoAdminCreator = fmt.Errorf("the creator must be an admin")
	// ErrAddCommunity is returned when an error occurs while adding a new
	// community from the community hub contract to the database
	ErrAddCommunity = fmt.Errorf("error adding community to the database")
	// ErrSendingTx is returned when an error occurs while sending the a
	// transaction to the community hub contract
	ErrSendingTx = fmt.Errorf("error estimating gas")
	// ErrEncodeCommunityID is returned when the community ID cannot be encoded
	// from the chain short name and the ID
	ErrEncodeCommunityID = fmt.Errorf("error encoding chain community ID")
	// ErrDecodeCommunityID is returned when the ID and the chain
	// short name cannot be decoded from the community ID
	ErrDecodeCommunityID = fmt.Errorf("error decoding chain community ID")
)

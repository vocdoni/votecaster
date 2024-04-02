// features package defines the features that the services supports based on
// the user reputation. Include constants for the features and the related
// reputation thresholds. Also, include the features names and descriptions,
// and methods to check if a feature is allowed based on the user reputation.
package features

// Feature represents a feature that the service supports based on the user
// reputation, and it is an alias of int.
type Feature int

const (
	// NOTIFY_USERS is a feature that allows to notify users when the election
	// starts.
	NOTIFY_USERS Feature = iota
)

// reputationThresholdsByFeature is a map that contains the reputation
// thresholds for each feature.
var reputationThresholdsByFeature = map[Feature]uint32{
	NOTIFY_USERS: 5,
}

// featuresStr is a map that contains the string representation of each feature.
var featuresStr = map[Feature]string{
	NOTIFY_USERS: "notifyUsers",
}

// featuresNames is a map that contains the name of each feature.
var featuresNames = map[Feature]string{
	NOTIFY_USERS: "Notify users",
}

// featuresDescription is a map that contains the description of each feature.
var featuresDescription = map[Feature]string{
	NOTIFY_USERS: "Allows to notify users when the election starts." +
		"The users must accept receiving notifications and can mute them at any time.",
}

// String returns the string representation of the feature.
func (f Feature) String() string {
	return featuresStr[f]
}

// Name returns the name of the feature.
func (f Feature) Name() string {
	return featuresNames[f]
}

// Description returns the description of the feature.
func (f Feature) Description() string {
	return featuresDescription[f]
}

// IsAllowed returns true if the feature is allowed based on the user reputation
// provided.
func IsAllowed(f Feature, userReputation uint32) bool {
	return userReputation >= reputationThresholdsByFeature[f]
}

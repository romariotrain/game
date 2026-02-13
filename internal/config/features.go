package config

// Features controls optional systems and UI visibility.
type Features struct {
	MinimalMode            bool
	Combat                 bool
	Events                 bool
	FailExpiredExpeditions bool
}

// DefaultFeatures returns the default feature configuration.
func DefaultFeatures() Features {
	return Features{
		MinimalMode:            true,
		Combat:                 true,
		Events:                 false,
		FailExpiredExpeditions: true,
	}
}

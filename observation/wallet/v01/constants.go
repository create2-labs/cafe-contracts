package v01

// Normative contract identifiers for cafe.discovery.wallet.observed at wire version v0.1.
const (
	EventTypeWalletObserved = "cafe.discovery.wallet.observed"
	EventVersion            = "v0.1"

	// ProducerCafeDiscovery is the normative JSON "producer" value for events emitted by the
	// Discovery service. Validate requires it so inbound messages match the expected contract
	// revision (distinct from auth: it is a wire-level producer label, not proof of origin).
	ProducerCafeDiscovery = "cafe-discovery"
)

package yaml

// Volumes represents a list of service volumes in compose file.
// It has several representation, hence this specific struct.
type Volumes struct {
	Volumes []Volume
}

// Volume represent a service volume
type Volume struct {
	name string
}

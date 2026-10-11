// Package platform contains read-only host services. It exposes no cleanup API.
package platform

type Capacity struct{ Total, Available uint64 }

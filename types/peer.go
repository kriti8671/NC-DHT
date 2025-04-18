package types

import "go.dedis.ch/kyber/v3"

// Peer represents a node inside a quorum
type Peer struct {
	ID           int
	PrivKeyShare kyber.Scalar // Share of quorum private key
	PubKeyShare  kyber.Point  // Share of quorum public key
}

package types

import (
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/edwards25519"
	"go.dedis.ch/kyber/v3/share"
)

type Quorum struct {
	ID         int
	Peers      []*Peer
	PubKey     kyber.Point
	PrivShares []*share.PriShare
}

// BuildQuorum creates a quorum of `size` peers with threshold `t`
func BuildQuorum(id, size, threshold int) (*Quorum, error) {
	suite := edwards25519.NewBlakeSHA256Ed25519()
	poly := share.NewPriPoly(suite, threshold, nil, suite.RandomStream())
	privShares := poly.Shares(size)
	pub := poly.Commit(nil).Eval(0).V

	q := &Quorum{
		ID:         id,
		Peers:      make([]*Peer, size),
		PubKey:     pub,
		PrivShares: privShares,
	}

	for i := 0; i < size; i++ {
		peer := &Peer{
			ID:           i,
			PrivKeyShare: privShares[i].V,
			PubKeyShare:  suite.Point().Mul(privShares[i].V, nil),
		}
		q.Peers[i] = peer
	}
	return q, nil
}

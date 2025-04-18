package crypto

import (
	"errors"
	"ncdht/types"

	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/group/edwards25519"
	"go.dedis.ch/kyber/v3/sign/schnorr"
)

var suite = edwards25519.NewBlakeSHA256Ed25519()

func SignMessage(msg []byte, peer *types.Peer, index int) ([]byte, error) {
	sig, err := schnorr.Sign(suite, peer.PrivKeyShare, msg)
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func VerifySignature(msg, sig []byte, pubKey kyber.Point) error {
	return schnorr.Verify(suite, pubKey, msg, sig)
}

func CombineSignatures(shares [][]byte) ([]byte, error) {
	if len(shares) == 0 {
		return nil, errors.New("no signature shares")
	}
	// Simplified: return first valid signature
	return shares[0], nil
}

// package crypto

// import (
// 	"errors"
// 	"ncdht/types"

// 	"go.dedis.ch/kyber/v3"
// 	"go.dedis.ch/kyber/v3/group/edwards25519"
// 	"go.dedis.ch/kyber/v3/sign/bls"
// )

// // Suite used globally
// var suite = edwards25519.NewBlakeSHA256Ed25519()

// // SignMessage returns a signature share for a message by a peer
// func SignMessage(msg []byte, peer *types.Peer, index int) ([]byte, error) {
// 	sig, err := bls.Sign(suite, peer.PrivKeyShare, msg)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return sig, nil
// }

// // VerifySignature checks the full quorum signature
// func VerifySignature(msg, sig []byte, pubKey kyber.Point) error {
// 	err := bls.Verify(suite, pubKey, msg, sig)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// // CombineSignatures takes threshold+1 signature shares and combines into full signature
// func CombineSignatures(shares [][]byte) ([]byte, error) {
// 	if len(shares) == 0 {
// 		return nil, errors.New("no signature shares")
// 	}
// 	// In simplified BLS, signature shares are directly combined by XOR (if additive)
// 	// Here we assume everyone signs same message, so any valid signature can be used
// 	// For real DKG, use pairing-based BLS threshold combining
// 	// Kyber bls doesn't support threshold combine natively — we simulate
// 	return shares[0], nil // ← placeholder (real BLS-TSS requires pairing lib)
// }

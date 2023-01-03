package model

import "github.com/spacemeshos/go-spacemesh/signing"

type signer interface {
	Sign(msg []byte) []byte
	PublicKey() *signing.PublicKey
}
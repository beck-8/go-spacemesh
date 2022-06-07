package core

import (
	"github.com/spacemeshos/go-scale"

	"github.com/spacemeshos/go-spacemesh/hash"
)

// Hash bytes into sha256 hash.
func Hash(bufs ...[]byte) Hash32 {
	hasher := hash.New()
	for _, buf := range bufs {
		hasher.Write(buf)
	}
	var rst Hash32
	hasher.Sum(rst[:])
	return rst
}

// ComputePrincipal address as the last 20 bytes from sha256(scale(template || nonce || args)).
func ComputePrincipal(template Address, args scale.Encodable) Address {
	hasher := hash.New()
	encoder := scale.NewEncoder(hasher)
	template.EncodeScale(encoder)
	args.EncodeScale(encoder)
	hash := hasher.Sum(nil)
	var rst Address
	copy(rst[:], hash[12:])
	return rst
}

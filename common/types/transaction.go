package types

import (
	"bytes"
	"fmt"
	xdr "github.com/nullstyle/go-xdr/xdr3"
	"github.com/spacemeshos/ed25519"
	"github.com/spacemeshos/go-spacemesh/log"
	"github.com/spacemeshos/go-spacemesh/signing"
	"strings"
)

// TransactionType is a transaction's kind and signing scheme
type TransactionType byte

const (
	// TxSimpleCoinEd is a simple coin transaction with ed signing scheme
	TxSimpleCoinEd TransactionType = 0
	// TxSimpleCoinEdPlus is a simple coin transaction with ed++ signing scheme
	TxSimpleCoinEdPlus TransactionType = 1
	// TxCallAppEd is a exec app transaction with ed signing scheme
	TxCallAppEd TransactionType = 2
	// TxCallAppEdPlus is a exec app transaction with ed++ signing scheme
	TxCallAppEdPlus TransactionType = 3
	// TxSpawnAppEd is a spawn app transaction with ed signing scheme
	TxSpawnAppEd TransactionType = 4
	// TxSpawnAppEdPlus is a spawn app transaction with ed++ signing scheme
	TxSpawnAppEdPlus TransactionType = 5

	// for support code transition to new transactions abstraction

	// TxOldCoinEd is a old coin transaction with ed signing scheme
	TxOldCoinEd TransactionType = 6
	// TxOldCoinEdPlus is a old coin transaction with ed++ signing scheme
	TxOldCoinEdPlus TransactionType = 7
)

// String returns string representation for TransactionType
func (tt TransactionType) String() string {
	switch tt {
	case TxSimpleCoinEd:
		return "TxSimpleCoinEd"
	case TxSimpleCoinEdPlus:
		return "TxSimpleCoinEdPlus"
	case TxCallAppEd:
		return "TxCallAppEd"
	case TxCallAppEdPlus:
		return "TxCallAppEdPlus"
	case TxSpawnAppEd:
		return "TxSpawnAppEd"
	case TxSpawnAppEdPlus:
		return "TxSpawnAppEdPlus"
	case TxOldCoinEd:
		return "TxOldCoinEd"
	case TxOldCoinEdPlus:
		return "TxOldCoinEdPlus"
	default:
		return "UnknownTransactionType"
	}
}

// IsEdPlus returns true if transaction type has Ed++ signing scheme
func (tt TransactionType) IsEdPlus() bool {
	switch tt {
	case TxSimpleCoinEdPlus, TxCallAppEdPlus, TxSpawnAppEdPlus, TxOldCoinEdPlus:
		return true
	case TxSimpleCoinEd, TxCallAppEd, TxSpawnAppEd, TxOldCoinEd:
		return false
	}
	// it must be impossible (in theory)
	panic(fmt.Errorf("unknown transaction type"))
}

// Decode decodes transaction bytes into the transaction object
func (tt TransactionType) Decode(pubKey TxPublicKey, signature TxSignature, txid TransactionID, data []byte) (Transaction, error) {
	switch tt {
	case TxSimpleCoinEd, TxSimpleCoinEdPlus:
		return DecodeSimpleCoinTx(data, signature, pubKey, txid, tt)
	case TxCallAppEd, TxCallAppEdPlus:
		return DecodeCallAppTx(data, signature, pubKey, txid, tt)
	case TxSpawnAppEd, TxSpawnAppEdPlus:
		return DecodeSpawnAppTx(data, signature, pubKey, txid, tt)
	case TxOldCoinEd, TxOldCoinEdPlus:
		return DecodeOldCoinTx(data, signature, pubKey, txid, tt)
	}
	// it must be impossible (in theory)
	return nil, fmt.Errorf("unknown transaction type")
}

// EdPlusTransactionFactory allowing to create transactions with Ed++ signing scheme
type EdPlusTransactionFactory interface {
	NewEdPlus() IncompleteTransaction
}

// EdTransactionFactory allowing to create transactions with Ed signing scheme
type EdTransactionFactory interface {
	NewEd() IncompleteTransaction
}

// IncompleteTransaction is the incomplete transaction having just a header
type IncompleteTransaction interface {
	fmt.Stringer // String()string

	AuthenticationMessage() (TransactionAuthenticationMessage, error)
	Type() TransactionType

	// extract internal transaction structure
	Extract(interface{}) bool

	// common attributes
	// they use Get prefix to not mess with struct attributes
	//    and do not pass function address to log/printf instead value

	GetRecipient() Address
	GetAmount() uint64
	GetNonce() uint64
	GetGasLimit() uint64
	GetGasPrice() uint64
	GetFee(gas uint64) uint64
}

// SignTransaction signs incomplete transaction and returns completed transaction object
func SignTransaction(itx IncompleteTransaction, signer *signing.EdSigner) (tx Transaction, err error) {
	txm, err := itx.AuthenticationMessage()
	if err != nil {
		return
	}
	stx, err := txm.Sign(signer)
	if err != nil {
		return
	}
	return stx.Decode()
}

// Transaction is a completed transaction interface
type Transaction interface {
	IncompleteTransaction

	Origin() Address
	ID() TransactionID
	Hash32() Hash32
	ShortString() string
	PubKey() TxPublicKey
	Signature() TxSignature
	Encode() (SignedTransaction, error)
}

// TransactionID is a 32-byte sha256 sum of the transaction, used as an identifier.
type TransactionID Hash32

// Hash32 returns the TransactionID as a Hash32.
func (id TransactionID) Hash32() Hash32 {
	return Hash32(id)
}

// ShortString returns a the first 10 characters of the ID, for logging purposes.
func (id TransactionID) ShortString() string {
	return id.Hash32().ShortString()
}

// String returns a hexadecimal representation of the TransactionID with "0x" prepended, for logging purposes.
// It implements the fmt.Stringer interface.
func (id TransactionID) String() string {
	return id.Hash32().String()
}

// Bytes returns the TransactionID as a byte slice.
func (id TransactionID) Bytes() []byte {
	return id[:]
}

// Field returns a log field. Implements the LoggableField interface.
func (id TransactionID) Field() log.Field { return id.Hash32().Field("tx_id") }

// TxIdsField returns a list of loggable fields for a given list of IDs
func TxIdsField(ids []TransactionID) log.Field {
	strs := []string{}
	for _, a := range ids {
		strs = append(strs, a.ShortString())
	}
	return log.String("tx_ids", strings.Join(strs, ", "))
}

// EmptyTransactionID is a canonical empty TransactionID.
var EmptyTransactionID = TransactionID{}

// TransactionAuthenticationMessage is an incomplete transaction binary representation
type TransactionAuthenticationMessage struct {
	NetID           NetworkID
	TxType          TransactionType
	TransactionData []byte
}

// Type returns transaction type
func (txm TransactionAuthenticationMessage) Type() TransactionType {
	return txm.TxType
}

// Sign signs transaction binary data
func (txm TransactionAuthenticationMessage) Sign(signer *signing.EdSigner) (_ SignedTransaction, err error) {
	bf := bytes.Buffer{}
	if _, err = xdr.Marshal(&bf, &txm); err != nil {
		return
	}
	signature := TxSignatureFromBytes(signer.Sign(bf.Bytes()))
	return txm.Encode(TxPublicKeyFromBytes(signer.PublicKey().Bytes()), signature)
}

// Encode encodes transaction into the independent form
func (txm TransactionAuthenticationMessage) Encode(pubKey TxPublicKey, signature TxSignature) (_ SignedTransaction, err error) {
	stl := SignedTransactionLayout{TxType: byte(txm.TxType), Data: txm.TransactionData, Signature: signature}
	if !txm.TxType.IsEdPlus() {
		stl.PubKey = pubKey.Bytes()
	}
	bf := bytes.Buffer{}
	if _, err = xdr.Marshal(&bf, &stl); err != nil {
		return
	}
	return bf.Bytes(), nil
}

// Verify verifies transaction bytes
func (txm TransactionAuthenticationMessage) Verify(pubKey TxPublicKey, sig TxSignature) bool {
	bf := bytes.Buffer{}
	if _, err := xdr.Marshal(&bf, &txm); err != nil {
		return false
	}
	return sig.Verify(pubKey.Bytes(), bf.Bytes())
}

// SignedTransactionLayout represents fields layout of a signed transaction
type SignedTransactionLayout struct {
	TxType    byte
	Signature [TxSignatureLength]byte
	Data      []byte
	PubKey    [] /*TODO:???*/ byte
}

// Type returns transaction type
func (stl SignedTransactionLayout) Type() TransactionType {
	return TransactionType(stl.TxType)
}

// SignedTransaction is the binary transaction independent form
type SignedTransaction []byte

// ID returns transaction identifier
func (stx SignedTransaction) ID() TransactionID {
	return TransactionID(CalcHash32(stx[:]))
}

// Decode decodes binary transaction into transaction object
func (stx SignedTransaction) Decode() (tx Transaction, err error) {
	stl := SignedTransactionLayout{}
	if _, err = xdr.Unmarshal(bytes.NewReader(stx[:]), &stl); err != nil {
		return
	}

	txm := TransactionAuthenticationMessage{GetNetworkID(), stl.Type(), stl.Data}
	bf := bytes.Buffer{}
	if _, err = xdr.Marshal(&bf, &txm); err != nil {
		return
	}

	pubKey := stl.PubKey
	if stl.Type().IsEdPlus() {
		if pubKey, err = ed25519.ExtractPublicKey(bf.Bytes(), stl.Signature[:]); err != nil {
			return tx, fmt.Errorf("failed to extract transaction public key: %v", err.Error())
		}
	} else {
		// TODO: signing does not have PublicKey length constant
		if len(pubKey) != ed25519.PublicKeySize || !signing.Verify(signing.NewPublicKey(pubKey), bf.Bytes(), stl.Signature[:]) {
			return tx, fmt.Errorf("failed to verify transaction signature")
		}
	}

	return stl.Type().Decode(TxPublicKeyFromBytes(pubKey), stl.Signature, stx.ID(), stl.Data)
}

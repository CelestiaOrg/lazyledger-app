package shares

import (
	"bytes"
	"encoding/binary"

	"github.com/celestiaorg/celestia-app/pkg/appconsts"
	"github.com/celestiaorg/nmt/namespace"
)

// NamespacedPaddedShare returns a share that acts as padding. Namespaced
// padding shares follow a blob so that the next blob may start at an index that
// conforms to non-interactive default rules. The ns parameter provided should
// be the namespace of the blob that precedes this padding in the data square.
func NamespacedPaddedShare(ns namespace.ID) Share {
	infoByte, err := NewInfoByte(appconsts.ShareVersionZero, true)
	if err != nil {
		panic(err)
	}

	sequenceLen := make([]byte, appconsts.SequenceLenBytes)
	binary.BigEndian.PutUint32(sequenceLen, uint32(0))

	padding := bytes.Repeat([]byte{0}, appconsts.ShareSize-appconsts.NamespaceSize-appconsts.ShareInfoBytes-appconsts.SequenceLenBytes)

	share := make([]byte, 0, appconsts.ShareSize)
	share = append(share, ns...)
	share = append(share, byte(infoByte))
	share = append(share, sequenceLen...)
	share = append(share, padding...)
	return share
}

// NamespacedPaddedShares returns n namespaced padded shares.
func NamespacedPaddedShares(ns namespace.ID, n int) []Share {
	shares := make([]Share, n)
	for i := 0; i < n; i++ {
		shares[i] = NamespacedPaddedShare(ns)
	}
	return shares
}

func IsNamespacedPadded(s Share) (bool, error) {
	isSequenceStart, err := s.IsSequenceStart()
	if err != nil {
		return false, err
	}
	sequenceLen, err := s.SequenceLen()
	if err != nil {
		return false, err
	}

	return isSequenceStart && sequenceLen == 0, nil
}

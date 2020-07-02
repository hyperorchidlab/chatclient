package chatcrypt

import (
	"bytes"
	"crypto/ed25519"
	"github.com/btcsuite/btcutil/base58"

	"errors"
)

func GenGroupAesKey2(mainPriv ed25519.PrivateKey, pubkeys []string) (aes []byte, groupKeys []string, err error) {
	var pksBytes [][]byte
	for i := 0; i < len(pubkeys); i++ {
		pksBytes = append(pksBytes, base58.Decode(pubkeys[i]))
	}

	aesk, gks, err := GenGroupAesKey(mainPriv, pksBytes)
	if err != nil {
		return nil, nil, err
	}

	for i := 0; i < len(gks); i++ {
		groupKeys = append(groupKeys, base58.Encode(gks[i]))
	}
	return aesk, groupKeys, nil
}

func GenGroupAesKey(mainPriv ed25519.PrivateKey, pubkeys [][]byte) (aes []byte, groupKeys [][]byte, err error) {
	if len(pubkeys) <= 0 {
		return
	}

	derivePub := mainPriv.Public()

	var pubkeys2 [][]byte

	for i := 0; i < len(pubkeys); i++ {
		if bytes.Compare(derivePub.(ed25519.PublicKey), pubkeys[i]) != 0 {
			pubkeys2 = append(pubkeys2,pubkeys[i])
		}
	}

	r := InsertionSortDArray(pubkeys2)
	priv := mainPriv

	groupKeys = append(groupKeys, derivePub.(ed25519.PublicKey))

	var pub ed25519.PublicKey

	for i := 0; i < len(r); i++ {
		aes, err = GenerateAesKey(r[i], priv)
		if err != nil {
			return
		}

		if i == len(r)-1 {
			break
		}

		pub, priv = DeriveKey(aes)
		groupKeys = append(groupKeys, pub)
	}

	return
}

func DeriveGroupKey(priv ed25519.PrivateKey, groupPKs [][]byte, pubkeys [][]byte) (aes []byte, err error) {
	derivePub := priv.Public().(ed25519.PublicKey)

	var pubkeys2 [][]byte

	for i := 0; i < len(pubkeys); i++ {
		if bytes.Compare(groupPKs[0], pubkeys[i]) != 0 {
			pubkeys2 = append(pubkeys2,pubkeys[i])
		}
	}

	if len(groupPKs) != len(pubkeys2) {
		return nil, errors.New("pubkeys errors")
	}

	grpidx := -1

	r := InsertionSortDArray(pubkeys2)

	for i := 0; i < len(r); i++ {
		if bytes.Compare(derivePub, r[i]) == 0 {
			grpidx = i
		}
	}

	if grpidx == -1 {
		return nil, errors.New("pubkey not found")
	}

	for i := grpidx; i < len(r); i++ {
		var pk ed25519.PublicKey
		if i == grpidx {
			pk = groupPKs[i]

		} else {
			pk = r[i]
		}
		aes, err = GenerateAesKey(pk, priv)
		if err != nil {
			return
		}

		if i == len(r)-1 {
			break
		}

		_, priv = DeriveKey(aes)

	}

	return
}

func DeriveGroupKey2(priv ed25519.PrivateKey, groupPKs []string, pubkeys []string) (aes []byte, err error) {
	var (
		gkeys [][]byte
		pkeys [][]byte
	)
	for i := 0; i < len(groupPKs); i++ {
		gkeys = append(gkeys, base58.Decode(groupPKs[i]))
	}
	for i := 0; i < len(pubkeys); i++ {
		pkeys = append(pkeys, base58.Decode(pubkeys[i]))
	}

	return DeriveGroupKey(priv, gkeys, pkeys)

}

func InsertionSortDArray(arr [][]byte) [][]byte {

	r := make([][]byte, 0)
	r = append(r, arr[0])

	for i := 1; i < len(arr); i++ {
		flag := false
		for j := 0; j < len(r); j++ {
			if bytes.Compare(r[j], arr[i]) > 0 {
				if j == 0 {
					r1 := make([][]byte, 0)
					r1 = append(r1, arr[i])
					r1 = append(r1, r...)
					r = r1
				} else {
					r1 := make([][]byte, 0)
					r2 := r[:j]
					r3 := r[j:]
					r1 = append(r1, r2...)
					r1 = append(r1, arr[i])
					r1 = append(r1, r3...)
					r = r1
				}
				flag = true
				break
			}
		}
		if !flag {
			r = append(r, arr[i])
		}

	}

	return r
}

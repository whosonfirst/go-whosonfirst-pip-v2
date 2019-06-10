package hash

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/ioutil"
)

type Hash struct {
	algo string
}

func NewWOFHash() (*Hash, error) {
     return NewHash("md5")
}

func NewHash(algo string) (*Hash, error) {

	switch algo {

	// this is really dumb - see if there is some way to let crypto.Hash
	// figure out what is availble... and import it?
	// (20170802/thisisaaronland)

	case "md5":
		// pass
	case "sha1":
		// pass
	default:
		return nil, errors.New("Unsupported hashing algorithm")
	}

	h := Hash{
		algo: algo,
	}

	return &h, nil
}

func (h *Hash) HashFile(path string) (string, error) {

	body, err := ioutil.ReadFile(path)

	if err != nil {
		return "", err
	}

	return h.HashBytes(body)
}

func (h *Hash) HashFromJSON(raw []byte) (string, error) {

	var stub interface{}

	err := json.Unmarshal(raw, &stub)

	if err != nil {
		return "", err
	}

	body, err := json.Marshal(stub)

	if err != nil {
		return "", err
	}

	return h.HashBytes(body)
}

func (h *Hash) HashString(body string) (string, error) {
	return h.HashBytes([]byte(body))
}

func (h *Hash) HashBytes(body []byte) (string, error) {

     	var str string

	switch h.algo {

	// this is still dumb - see notes above
	// (20170802/thisisaaronland)

	case "md5":
		hash := md5.Sum(body)
		str = hex.EncodeToString(hash[:])
	case "sha1":
		hash := sha1.Sum(body)
		str = hex.EncodeToString(hash[:])
	default:
		return "", errors.New("How did we even get this far")
	}

	return str, nil
}

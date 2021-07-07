package api

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"golang.org/x/xerrors"
)

var (
	ErrInvalidHMACSignature = xerrors.Errorf("invalid hmac signature")
)

type AuthHMAC struct {
	key       string
	signParam string
}

func NewAuthHMAC(key string, signParam string) *AuthHMAC {
	return &AuthHMAC{
		key:       key,
		signParam: signParam,
	}
}

func (auth *AuthHMAC) Allow(ctx context.Context, r *http.Request) error {
	qs := r.URL.Query()

	sign := qs.Get(auth.signParam)
	qs.Del(auth.signParam)

	return auth.validHMAC(qs, sign)
}

// ValidMAC reports whether messageMAC is a valid HMAC tag for message.
func (auth *AuthHMAC) validHMAC(params url.Values, signature string) error {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	msgParts := make([]string, len(keys))

	for i, k := range keys {
		msgParts[i] = k + "=" + params.Get(k)
	}

	msg := strings.Join(msgParts, "|")

	mac := hmac.New(sha256.New, []byte(auth.key))
	mac.Write([]byte(msg))
	expectedMAC := mac.Sum(nil)

	sign, err := hex.DecodeString(signature)
	if err != nil {
		return xerrors.Errorf("signature is not hex encoded")
	}

	if !hmac.Equal(sign, expectedMAC) {
		return ErrInvalidHMACSignature
	}

	return nil
}

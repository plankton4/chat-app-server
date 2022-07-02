package network

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

type AppleTokenData struct {
	UID      string
	EMail    string
	Verified bool
}

func VerifyAppleToken(fromClientToken string) (*AppleTokenData, error) {
	token, err := parseJWT(fromClientToken)

	unmarshalledToken, _ := json.MarshalIndent(token, "", "    ")
	fmt.Println("Parsed token ", string(unmarshalledToken))

	if err != nil {
		return nil, err
	}
	ok, err := verifyToken(token)
	if err != nil {
		fmt.Println("Error in verifyToken ", err)
		return nil, err
	}
	if !ok {
		return nil, errors.New("token not verified")
	}

	var verified bool
	switch x := token.Claims.(jwt.MapClaims)["email_verified"].(type) {
	case string:
		verified = x == "true"
	case bool:
		verified = x
	}

	fmt.Println("verified ", verified)

	at := AppleTokenData{
		UID:      token.Claims.(jwt.MapClaims)["sub"].(string),
		EMail:    "", //token.Claims.(jwt.MapClaims)["email"].(string),
		Verified: verified,
	}

	return &at, nil
}

func verifyToken(token *jwt.Token) (bool, error) {
	if !token.Valid {
		return false, nil
	}

	iss, ok := token.Claims.(jwt.MapClaims)["iss"].(string)
	if !ok {
		return false, errors.New("unexpected `iss`")
	}
	aud, ok := token.Claims.(jwt.MapClaims)["aud"].(string)
	if !ok {
		return false, errors.New("unexpected `aud`")
	}
	expiration, ok := token.Claims.(jwt.MapClaims)["exp"].(float64)
	if !ok {
		return false, errors.New("unexpected `exp`")
	}

	const (
		requiredISS = "https://appleid.apple.com"
		requiredAUD = "com.plankton42.chatapp"
	)

	if !strings.Contains(iss, requiredISS) {
		return false, errors.New("`iss` mismatch")
	}

	if aud != requiredAUD {
		return false, errors.New("`aud` mismatch")
	}

	fmt.Printf("Token expiration date %f", expiration)
	fmt.Println("Now ", time.Now().UTC().Unix())

	if int64(expiration) < time.Now().UTC().Unix() {
		return false, errors.New("token expired")
	}

	return true, nil
}

func parseJWT(tokenStr string) (*jwt.Token, error) {
	return jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		key, err := fetchKeyByKID(context.Background(), token.Header["kid"].(string))
		if err != nil {
			panic(err)
		}

		if token.Header["alg"] != key.ALG {
			return nil, errors.New("header and token algs mismatch")
		}

		modStr := strings.Replace(strings.Replace(key.N, "-", "+", -1), "_", "/", -1)
		modStr += strings.Repeat("=", 4-len(modStr)%4)
		mod, err := base64.StdEncoding.DecodeString(modStr)
		if err != nil {
			panic(err)
		}
		n := big.NewInt(0)
		n.SetBytes(mod)

		eDecoded, err := base64.StdEncoding.DecodeString(key.E)
		if err != nil {
			panic(err)
		}

		var expBytes []byte
		if len(eDecoded) < 8 {
			expBytes = make([]byte, 8-len(eDecoded), 8)
			expBytes = append(expBytes, eDecoded...)
		} else {
			expBytes = eDecoded
		}

		e := binary.BigEndian.Uint64(expBytes)
		return &rsa.PublicKey{N: n, E: int(e)}, nil
	})
}

type Key struct {
	KTY string `json:"kty"`
	KID string `json:"kid"`
	USE string `json:"use"`
	ALG string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

func fetchKeyByKID(ctx context.Context, kid string) (*Key, error) {
	const (
		appleServerURL = "https://appleid.apple.com/auth/keys"
	)

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, appleServerURL, nil)
	if err != nil {
		return nil, errors.WithMessage(err, "cannot prepare apple public keys request")
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.WithMessage(err, "cannot fetch apple public keys")
	}

	defer func() {
		if resp.Body != nil {
			resp.Body.Close()
		}
	}()

	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithMessage(err, "cannot read response body")
	}

	type Keys struct {
		Keys []Key
	}

	var keys Keys

	err = json.Unmarshal(d, &keys)
	if err != nil {
		return nil, errors.WithMessage(err, "cannot unmarshal keys")
	}

	for _, k := range keys.Keys {
		if k.KID == kid {
			return &k, nil
		}
	}

	return nil, errors.New("apple public key not found")
}

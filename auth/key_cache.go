package auth

import (
	"os"

	"github.com/golang-jwt/jwt"
	"goyave.dev/goyave/v4/config"
)

var (
	keyCache = map[string]interface{}{}
)

func loadKey(cfg string) (interface{}, error) {
	if k, ok := keyCache[cfg]; ok {
		return k, nil
	}

	data, err := os.ReadFile(config.GetString(cfg))
	if err != nil {
		return nil, err
	}

	var key interface{}
	switch cfg {
	case "auth.jwt.rsa.private":
		if config.Has("auth.jwt.rsa.password") {
			key, err = jwt.ParseRSAPrivateKeyFromPEMWithPassword(data, config.GetString("auth.jwt.rsa.password"))
		} else {
			key, err = jwt.ParseRSAPrivateKeyFromPEM(data)
		}
	case "auth.jwt.rsa.public":
		key, err = jwt.ParseRSAPublicKeyFromPEM(data)
	case "auth.jwt.ecdsa.private":
		key, err = jwt.ParseECPrivateKeyFromPEM(data)
	case "auth.jwt.ecdsa.public":
		key, err = jwt.ParseECPublicKeyFromPEM(data)
	}

	if err == nil {
		keyCache[cfg] = key
	}
	return key, err
}

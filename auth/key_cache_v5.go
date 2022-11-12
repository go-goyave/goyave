package auth

import (
	"errors"
	"os"

	"github.com/golang-jwt/jwt"
	"goyave.dev/goyave/v4/config"
)

type keyCacheV5 map[string]any

func (c keyCacheV5) loadKey(cfg *config.Config, entry string) (any, error) {
	if k, ok := keyCache[entry]; ok {
		return k, nil
	}

	data, err := os.ReadFile(cfg.GetString(entry)) // TODO support embeds?
	if err != nil {
		return nil, err
	}

	var key interface{}
	switch entry {
	case "auth.jwt.rsa.private":
		if cfg.Has("auth.jwt.rsa.password") {
			key, err = jwt.ParseRSAPrivateKeyFromPEMWithPassword(data, cfg.GetString("auth.jwt.rsa.password"))
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
		keyCache[entry] = key
	}
	return key, err
}

func (c keyCacheV5) getPrivateKey(cfg *config.Config, signingMethod jwt.SigningMethod) (any, error) {
	switch signingMethod.(type) {
	case *jwt.SigningMethodRSA:
		return c.loadKey(cfg, "auth.jwt.rsa.private")
	case *jwt.SigningMethodECDSA:
		return c.loadKey(cfg, "auth.jwt.ecdsa.private")
	case *jwt.SigningMethodHMAC:
		return []byte(cfg.GetString("auth.jwt.secret")), nil
	default:
		return nil, errors.New("Unsupported JWT signing method: " + signingMethod.Alg())
	}
}

var globalKeyCache = keyCacheV5{}

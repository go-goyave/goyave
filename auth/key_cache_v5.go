package auth

import (
	"errors"
	"os"
	"sync"

	"github.com/golang-jwt/jwt"
	"goyave.dev/goyave/v4/config"
)

type keyCacheV5 struct { // TODO move this to a service?
	config *config.Config
	cache  sync.Map
}

func (c *keyCacheV5) loadKey(entry string) (any, error) {
	if k, ok := c.cache.Load(entry); ok {
		return k, nil
	}

	data, err := os.ReadFile(c.config.GetString(entry)) // TODO support embeds?
	if err != nil {
		return nil, err
	}

	var key interface{}
	switch entry {
	case "auth.jwt.rsa.private":
		if c.config.Has("auth.jwt.rsa.password") {
			key, err = jwt.ParseRSAPrivateKeyFromPEMWithPassword(data, c.config.GetString("auth.jwt.rsa.password"))
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
		c.cache.Store(entry, key)
	}
	return key, err
}

func (c *keyCacheV5) getPrivateKey(signingMethod jwt.SigningMethod) (any, error) {
	switch signingMethod.(type) {
	case *jwt.SigningMethodRSA:
		return c.loadKey("auth.jwt.rsa.private")
	case *jwt.SigningMethodECDSA:
		return c.loadKey("auth.jwt.ecdsa.private")
	case *jwt.SigningMethodHMAC:
		return []byte(c.config.GetString("auth.jwt.secret")), nil
	default:
		return nil, errors.New("Unsupported JWT signing method: " + signingMethod.Alg())
	}
}

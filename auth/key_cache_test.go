package auth

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"testing"

	"goyave.dev/goyave/v3"
	"goyave.dev/goyave/v3/config"
)

type KeyCacheTestSuite struct {
	goyave.TestSuite
}

func (suite *KeyCacheTestSuite) TestLoadKeyCached() {
	key, err := loadKey("auth.jwt.rsa.public")
	suite.Nil(err)
	suite.NotNil(key)
	rsaPubKey, ok := key.(*rsa.PublicKey)
	suite.NotNil(rsaPubKey)
	suite.True(ok)

	cached, ok := keyCache["auth.jwt.rsa.public"]
	suite.True(ok)
	suite.Same(rsaPubKey, cached)

	cached2, err := loadKey("auth.jwt.rsa.public")
	suite.Nil(err)
	suite.Same(rsaPubKey, cached2)
	suite.Same(cached, cached2)
}

func (suite *KeyCacheTestSuite) TestLoadKeyFileDoesntExist() {
	prev := config.GetString("auth.jwt.rsa.public")
	defer config.Set("auth.jwt.rsa.public", prev)

	config.Set("auth.jwt.rsa.public", "resource/notafile")
	key, err := loadKey("auth.jwt.rsa.public")
	suite.Nil(key)
	suite.NotNil(err)

	cached, ok := keyCache["auth.jwt.rsa.public"]
	suite.False(ok)
	suite.Nil(cached)
}

func (suite *KeyCacheTestSuite) TestLoadKeyECDSAPrivate() {
	key, err := loadKey("auth.jwt.ecdsa.private")
	suite.Nil(err)
	suite.NotNil(key)
	ecdsaKey, ok := key.(*ecdsa.PrivateKey)
	suite.NotNil(ecdsaKey)
	suite.True(ok)
}

func (suite *KeyCacheTestSuite) TestLoadKeyECDSAPublic() {
	key, err := loadKey("auth.jwt.ecdsa.public")
	suite.Nil(err)
	suite.NotNil(key)
	ecdsaPubKey, ok := key.(*ecdsa.PublicKey)
	suite.NotNil(ecdsaPubKey)
	suite.True(ok)
}

func (suite *KeyCacheTestSuite) TestLoadKeyRSAPrivate() {
	key, err := loadKey("auth.jwt.rsa.private")
	suite.Nil(err)
	suite.NotNil(key)
	rsaKey, ok := key.(*rsa.PrivateKey)
	suite.NotNil(rsaKey)
	suite.True(ok)
}

func (suite *KeyCacheTestSuite) TestLoadKeyRSAPrivateWithPassword() {
	prev := config.GetString("auth.jwt.rsa.private")
	defer config.Set("auth.jwt.rsa.private", prev)
	config.Set("auth.jwt.rsa.private", "resources/rsa/private-with-pass.pem")
	config.Set("auth.jwt.rsa.password", "rsa-password")
	defer config.Set("auth.jwt.rsa.password", nil)
	key, err := loadKey("auth.jwt.rsa.private")
	suite.Nil(err)
	suite.NotNil(key)
	rsaKey, ok := key.(*rsa.PrivateKey)
	suite.NotNil(rsaKey)
	suite.True(ok)
}

func (suite *KeyCacheTestSuite) TestLoadKeyRSAPublic() {
	key, err := loadKey("auth.jwt.rsa.public")
	suite.Nil(err)
	suite.NotNil(key)
	rsaPubKey, ok := key.(*rsa.PublicKey)
	suite.NotNil(rsaPubKey)
	suite.True(ok)
}

func (suite *KeyCacheTestSuite) TestLoadKeyUnspported() {
	suite.Panics(func() {
		loadKey("not a config entry")
	})
}

func (suite *KeyCacheTestSuite) TearDownTest() {
	keyCache = map[string]interface{}{}
}

func TestKeyCacheAuthenticatorSuite(t *testing.T) {
	goyave.RunTest(t, new(KeyCacheTestSuite))
}

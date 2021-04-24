package key

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"io/ioutil"
	"math/big"
	"os"

	"github.com/hugefiver/qush/util"

	"github.com/rs/zerolog/log"
)

//var Pri crypto.PrivateKey
//var Pub crypto.PublicKey

func CreateEd25519Key() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(nil)
}

func CreateRSAKey(bits int) (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, bits)
}

func CreateEcdsaKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
}

func MarshalPriKey(pri crypto.PrivateKey) (bytes []byte, err error) {
	switch v := pri.(type) {
	case *rsa.PrivateKey:
		cur := x509.MarshalPKCS1PrivateKey(v)
		return pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: cur,
		}), nil
	case *ecdsa.PrivateKey:
		cur, err := x509.MarshalECPrivateKey(v)
		if err == nil {
			block := &pem.Block{
				Type:  "EC PRIVATE KEY",
				Bytes: cur,
			}
			return pem.EncodeToMemory(block), nil
		}
	default:
		// For `ed25519` or else
		cur, err := x509.MarshalPKCS8PrivateKey(v)
		if err == nil {
			block := &pem.Block{
				Type:  "PRIVATE KEY",
				Bytes: cur,
			}
			return pem.EncodeToMemory(block), nil
		}
	}
	return
}

func MarshalPubKey(pub crypto.PublicKey) (bytes []byte, err error) {
	b, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: b,
	}), nil
}

func LoadHostKey(path string) (crypto.PrivateKey, error) {
	file, err := os.Open(util.GetPath(path))
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	b, _ := pem.Decode(bytes)
	if b == nil {
		return nil, errors.New("cannot parse pem file of private key")
	}
	key, err := x509.ParsePKCS8PrivateKey(b.Bytes)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func GenTlsConfig(pub []byte, pri []byte) (*tls.Config, error) {
	b, r := pem.Decode(pri)
	if b == nil {
		log.Error().Msg("Get none private key from bytes")
		return nil, errors.New("wrong private key bytes ")
	}
	if len(r) > 0 {
		log.Debug().Msgf("Rest of private key bytes: %v", r)
	}
	priKey, err := x509.ParsePKCS8PrivateKey(b.Bytes)
	if err != nil {
		log.Err(err).Msgf("cannot parse private key")
		return nil, err
	}

	b, r = pem.Decode(pub)
	if b == nil {
		log.Error().Msg("Get none public key from bytes")
		return nil, errors.New("wrong public key bytes ")
	}
	if len(r) > 0 {
		log.Debug().Msgf("Rest of public key bytes: %v", r)
	}
	pubKey, err := x509.ParsePKIXPublicKey(b.Bytes)
	if err != nil {
		log.Err(err).Msgf("Cannot parse public key")
		return nil, err
	}

	templ := x509.Certificate{SerialNumber: big.NewInt(1)}

	cert, err := x509.CreateCertificate(rand.Reader, &templ, &templ, pubKey, priKey)
	if err != nil {
		return nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert})

	tlsCert, err := tls.X509KeyPair(certPEM, pri)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	}, nil
}

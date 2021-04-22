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
	"math/big"

	"github.com/rs/zerolog/log"
)

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
		// For `ed25519`, `curve25519` or else
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
	return x509.MarshalPKIXPublicKey(pub)
}

func GenTlsConfig(pub []byte, pri []byte) (*tls.Config, error) {
	priKey, err := x509.ParsePKCS8PrivateKey(pri)
	if err != nil {
		log.Error().Msgf("Cannot parse private key")
		return nil, err
	}

	pubKey, err := x509.ParsePKIXPublicKey(pub)
	if err != nil {
		log.Error().Msgf("Cannot parse public key")
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

package util

import (
	"crypto/tls"
)

func GetTLSVersionName(v uint16) string {
	switch v {
	case tls.VersionTLS10:
		return "tls-1.0"
	case tls.VersionTLS11:
		return "tls-1.1"
	case tls.VersionTLS12:
		return "tls-1.2"
	case tls.VersionTLS13:
		return "tls-1.3"

	case tls.VersionSSL30:
		return "ssl-3.0"
	default:
		return "unknown"
	}
}

func GetTLSCipherSuitesName(v uint16) string {
	switch v {
	case tls.TLS_RSA_WITH_RC4_128_SHA:
		return "tls-rsa-with-rc4-128-sha"
	case tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA:
		return "tls-rsa-with-3des-ede-cbc-sha"
	case tls.TLS_RSA_WITH_AES_128_CBC_SHA:
		return "tls-rsa-with-aes-128-cbc-sha"
	case tls.TLS_RSA_WITH_AES_256_CBC_SHA:
		return "tls-rsa-with-aes-256-cbc-sha"
	case tls.TLS_RSA_WITH_AES_128_CBC_SHA256:
		return "tls-rsa-with-aes-128-cbc-sha256"
	case tls.TLS_RSA_WITH_AES_128_GCM_SHA256:
		return "tls-rsa-with-aes-128-gcm-sha256"
	case tls.TLS_RSA_WITH_AES_256_GCM_SHA384:
		return "tls-rsa-with-aes-256-gcm-sha384"
	case tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA:
		return "tls-ecdhe-ecdsa-with-rc4-128-sha"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA:
		return "tls-ecdhe-ecdsa-with-aes-128-sha"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA:
		return "tls-ecdhe-ecdsa-with-aes-256-cbc-sha"
	case tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA:
		return "tls-ecdhe-rsa-with-rc4-128-sha"
	case tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA:
		return "tls-ecdhe-rsa-with-3des-ede-cbc-sha"
	case tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA:
		return "tls-ecdhe-rsa-with-aes-128-cbc-sha"
	case tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA:
		return "tls-ecdhe-rsa-with-aes-256-cbc-sha"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256:
		return "tls-ecdhe-ecdsa-with-aes-128-cbc-sha256"
	case tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256:
		return "tls-ecdhe-rsa-with-aes-128-cbc-sha256"
	case tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:
		return "tls-ecdhe-rsa-with-aes-128-gcm-sha256"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256:
		return "tls-ecdhe-ecdsa-with-aes-128-gcm-sha256"
	case tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384:
		return "tls-ecdhe-rsa-with-aes-256-gcm-sha256"
	case tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384:
		return "tls-ecdhe-ecdsa-with-aes-256-gcm-sha256"
	case tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256:
		return "tls-ecdhe-rsa-with-chacha20-poly1305-sha256"
	case tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256:
		return "tls-ecdhe-ecdsa-with-chacha20-poly1305-sha256"
	default:
		return "unknown"
	}
}

package security

import "time"

// Test key pair (RSA 1024) for unit tests only. Do not use in production.
const (
	testPrivateKeyPEM = `-----BEGIN PRIVATE KEY-----
MIICdgIBADANBgkqhkiG9w0BAQEFAASCAmAwggJcAgEAAoGBALaFESlPtNpfbP8t
EuN1tar+0Hfqr5xNBYW8XJc4Fg+Sbs3KylmSC7x5wJhiVlu72H5xTAhgd/BjENgS
H9VhKI6SPOS/w31muJLvqihD6Ha1LevS92k93t1cBqxP2uccNoSCl+MF3Lc+5iqp
bC+kdqBi8yhL52V8z38McxXMxxlPAgMBAAECgYAa4Akg3h2xMe/ouwhW+dQgM5ka
rzHgf+7aPFwd4CJPdK5gGwYknj6gKAVV6tTweP5tz9z0NtAyU0P9rN2HG+FOrUGc
Z01PYDw0kGcqVL4GT5UNzAiGXVnY7mW9+1H9GOSyKE8cMr1aNLHWW235H1ujPROB
kR+YV1dlyDFp/pYxwQJBAOCIdxeO7+pVdk8XrDiu2sbKh8r539B0ZNgqH7YWU3dE
hkvtoVrp74kzidU8wZJCIjiL4g3XG6psKsMBl1AA/F8CQQDQGUx44tOxPjdMe+p1
OTWzZ90vPnfQ1s4/qljlHA6APD60RTj4bGorRVsho8Txct89skeohKgUSq5V4Ue7
iQkRAkAPDPa2rI0mbw4cJSEVN5tQofjSQUegaHzuBHzVrs9vejdqVYZwWqgE0WCW
25i6Hha/JZlEhjvDg7amFbA326kPAkEAv7Oei/pBE5WB8bZxnT1vp+71hnEghUVs
yJ+Ptreq8B0Pkpf2THvrLiN9OTcZ1WeCGd7jPm2+PLszcK/QmgU6UQJAEAyGNFKH
39EU4f+vQu+H6bllsK1lnAFWz+Je6gNSL/zAH6rkK6Pq7Yf0AAw7SVzINtjCA6n8
TSXVFvM2qUiMFA==
-----END PRIVATE KEY-----`
	testPublicKeyPEM = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC2hREpT7TaX2z/LRLjdbWq/tB3
6q+cTQWFvFyXOBYPkm7NyspZkgu8ecCYYlZbu9h+cUwIYHfwYxDYEh/VYSiOkjzk
v8N9ZriS76ooQ+h2tS3r0vdpPd7dXAasT9rnHDaEgpfjBdy3PuYqqWwvpHagYvMo
S+dlfM9/DHMVzMcZTwIDAQAB
-----END PUBLIC KEY-----`
)

// NewTestTokenProvider returns a TokenProvider using the embedded test key pair.
// For unit tests only. Callers must not use in production.
func NewTestTokenProvider() (*TokenProvider, error) {
	signer, err := ParsePrivateKey(testPrivateKeyPEM)
	if err != nil {
		return nil, err
	}
	pub, err := ParsePublicKey(testPublicKeyPEM)
	if err != nil {
		return nil, err
	}
	return NewTokenProvider(signer, pub, "test-issuer", "test-audience", 15*time.Minute, 24*time.Hour), nil
}

package auth_test

import (
	"crypto/rsa"
	"errors"
	"testing"
	"time"

	"github.com/dgraph-io/travel/business/data/auth"
	"github.com/dgraph-io/travel/foundation/tests"
	"github.com/dgrijalva/jwt-go"
)

func TestAuthenticator(t *testing.T) {
	t.Log("Given the need to be able to authenticate and authorize access.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen handling a single user.", testID)
		{
			privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateRSAKey))
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to parse the private key from pem: %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to parse the private key from pem.", tests.Success, testID)

			// The key id we are stating represents the public key in the
			// public key store.
			const keyID = "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"

			keyLookupFunc := func(publicKID string) (*rsa.PublicKey, error) {
				if publicKID != keyID {
					return nil, errors.New("no public key found")
				}
				return &privateKey.PublicKey, nil
			}
			a, err := auth.New(privateKey, keyID, "RS256", keyLookupFunc)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to create an authenticator: %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to create an authenticator.", tests.Success, testID)

			claims := auth.Claims{
				StandardClaims: jwt.StandardClaims{
					Issuer:    "travel project",
					Subject:   "0x01",
					Audience:  "students",
					ExpiresAt: time.Now().Add(8760 * time.Hour).Unix(),
					IssuedAt:  time.Now().Unix(),
				},
				Auth: auth.StandardClaims{
					Role: auth.RoleAdmin,
				},
			}

			token, err := a.GenerateToken(claims)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to generate a JWT: %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to generate a JWT.", tests.Success, testID)

			parsedClaims, err := a.ValidateToken(token)
			if err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to parse the claims: %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to parse the claims.", tests.Success, testID)

			if exp, got := claims.Auth.Role, parsedClaims.Auth.Role; exp != got {
				t.Logf("\t\tTest %d:\texp: %v", testID, exp)
				t.Logf("\t\tTest %d:\tgot: %v", testID, got)
				t.Fatalf("\t%s\tTest %d:\tShould have the expexted roles: %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould have the expexted roles.", tests.Success, testID)
		}
	}
}

// Output of:
// openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048
// ./sales-admin keygen
const privateRSAKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAnZ/BW/tuLr0uxZFw1Q5mP1JpIksU46o+kIaqIXZjSAduma18
m+oSgd1L19Fs9otAjfAlkyU8HF1hJNj/PVv8MY72vhIWv60xBB4caXuLmflAiJEt
vxHfw3WtVR9npQqEowcwrsf7MSSfdHwM4S+FbMmcl/mE9c7DUrYJBUgu1IbdI7vr
EoPE65GFafjZQHkPLUX8OaRXOt4rkT6HfYv+XqaCs6Ie+dt6xL5HiQpO90/89CAJ
hi2q8AXvhfxqCVVfLxxd3jNJVq2olkCOLJREuJ29Bb460yKOAiDigEUobUpmvT6g
gUZNrX71yP0GZxQFBhq9j1IRgPVg4CDA0Pw5FQIDAQABAoIBAQCBehtRPYXSquBi
tgfjW4Kt/ToTS22LXesquRPDjQYcws4dOp8jS/GL74Y/b+57zwNmFKAo8Oshuar0
o7N2absN0ovosd8x8EhVQ46/LxcLke1qwSa8zyfp3R5W0AdJUQyHBn7885TpV1YM
T2IdD/Yf2LTjObn4WLGlnZZnWlXtiNitjj6FRGC2kSxXMMl3ptZN7pQF+wPbAqzL
007XNYMMXNptgBnwUvbUUXyON9Ow+1hox/9crUHuHn60ITCKgRu0+OrgrqfOK6bJ
f99rR5yl5YQYRkVoFGb68Pg7eTVOU260Tl1pgl0GCLojk1O4TFYuBuLZR1dOlx9I
1b30vrj5AoGBAMLHJVOXSebm2vor76lIgJqWL5kf9e3lZ4Y6zN7rrM+lKSia6fT5
cAGfw+ce1ioyxkJZZ96bkq7EHwypC1GekntAEYixkyEW7H9H3TnPhyLn3ySHnBYb
OKIHShK3XK8kes9khNKJ7FVY1fOj5JC67wQZRRhWlEyOFxKzH9KtygyTAoGBAM8r
A5WNkWT9com4CLVuMmKrGAN+9LwHh7WA5jpqCgvNQ03kgzYH2lf75lVYhX09+bYF
BM3obKyqM8RUp1iYyQr0sr7Ca/DpaMiAKfm9aLOd90xyLTmVUI5x7rwr7UXhlmrY
4K0bdvc3T7FBOxT/bfyRR4DosEyjcTyvj9gR/1S3AoGBAIf6seNmtlA+ENggfkNn
e2jwurAjMPTxd9GtEUP7snyQaGiRpg3BamGn4QNkcs2o/uJpOmudnszl3GthRKap
lsf21Ybhub6bG2ZMjHSEnmpPCGifR+fi/ymW/y6L1mfrhtVs7pFxeo2m5E8gtzwX
VTA+WA+Cuiur8w26Adh6PZmDAoGAdFjN7IHTNAp69wlaKrq2pV89X0k/nRIFj1PS
+N9wwOwIboh1gDSs1VjtJOVQIuRZh3YOGq37yoTUCeEZEtLLpdGDSUrbYDNV27TO
3ikX0jhXGKHO8FYBJd6qmxd4bBSja2Jd3Bpel7yCjyP5UHObi4rzw1vrFz97av+W
I10ILsUCgYBmMCXEWDtM/f+Gq53yHV2XyZ7N0fDftPKFwyBgu+VvytacyMT8yFLO
8yePQjXKUmm1OE+LoVciT+dyibh0XfKmx936bK7GvHL6TKRYMfbuUqh5CQlGT2WE
khtQ09sZjN4h5zTB5TO4JIPvOHQnxhpEnrw8kXkjQx/yVCM4TEHrbw==
-----END RSA PRIVATE KEY-----`

// To generate a public key PEM file.
// openssl rsa -pubout -in private.pem -out public.pem
// ./sales-admin keygen

package commands

import (
	"context"
	"crypto/rsa"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/dgraph-io/travel/business/data"
	"github.com/dgraph-io/travel/business/data/user"
	"github.com/dgraph-io/travel/business/sys/auth"
	"github.com/dgraph-io/travel/foundation/keystore"
	"github.com/dgrijalva/jwt-go/v4"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// GenToken generates a JWT for the specified user.
func GenToken(log *log.Logger, gqlConfig data.GraphQLConfig, email string, privateKeyFile string, algorithm string) error {
	if email == "" || privateKeyFile == "" || algorithm == "" {
		fmt.Println("help: gentoken <email> <private_key_file> <algorithm>")
		fmt.Println("algorithm: RS256, HS256")
		return ErrHelp
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	store := user.NewStore(
		log,
		data.NewGraphQL(gqlConfig),
	)
	traceID := uuid.New().String()

	// Retrieve the user by email so we have the roles for this user.
	usr, err := store.QueryByEmail(ctx, traceID, email)
	if err != nil {
		return errors.Wrap(err, "getting user")
	}

	// limit PEM file size to 1 megabyte. This should be reasonable for
	// almost any PEM file and prevents shenanegans like linking the file
	// to /dev/random or something like that.
	pkf, err := os.Open(privateKeyFile)
	if err != nil {
		return errors.Wrap(err, "opening PEM private key file")
	}
	defer pkf.Close()
	privatePEM, err := io.ReadAll(io.LimitReader(pkf, 1024*1024))
	if err != nil {
		return errors.Wrap(err, "reading PEM private key file")
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privatePEM)
	if err != nil {
		return errors.Wrap(err, "parsing PEM into private key")
	}

	// In a production system, a key id (KID) is used to retrieve the correct
	// public key to parse a JWT for auth and claims.
	const keyID = "54bb2165-71e1-41a6-af3e-7da4a0e1e2c1"

	// An authenticator maintains the state required to handle JWT processing.
	// It requires a keystore to lookup private and public keys based on a
	// key id. There is a keystore implementation in the project.
	a, err := auth.New("RS256", keystore.NewMap(map[string]*rsa.PrivateKey{keyID: privateKey}))
	if err != nil {
		return errors.Wrap(err, "constructing authenticator")
	}

	// Generating a token requires defining a set of claims. In this applications
	// case, we only care about defining the subject and the user in question and
	// the roles they have on the database. This token will expire in a year.
	//
	// iss (issuer): Issuer of the JWT
	// sub (subject): Subject of the JWT (the user)
	// aud (audience): Recipient for which the JWT is intended
	// exp (expiration time): Time after which the JWT expires
	// nbf (not before time): Time before which the JWT must not be accepted for processing
	// iat (issued at time): Time at which the JWT was issued; can be used to determine age of the JWT
	// jti (JWT ID): Unique identifier; can be used to prevent the JWT from being replayed (allows a token to be used only once)
	claims := auth.Claims{
		StandardClaims: jwt.StandardClaims{
			Issuer:    "travel project",
			Subject:   usr.ID,
			ExpiresAt: jwt.At(time.Now().Add(8760 * time.Hour)),
			IssuedAt:  jwt.Now(),
		},
		Auth: auth.StandardClaims{
			Role: usr.Role,
		},
	}

	// This will generate a JWT with the claims embedded in them. The database
	// with need to be configured with the information found in the public key
	// file to validate these claims. Dgraph does not support key rotate at
	// this time.
	token, err := a.GenerateToken(keyID, claims)
	if err != nil {
		return errors.Wrap(err, "generating token")
	}

	fmt.Printf("-----BEGIN TOKEN-----\n%s\n-----END TOKEN-----\n", token)
	return nil
}

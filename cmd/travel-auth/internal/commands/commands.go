package commands

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/dgraph-io/travel/internal/data"
	"github.com/dgraph-io/travel/internal/platform/auth"
	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

// AddUser handles the creation of users.
func AddUser(dgraph data.Dgraph, newUser data.NewUser) error {
	if newUser.Name == "" ||
		newUser.Email == "" ||
		newUser.Password == "" ||
		newUser.Roles == nil {
		return errors.New("adduser command requires an name, email, password and role")
	}

	fmt.Printf("Admin user will be created with email %q and password %q\n", newUser.Email, newUser.Password)
	fmt.Print("Continue? (1/0) ")

	var confirm bool
	if _, err := fmt.Scanf("%t\n", &confirm); err != nil {
		return errors.Wrap(err, "processing response")
	}

	if !confirm {
		fmt.Println("Canceling")
		return nil
	}

	db, err := data.NewDB(dgraph)
	if err != nil {
		return errors.Wrap(err, "init database")
	}

	ctx := context.Background()

	u, err := db.Mutate.AddUser(ctx, newUser, time.Now())
	if err != nil {
		return errors.Wrap(err, "adding user")
	}

	fmt.Println("User created with id:", u.ID)
	return nil
}

// GetUser returns information about a user by email.
func GetUser(dgraph data.Dgraph, email string) error {
	if email == "" {
		return errors.New("getuser command requires an email")
	}

	db, err := data.NewDB(dgraph)
	if err != nil {
		log.Printf("feed: Work: New Data: ERROR: %v", err)
		return errors.Wrap(err, "init database")
	}

	ctx := context.Background()

	u, err := db.Query.UserByEmail(ctx, email)
	if err != nil {
		return errors.Wrap(err, "getting user")
	}

	fmt.Printf("User: %#v\n", u)
	return nil
}

// GenerateKeys creates an x509 private key for signing auth tokens.
func GenerateKeys() error {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return errors.Wrap(err, "generating keys")
	}

	privateFile, err := os.Create("private.pem")
	if err != nil {
		return errors.Wrap(err, "creating private file")
	}
	defer privateFile.Close()

	privateBlock := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	if err := pem.Encode(privateFile, &privateBlock); err != nil {
		return errors.Wrap(err, "encoding to private file")
	}

	publicFile, err := os.Create("public.pem")
	if err != nil {
		return errors.Wrap(err, "creating public file")
	}
	defer privateFile.Close()

	publicBlock := pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: x509.MarshalPKCS1PublicKey(&key.PublicKey),
	}

	if err := pem.Encode(publicFile, &publicBlock); err != nil {
		return errors.Wrap(err, "encoding to public file")
	}

	return nil
}

// GenerateToken generates a JWT for the specified user.
func GenerateToken(dgraph data.Dgraph, email string, privateKeyFile string) error {
	if email == "" {
		return errors.New("gentoken command requires an email")
	}

	db, err := data.NewDB(dgraph)
	if err != nil {
		log.Printf("feed: Work: New Data: ERROR: %v", err)
		return errors.Wrap(err, "init database")
	}

	ctx := context.Background()

	user, err := db.Query.UserByEmail(ctx, email)
	if err != nil {
		return errors.Wrap(err, "getting user")
	}

	keyContents, err := ioutil.ReadFile(privateKeyFile)
	if err != nil {
		return errors.Wrap(err, "reading auth private key")
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyContents)
	if err != nil {
		return errors.Wrap(err, "parsing auth private key")
	}

	keyID := "1234"
	f := auth.NewSimpleKeyLookupFunc(keyID, privateKey.Public().(*rsa.PublicKey))
	authenticator, err := auth.NewAuthenticator(privateKey, keyID, "RS256", f)
	if err != nil {
		return errors.Wrap(err, "constructing authenticator")
	}

	claims := auth.Claims{
		Roles: user.Roles,
	}
	token, err := authenticator.GenerateToken(claims)
	if err != nil {
		return errors.Wrap(err, "generating token")
	}

	fmt.Printf("-----BEGIN TOKEN-----\n%s\n-----BEGIN TOKEN-----\n", token)
	return nil
}

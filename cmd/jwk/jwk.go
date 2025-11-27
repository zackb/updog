package main

import (
	crand "crypto/rand"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/big"
	"time"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/zackb/updog/user"
)

func main() {
	c := flag.Bool("c", false, "Create a new JWK signing key")
	clientId := flag.String("api_key", "", "Create a new API Key (ClientId/ClientSecret)")
	flag.Parse()

	if *c {
		printNewSymmetricKey()
	} else if *clientId != "" {
		printNewApiKey(*clientId)
	}
}

func printNewApiKey(clientId string) {
	secret := randomString(32)
	encryptedSecret, err := user.HashPassword(secret)
	if err != nil {
		panic(err)
	}
	fmt.Printf("ClientID: %s\n", clientId)
	fmt.Printf("ClientSecret: %s\n", secret)
	fmt.Printf("Encrypted ClientSecret: %s\n", encryptedSecret)
}

func createKeyId() string {
	t := time.Now()
	return t.Format("2006-01-02-") + fmt.Sprintf("%d", t.Unix())
}

func printNewSymmetricKey() {
	key, err := createSymmetricKey(createKeyId())
	if err != nil {
		log.Fatalln(err)
	}
	j, err := json.MarshalIndent(key, "", "   ")
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(string(j))
}

func createSymmetricKey(id string) (jwk.Key, error) {
	keyData := randomString(312)
	key, err := jwk.New([]byte(keyData))
	if err != nil {
		return nil, err
	}
	err = key.Set(jwk.KeyIDKey, id)
	if err != nil {
		return nil, err
	}

	err = key.Set(jwk.AlgorithmKey, jwa.HS256)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		ri, err := crand.Int(crand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			log.Fatalln(err)
		}
		s[i] = letters[ri.Int64()]
	}
	return string(s)
}

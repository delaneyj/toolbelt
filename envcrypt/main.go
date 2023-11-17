package main

import (
	"context"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/dustin/go-humanize"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	ctx := context.Background()
	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

var CLI struct {
	Encrypt EncryptCmd `cmd:"" help:"Encrypt environment variables locally"`
	Decrypt DecryptCmd `cmd:"" help:"Decrypt environment variables locally"`
}

func run(ctx context.Context) error {
	godotenv.Load()
	cliCtx := kong.Parse(&CLI, kong.Bind(ctx))
	if err := cliCtx.Run(ctx); err != nil {
		return fmt.Errorf("failed to run cli: %w", err)
	}

	return nil
}

func parse(password, salt, extension string) (aead cipher.AEAD, filepaths []string, err error) {

	key := argon2.Key([]byte(password), []byte(salt), 3, 64*1024, 4, 32)
	aead, err = chacha20poly1305.New(key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create chacha20poly1305: %w", err)
	}

	if err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) != extension {
			return nil
		}

		filepaths = append(filepaths, path)
		return nil
	}); err != nil {
		return nil, nil, fmt.Errorf("failed to read env files: %w", err)
	}

	return
}

type EncryptCmd struct {
	Password string `short:"p" env:"ENVCRYPT_PASSWORD" help:"Secret to encrypt"`
	Salt     string `short:"s" env:"ENVCRYPT_SALT" help:"Salt to use for encryption"`
}

func (cmd *EncryptCmd) Run() error {
	aead, envFilepaths, err := parse(cmd.Password, cmd.Salt, ".env")
	if err != nil {
		return fmt.Errorf("failed to parse: %w", err)
	}

	for _, envFilepath := range envFilepaths {
		msg, err := os.ReadFile(envFilepath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", envFilepath, err)
		}

		nonce := make([]byte, aead.NonceSize(), aead.NonceSize()+len(msg)+aead.Overhead())
		if _, err := rand.Read(nonce); err != nil {
			panic(err)
		}

		// Encrypt the message and append the ciphertext to the nonce.
		encryptedMsg := aead.Seal(nonce, nonce, msg, nil)
		based := base32.StdEncoding.EncodeToString(encryptedMsg)

		encryptedFilename := fmt.Sprintf("%scrypt", envFilepath)

		fullpath := filepath.Join(filepath.Dir(envFilepath), encryptedFilename)
		if err := os.WriteFile(fullpath, []byte(based), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", fullpath, err)
		}

		log.Printf("wrote %s to %s, size: %s", envFilepath, fullpath, humanize.Bytes(uint64(len(based))))
	}

	return nil
}

type DecryptCmd struct {
	Password string `short:"p" env:"ENVCRYPT_PASSWORD" help:"Secret to encrypt"`
	Salt     string `short:"s" env:"ENVCRYPT_SALT" help:"Salt to use for encryption"`
}

func (cmd *DecryptCmd) Run() error {
	aead, envcryptFilepaths, err := parse(cmd.Password, cmd.Salt, ".envcrypt")
	if err != nil {
		return fmt.Errorf("failed to parse: %w", err)
	}

	for _, envcryptFilepath := range envcryptFilepaths {

		based, err := os.ReadFile(envcryptFilepath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", envcryptFilepath, err)
		}

		encryptedMsg, err := base32.StdEncoding.DecodeString(string(based))
		if err != nil {
			return fmt.Errorf("failed to decode %s: %w", envcryptFilepath, err)
		}

		if len(encryptedMsg) < aead.NonceSize() {
			panic("ciphertext too short")
		}

		// Split nonce and ciphertext.
		nonce, ciphertext := encryptedMsg[:aead.NonceSize()], encryptedMsg[aead.NonceSize():]

		// Decrypt the message and check it wasn't tampered with.
		plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			panic(err)
		}

		envFilepath := envcryptFilepath[:len(envcryptFilepath)-len("crypt")]
		if err := os.WriteFile(envFilepath, plaintext, 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", envFilepath, err)
		}

		log.Printf("wrote %s to %s, size: %s", envcryptFilepath, envFilepath, humanize.Bytes(uint64(len(plaintext))))
	}
	return nil
}

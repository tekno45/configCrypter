package cmd

import (
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/kms"

	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
)

type mockKMS struct {
	kmsiface.KMSAPI
	testText []byte
}

func (m mockKMS) Decrypt(input *kms.DecryptInput) (output *kms.DecryptOutput, err error) {
	output = &kms.DecryptOutput{
		Plaintext: m.testText,
	}
	return output, nil
}

func (m mockKMS) Encrypt(input *kms.EncryptInput) (*kms.EncryptOutput, error) {

	return &kms.EncryptOutput{
		CiphertextBlob: m.testText,
	}, nil
}

func TestDecryptFile(t *testing.T) {

	testData := "ThisIsTest Data"
	testReader := strings.NewReader(testData)
	response := decryptData(testReader, mockKMS{testText: []byte(testData)})
	if string(response) != testData {
		t.Fail()
	}
	return

}

func TestEncrypyFile(t *testing.T) {

	testBytes := strings.NewReader("{'Thingy':'stuff'}")
	testPipe := make(chan []byte, 100) // not sure why i need a buffered channel during testing
	kmsID := "111111111"
	encryptFile(testBytes, &kmsID, mockKMS{testText: []byte("{'Thingy':'stuff'}")}, testPipe)

	select {
	case h := <-testPipe:
		if string(h) != "{'Thingy':'stuff'}" {
			t.Error("Failed")
		}
	default:
	}

	return
}

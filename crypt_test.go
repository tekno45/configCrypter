package main

import (
	"bytes"
	"testing"

	"github.com/aws/aws-sdk-go/service/kms"

	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
)

type mockKMS struct {
	kmsiface.KMSAPI
}

func (m mockKMS) Encrypt(*kms.EncryptInput) (*kms.EncryptOutput, error) {
	return &kms.EncryptOutput{
		CiphertextBlob: []byte("1111"),
	}, nil
}

func TestEncrypyFile(t *testing.T) {
	var testBytes bytes.Buffer
	testBytes.Write([]byte("{'Thingy':'stuff'}"))
	testPipe := make(chan []byte, 100) // not sure why i need a buffered channel during testing
	kmsID := "111111111"
	encryptFile(testBytes, &kmsID, mockKMS{}, testPipe)

	select {
	case h := <-testPipe:
		if string(h) != "1111" {
			t.Error("Failed")
		}
	default:
	}

	return
}

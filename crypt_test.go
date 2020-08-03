package main

import (
	"bytes"
	"io/ioutil"
	"os"
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
	var testPath string
	var testData []byte
	var testPerms int

	testPath = "./testFile.tmp"
	testData = []byte("ThisIsTest Data")
	testPerms = 0666
	ioutil.WriteFile(testPath, testData, os.FileMode(testPerms))
	if decryptFile(&testPath, mockKMS{testText: testData}) != string(testData) {
		t.Fail()
	}
	os.Remove(testPath)

}

func TestEncrypyFile(t *testing.T) {
	var testBytes bytes.Buffer
	testBytes.Write([]byte("{'Thingy':'stuff'}"))

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

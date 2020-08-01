package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"

	"github.com/aws/aws-sdk-go/service/kms"
)

func decryptFile(targetFile *string, client kms.KMS) {
	text, err := ioutil.ReadFile(*targetFile)
	if err != nil {
		log.Fatal("can't read secret file")
	}
	payload := &kms.DecryptInput{
		CiphertextBlob: text,
	}

	response, err := client.Decrypt(payload)
	fmt.Println("Decrypting")
	fmt.Println(string(response.Plaintext))

}

//encryptFile takes in a file path and returns the KMS encrypted data
func encryptFile(targetFile *string, kmsID *string, client kms.KMS, pipe chan<- []byte) {
	text, err := ioutil.ReadFile(*targetFile)
	if err != nil {
		log.Fatal("Cannot read file: ", *targetFile, "\n", err)
	}

	var input kms.EncryptInput
	input.KeyId = kmsID
	input.Plaintext = text

	output, err := client.Encrypt(&kms.EncryptInput{
		KeyId:     aws.String(*kmsID),
		Plaintext: text})
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(text))
	fmt.Println(output.String())
	fmt.Println(string(output.CiphertextBlob))
	pipe <- output.CiphertextBlob
}

//writeEncryptedFile writes the encrypted data to disk and creates the folder to hold them
func writeEncryptedFile(outputFolder *string, osPerms *int, path *string, pipe chan []byte) string {
	file := <-pipe
	perms := os.FileMode(*osPerms)
	filename := *outputFolder + filepath.Base(*path)

	if err := ioutil.WriteFile(filename, file, perms); err != nil {
		if err := os.Mkdir(filepath.Base(*outputFolder), perms); err != nil {
			log.Fatal(err)
		}
		err := ioutil.WriteFile(filename, file, perms)
		if err != nil {
			log.Fatal(err)
		}
	}
	return filename
}

func main() {
	outputFolder := flag.String("output", "./encrypted/", "folder to output encrytped files to")
	kmsID := flag.String("kms", "", "KMS Key to use to encrypt the file")
	region := flag.String("region", "us-west-1", "region with KMS key")
	flag.Parse()
	config_file, err := ioutil.ReadFile("file_list.txt")
	if err != nil {
		log.Fatal(err)
	}
	files := strings.Split(string(config_file), "\n")
	if len(files) < 1 {
		log.Fatal("usage: ./cfgcrpyt -o=./encryptedOutPut/ /path1/file1 /path2/file2")
	}
	sess := session.Must(session.NewSession())

	client := *kms.New(sess, aws.NewConfig().WithRegion(*region))
	osPerms := int(0667)
	pipe := make(chan []byte)

	for x := range files {
		fmt.Println("Encrypting: ", files[x])
		file := strings.TrimSpace(files[x])
		if file != "" {
			go encryptFile(&file, kmsID, client, pipe)
			fmt.Println(writeEncryptedFile(outputFolder, &osPerms, &files[x], pipe))
		}
	}
}

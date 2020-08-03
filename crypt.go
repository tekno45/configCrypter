package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3/s3iface"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
)

//decryptFile takes KMS encrypted file requests AWS KMS to decrypt it
func decryptFile(targetFile *string, kmsClient kmsiface.KMSAPI) {
	text, err := ioutil.ReadFile(*targetFile)
	if err != nil {
		log.Fatal("can't read secret file")
	}
	payload := &kms.DecryptInput{
		CiphertextBlob: text,
	}

	response, err := kmsClient.Decrypt(payload)
	fmt.Println("Decrypting")
	fmt.Println(string(response.Plaintext))

}

//encryptFile takes in a file path and puts the KMS encrypted data to a channel
func encryptFile(buf bytes.Buffer, kmsID *string, kmsClient kmsiface.KMSAPI, pipeInput chan<- []byte) {
	text := buf.Bytes()
	var input kms.EncryptInput
	input.KeyId = kmsID
	input.Plaintext = text
	output, err := kmsClient.Encrypt(&kms.EncryptInput{
		KeyId:     aws.String(*kmsID),
		Plaintext: text})

	if err != nil {
		fmt.Println(err)
	}
	pipeInput <- output.CiphertextBlob

}

//writeEncryptedFile writes the encrypted data to disk and creates the folder to hold them
func writeEncryptedFile(outputFolder *string, osPerms *int, path *string, pipeInput chan []byte) string {
	file := <-pipeInput
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
	configFile, err := ioutil.ReadFile("file_list.txt")
	if err != nil {
		log.Fatal(err)
	}
	files := strings.Split(string(configFile), "\n")
	if len(files) < 1 {
		log.Fatal("usage: ./cfgcrpyt -o=./encryptedOutPut/ /path1/file1 /path2/file2")
	}
	sess := session.Must(session.NewSession())

	kmsClient := kms.New(sess, aws.NewConfig().WithRegion(*region))
	//s3Client := s3.New(sess, aws.NewConfig().WithRegion(*region))
	osPerms := int(0667)
	pipeInput := make(chan []byte)
	//pipeOutput := make(chan []byte)

	for x := range files {
		file := strings.TrimSpace(files[x])
		if file != "" {
			fmt.Println("Encrypting: ", files[x])
			text, err := ioutil.ReadFile(file)
			if err != nil {
				log.Fatal(err)
			}

			var buf bytes.Buffer
			buf.Write(text)
			//uri, err := url.Parse(file)
			if err != nil {
				go encryptFile(buf, kmsID, kmsClient, pipeInput)
				writeEncryptedFile(outputFolder, &osPerms, &files[x], pipeInput)
			}
			go encryptFile(buf, kmsID, kmsClient, pipeInput)
			writeEncryptedFile(outputFolder, &osPerms, &files[x], pipeInput)
			// Block for checking URL scheme for s3 upload
			//	switch uri.Scheme {
			//	case "s3":
			//		return
			//	default:
			//		return
			//	}

		}

	}
}

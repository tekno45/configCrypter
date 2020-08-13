/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/spf13/cobra"
)

var kmsKey string

//encryptFile takes in a file path and puts the KMS encrypted data to a channel
func encryptFile(data io.Reader, kmsID *string, kmsClient kmsiface.KMSAPI, pipeInput chan<- []byte) {
	text, err := ioutil.ReadAll(data)
	var input kms.EncryptInput
	input.KeyId = kmsID
	input.Plaintext = text
	output, err := kmsClient.Encrypt(&kms.EncryptInput{
		KeyId:     aws.String(*kmsID),
		Plaintext: text})

	if err != nil {
		log.Fatal(err)
	}
	pipeInput <- output.CiphertextBlob

}

func readFileList(fileList []byte) (list []os.File, files []string) {
	files = strings.Split(string(fileList), "\n")
	for x := range files {
		file := strings.TrimSpace(files[x])
		if file != "" {
			text, err := os.Open(file)
			if err != nil {
				fmt.Println(err)
				continue
			}
			list = append(list, *text)

		}
	}
	return list, files
}

//writeEncryptedFile writes the encrypted data to disk and creates the folder to hold them
func writeEncryptedFile(outputFolder *string, osPerms *int, wg *sync.WaitGroup, path *string, pipeInput chan []byte) (filename string) {
	defer wg.Done()
	file := <-pipeInput
	perms := os.FileMode(*osPerms)
	filename = filepath.Join(*outputFolder, filepath.Base(*path))
	if _, err := os.Stat(filepath.Base(*outputFolder)); os.IsNotExist(err) {
		os.Mkdir(filepath.Base(*outputFolder), perms)
	}
	err := ioutil.WriteFile(filename, file, perms)
	if err != nil {
		log.Fatal(err)
	}
	return
}

func findConfig(configFile string, startingDir string) ([]byte, string, error) {

	file, err := ioutil.ReadFile(filepath.Join(startingDir, configFile))
	if err != nil {
		if startingDir == "/" {
			log.Fatal(err)
		}
		return findConfig(configFile, filepath.Dir(startingDir))
	}
	return file, startingDir, nil
}

// encryptCmd represents the encrypt command
var encryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypt files mode",
	Long: `Used to encrypt files or 

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		kmsID := &kmsKey
		region := flag.String("region", "us-west-1", "region with KMS key")
		cwd, _ := os.Getwd()
		targetListPath := flag.String("f", "file_list.txt", "list of files to encrypt")

		configFile, path, err := findConfig(fileList, cwd)
		os.Chdir(path)
		outputFolder := flag.String("output", "encrypted/", "folder to output encrytped files to")
		flag.Parse()
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
		pipeInput := make(chan []byte, 100)
		//pipeOutput := make(chan []byte)
		var wg sync.WaitGroup
		fileListData, _ := ioutil.ReadFile(*targetListPath)
		fileList, fileListPaths := readFileList(fileListData)
		for x := range fileList {
			var reader io.Reader = (&fileList[x])
			//uri, err := url.Parse(file) // check file name for URI scheme, assume no scheme = file path
			//if err != nil {
			//	go encryptFile(buf, kmsID, kmsClient, pipeInput)
			//	writeEncryptedFile(outputFolder, &osPerms, &files[x], pipeInput)	}
			wg.Add(1)
			go encryptFile(reader, kmsID, kmsClient, pipeInput)
			go writeEncryptedFile(outputFolder, &osPerms, &wg, &fileListPaths[x], pipeInput)

			// Block for checking URL scheme for s3 upload
			//	switch uri.Scheme {
			//	case "s3":
			//		return
			//	default:
			//		return	}

		}
		wg.Wait()
	},
}

func init() {
	rootCmd.AddCommand(encryptCmd)
	encryptCmd.Flags().StringVar(&kmsKey, "kms", "", "KMS key for encrypting files")
	encryptCmd.MarkFlagRequired("kms")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// encryptCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// encryptCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

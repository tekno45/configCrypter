package cmd

/*
Copyright © 2020 NAME HERE josephgreene78@gmail.com

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
import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/spf13/cobra"
)

//decryptData takes in a reader and sends to KMS for decryption, returning byte slice
func decryptData(targetFile io.Reader, kmsClient kmsiface.KMSAPI) (output []byte) {
	text, err := ioutil.ReadAll(targetFile)
	if err != nil {
		log.Fatal("can't read secret data")
	}
	payload := &kms.DecryptInput{
		CiphertextBlob: text,
	}

	response, err := kmsClient.Decrypt(payload)
	if err != nil {
		fmt.Println(err)
	}
	output = response.Plaintext
	return

}

// decryptCmd represents the decrypt command
var decryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		//Everything needed to run on CLI args declared here
		sess := session.Must(session.NewSession())
		region, _ := cmd.Flags().GetString("region") // can't fail since default is set
		kmsClient := kms.New(sess, aws.NewConfig().WithRegion(region))
		cwd, _ := os.Getwd()

		//Run if  there are CLI args
		count := 0
		for x := range args {
			file, err := os.Open(args[x])
			defer file.Close()
			if err != nil {
				fmt.Println(err)
				continue
			}
			output := decryptData(file, kmsClient)
			fmt.Println(string(output))
			count++
		}

		// Exit out if CLI args were provided
		if count > 0 {
			return
		}

		//Read in lines from file
		list, _ := cmd.Flags().GetString("fileList")

		//Change to Dir config is found
		fileList, path, _ := findConfig(list, cwd)
		os.Chdir(path)

		//OpenFiles
		_, filePaths := readFileList(fileList)
		for x := range filePaths {
			filename := filepath.Join("encrypted/", filepath.Base(filePaths[x]))
			file, err := os.Open(filename)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(filename)
			output := decryptData(file, kmsClient)
			fmt.Println("output", string(output))
		}

	},
}

func init() {
	rootCmd.AddCommand(decryptCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// decryptCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// decryptCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

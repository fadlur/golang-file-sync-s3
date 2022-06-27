package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main()  {
	var listDirectory [4]string
	var listBucket []string

	welcomeText := `
	Sync file to AWS S3
	############## SETUP ##############
	1. Setup aws config using AWS CLI
	2. Load S3 Bucket
	3. Choose S3 Bucket
	4. Copy paste main directory
	5. Copy paste success directory (file which is success to upload to s3)
	6. Copy paste error directory (file which is failed to upload to s3)
	`
	fmt.Println(welcomeText)
	fmt.Println("Loading bucket s3")
	sess, err := session.NewSession(&aws.Config{})

	if err != nil {
		exitErrorf("Unable to find config, %v", err)
	}
	// Create S3 Service client
	svc := s3.New(sess)

	result, err := svc.ListBuckets(nil)
	if err != nil {
		exitErrorf("Unable to list buckets, %v", err)
	}

	fmt.Println("Buckets: ")
	for index, b := range result.Buckets {
		fmt.Printf("* %s created on %s\n", aws.StringValue(b.Name), aws.TimeValue(b.CreationDate))
		listBucket[index] = aws.StringValue(b.Name)
	}
	// get main directory
	for i := 0; i < len(listDirectory) - 1; i++ {
		switch i {
		case 0:
			fmt.Println("Please enter main directory: ")
		case 1:
			fmt.Println("Please enter success directory: ")
		case 2:
			fmt.Println("Please enter error directory: ")
		}
		reader := bufio.NewReader(os.Stdin)
		directory, err := reader.ReadString('\n')

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return
		}
		listDirectory[i] = directory
	}
}

func exitErrorf(msg string, args ...interface{})  {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}


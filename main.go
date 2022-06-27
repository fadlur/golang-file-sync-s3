package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

func main()  {
	var listConfig [4]string

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
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-southeast-1"),
	})

	if err != nil {
		exitErrorf("Unable to find config, %v", err)
	}
	// Create S3 Service client
	svc := s3.New(sess)

	result, err := svc.ListBuckets(nil)
	if err != nil {
		exitErrorf("Unable to list buckets, %v", err)
	}

	listBucketS3 := make([]string, len(result.Buckets))
	indexBucket := 0
	for _, b := range result.Buckets {
		// fmt.Printf("* %s created on %s\n", aws.StringValue(b.Name), aws.TimeValue(b.CreationDate))
		listBucketS3[indexBucket] = aws.StringValue(b.Name)
		indexBucket++
	}
	// get main directory
	for i := 0; i < len(listConfig) - 1; i++ {
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
		listConfig[i] = strings.TrimRight(directory, "\r\n")
	}
	// choose bucket name
	fmt.Println("Please enter number of bucket name: ")
	for index, name := range listBucketS3 {
		fmt.Printf("%d. %s \n", index, name)
	}
	reader := bufio.NewReader(os.Stdin)
	numberString, err := reader.ReadString('\n')

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	number, err := strconv.Atoi(strings.TrimRight(numberString, "\r\n"))
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	if (number - 1) > len(listBucketS3) {
		fmt.Println("Your choice out of range")
		os.Exit(1)
	}
	listConfig[3] = listBucketS3[number]

	for _, a := range listConfig {
		fmt.Println(a)
	}
}

func exitErrorf(msg string, args ...interface{})  {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}


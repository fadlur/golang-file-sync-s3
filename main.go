package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/fsnotify/fsnotify"
)

var (
	s3session *s3.S3
	listConfig [4]string
	pathSeparator string
)

func init()  {
	fmt.Println("Enter bucket region:")
	readerRegion := bufio.NewReader(os.Stdin)
	regionName, err := readerRegion.ReadString('\n')

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	s3session = s3.New(session.Must(session.NewSession(&aws.Config{
		Region: aws.String(strings.TrimRight(regionName, "\r\n")),
	})))

	if runtime.GOOS == "windows" {
		pathSeparator = "\\"
	} else {
		pathSeparator = "/"
	}
}

func main()  {
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
	

	// sess, err := session.NewSession(&aws.Config{
	// 	Region: aws.String(strings.TrimRight(regionName, "\r\n")),
	// })

	// if err != nil {
	// 	exitErrorf("Unable to find config, %v", err)
	// }
	// Create S3 Service client
	// svc := s3.New(s3session)

	result, err := s3session.ListBuckets(nil)
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

	if listConfig[0] != "" {
		watcherFile(listConfig[0], listConfig[1], listConfig[2], listConfig[3])
	}
}

func watcherFile(mainDirectory, successDirectory, errorDirectory, bucketName string)  {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	defer watcher.Close()

	done := make(chan bool)

	go handleFileWatcher(*watcher, successDirectory, errorDirectory, bucketName)

	err = watcher.Add(mainDirectory)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func handleFileWatcher(watcher fsnotify.Watcher, successDirectory, errorDirectory, bucketName string) {
	for {
		select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				log.Println("event: ", event)
				if (event.Op.String() == "CREATE") {
					time.Sleep(3 * time.Second)
					// upload file
					log.Println(uploadObject(event.Name, bucketName))
				}
			case err, ok := <- watcher.Errors:
				if !ok {
					return
				}
				log.Println("error: ", err)
		}
	
	}
}
//(resp *s3.PutObjectOutput)
func uploadObject(filename, bucketName string) (resp *s3.PutObjectOutput) {
	f, err := os.Open(filename)
	if err != nil {
		log.Println("File not found : ", err)
	}
	log.Println("Uploading : ", filename)
	filenameSplit := strings.Split(filename, pathSeparator)
	resp, err = s3session.PutObject(&s3.PutObjectInput{
		Body:                      f,
		Bucket:                    aws.String(bucketName),
		Key: aws.String(filenameSplit[len(filenameSplit)-1]),
	})

	if err != nil {
		log.Println("Upload error: ", err)
		moveFile(filename, listConfig[2]+pathSeparator+filenameSplit[len(filenameSplit)-1])
	} else {
		moveFile(filename, listConfig[1]+pathSeparator+filenameSplit[len(filenameSplit)-1])
	}
	return resp
}

func moveFile(sourcePath, destPath string) error  {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("Could't open file : %s", err)
	}

	outputFile, err := os.Create(destPath)
	if err != nil {
		inputFile.Close()
		return fmt.Errorf("could't open destination file: %s", err)
	}
	defer outputFile.Close()
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return fmt.Errorf("Writing to outpu file failed: %s", err)
	}

	err = os.Remove(sourcePath)
	if err != nil {
		return fmt.Errorf("Failed removing original file: %s", err)
	}
	return nil
}

func exitErrorf(msg string, args ...interface{})  {
	fmt.Fprintf(os.Stderr, msg+"\n", args...)
	os.Exit(1)
}


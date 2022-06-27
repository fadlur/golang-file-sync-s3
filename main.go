package main

import "fmt"

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
	fmt.Print(welcomeText)
	
}


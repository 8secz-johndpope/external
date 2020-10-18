package s3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"gitlab.com/projectreferral/util/client/s3/models"
	"log"
	"net/http"
	"os"
)

// exposed methods that implementation/consumer can access
type Client interface {
	Init()
	UploadFile(r *http.Request, name string) (*s3manager.UploadOutput,error)
	DownloadFile(name string) (*os.File,error)
	PutEncryption(key string) (*s3.PutBucketEncryptionOutput,error)
}

type DefaultBucketClient struct {
	Session *session.Session
	configs *models.S3Configs
}

//Loads the S3 Key and creates a new session
func (c *DefaultBucketClient) Init() {

	if c.configs == nil {
		log.Fatal("S3 Configs not set")
	}

	if c.configs.Key == "" {
		log.Println("No s3 key found")
		os.Exit(1)
	}
	c.Session = session.Must(session.NewSession(&aws.Config{
		Region: aws.String(c.configs.Region)},
	))
	if c.Session == nil {
		log.Println("Initiation failed")
		os.Exit(1)
	}
	log.Println("Session Initiated")
}

func (c *DefaultBucketClient) UploadFile(r *http.Request, name string) (*s3manager.UploadOutput,error) {

	// Limit file size
	sizeErr := r.ParseMultipartForm(c.configs.PartSize)

	if !HandleError(sizeErr)  {

		file, header, fErr := r.FormFile(name)
		if !HandleError(fErr) {

			defer file.Close()

			filename := header.Filename
			size := header.Size
			log.Printf("file name %s size %d",filename,size)

			input := &s3manager.UploadInput{
				// Bucket to be used
				Bucket: aws.String(c.configs.Bucket),
				// Name of the file to be saved
				Key:    aws.String(filename),
				// File body
				Body:   file,
				// Encrypt file
				SSECustomerAlgorithm: aws.String(c.configs.EncryptionAlgorithm),
				SSECustomerKey : aws.String(c.configs.Key),
			}
			uploader := s3manager.NewUploader(c.Session)
			log.Println("uploader created")

			// Perform upload with multipart
			result, uErr := uploader.Upload(input, func(u *s3manager.Uploader) {
				u.PartSize = c.configs.PartSize
				u.LeavePartsOnError = true    // Don't delete the parts if the upload fails.
			})

			if !HandleError(uErr){
				log.Printf("%+v",result)
				return result, nil
			}

			return nil, uErr
		}
		return nil, fErr
	}
	return nil, sizeErr
}

func (c *DefaultBucketClient) DownloadFile(name string) (*os.File,error) {
	file, fErr := os.Create(c.configs.DownloadLocation + name)
	if !HandleError(fErr) && file == nil {
		log.Println("error creating file")
		return nil,  nil
	}

	defer file.Close()

	downloader := s3manager.NewDownloader(c.Session)
	_, err := downloader.Download(file, &s3.GetObjectInput{
		Bucket: aws.String(c.configs.Bucket),
		Key:    aws.String(name),
		SSECustomerAlgorithm: aws.String(c.configs.EncryptionAlgorithm),
		SSECustomerKey : aws.String(c.configs.Key),

	})

	//if there is an error
	if HandleError(err) {
		os.Remove(c.configs.DownloadLocation + name)
		return nil, err
	}

	return file, nil
}

//function for putting KMS key
func (c *DefaultBucketClient) PutEncryption(key string) (*s3.PutBucketEncryptionOutput,error) {
	defEnc := &s3.ServerSideEncryptionByDefault{KMSMasterKeyID: aws.String(key), SSEAlgorithm: aws.String(c.configs.EncryptionAlgorithm)}
	rule := &s3.ServerSideEncryptionRule{ApplyServerSideEncryptionByDefault: defEnc}
	rules := []*s3.ServerSideEncryptionRule{rule}
	serverConfig := &s3.ServerSideEncryptionConfiguration{Rules: rules}
	input := &s3.PutBucketEncryptionInput{Bucket: aws.String(c.configs.Bucket), ServerSideEncryptionConfiguration: serverConfig}
	svc := s3.New(c.Session)
	result, err := svc.PutBucketEncryption(input)
	if !HandleError(err) {
		log.Printf("Bucket %s now has KMS encryption by default %+v\n", c.configs.Bucket, result)
		return result, nil
	}
	return result, err
}

//Sets configs which are loaded on the API
func (c *DefaultBucketClient) SetConfigs(cfg *models.S3Configs) {

	if cfg != nil {
		c.configs = cfg
	}
}

//Custom made error
func HandleError(err error) bool {
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			default:
				log.Println(aerr.Error())
			}
		}else {
			log.Println(err.Error())
		}
		return true
	}
	return false
}

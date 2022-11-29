package main

import (
	"context"
	"log"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/joho/godotenv"
	"net/http"
)

func main() {
	//!load env variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return
	}

	//!setup gin
	r := gin.Default()
	r.Static("/assets", "./assets")
	r.LoadHTMLGlob("templates/*")
	r.MaxMultipartMemory = 10 << 20

	//!setup s3 uploader
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Printf("error: %v", err)
		return
	}
	client := s3.NewFromConfig(cfg)
	uploader := manager.NewUploader(client)

	r.GET("/upload", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Main website",
		})
	})
	r.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("image")
		if err != nil {
			c.HTML(404, "index.html", gin.H{
				"error": "Failed to upload image!",
			})
			return
		}
		err = c.SaveUploadedFile(file, "assets/uploads/"+file.Filename)
		if err != nil {
			c.HTML(404, "index.html", gin.H{
				"error": "Failed to upload image!",
			})
			return
		}
		f, err := file.Open()
		if err != nil {
			c.HTML(404, "index.html", gin.H{
				"error": "Failed to upload image!",
			})
			return
		}

		// defer file.Close()
		buff := make([]byte, 512) // docs tell that it take only first 512 bytes into consideration
		if _, err = f.Read(buff); err != nil {
			c.HTML(404, "index.html", gin.H{
				"error": "Failed to upload image!",
			})
			return
		}

		fmt.Println(http.DetectContentType(buff)) // do something based on your detection.

		switch http.DetectContentType(buff) {
		case "image/jpeg":
			result, uploadErr := uploader.Upload(context.TODO(), &s3.PutObjectInput{
				Bucket:      aws.String("image-uploader-ali"),
				Key:         aws.String(file.Filename),
				Body:        f,
				ACL:         "public-read",
				ContentType: aws.String("image/jpeg"),
			})
			if uploadErr != nil {
				c.HTML(404, "index.html", gin.H{
					"error": "Failed to upload image!",
				})
				return
			}
			c.HTML(http.StatusOK, "index.html", gin.H{
				"image": result.Location,
			})
			fmt.Println(result.Location)
			return
		case "image/png":
			result, uploadErr := uploader.Upload(context.TODO(), &s3.PutObjectInput{
				Bucket:      aws.String("image-uploader-ali"),
				Key:         aws.String(file.Filename),
				Body:        f,
				ACL:         "public-read",
				ContentType: aws.String("image/png"),
			})
			if uploadErr != nil {
				c.HTML(404, "index.html", gin.H{
					"error": "Failed to upload image!",
				})
				return
			}
			c.HTML(http.StatusOK, "index.html", gin.H{
				"image": result.Location,
			})
			fmt.Println(result.Location)
			return
		default:
			c.HTML(404, "index.html", gin.H{
				"error": "Please upload jpeg or png format!",
			})
		}
	})
	r.Run() // listen and serve on 0.0.0.0:8080
}

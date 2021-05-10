package main

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"syscall"
	"time"

	"github.com/YangSen-qn/go-curl/v2/curl"
	"github.com/qiniu/go-sdk/v7/client"
	"github.com/qiniu/go-sdk/v7/storage"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	transport := &curl.Transport{
		Transport: &http.Transport{
			IdleConnTimeout: 4 * time.Second * 60,
		},
		//CAPath:     "/Users/senyang/Desktop/QiNiu/Test/Go/test/examples/http-transport/curl/lib/resource/cacert.pem",
		ForceHTTP3: true,
	}
	client.DefaultClient.Client = &http.Client{Transport: transport}

	upload(10000, 10)

	fmt.Println("======= Done =======")
}

func upload(uploadCount int, goroutineCount int) {

	source := make(chan int, 100)
	go func() {
		for i := 0; i < uploadCount; i++ {
			source <- i + 1
		}
		close(source)
	}()

	wait := &sync.WaitGroup{}
	wait.Add(goroutineCount)

	for i := 0; i < goroutineCount; i++ {
		go func(source <-chan int, goroutineIndex int) {
			defer func(goroutineIndex int) {
				fmt.Printf("== goroutineIndex:%d done \n", goroutineIndex)
				wait.Done()
			}(goroutineIndex)

			for {
				index, ok := <-source
				if !ok {
					break
				} else {
					key := "http3_test_" + time.Now().Format("2006/01/02 15:04:05.999999")
					done := make(chan bool)
					go func() {
						select {
						case <-done:
							break
						case <-time.After(60 * time.Second):
							fmt.Printf("exit by timeout goroutineIndex:%d, index:%d key:%v \n", goroutineIndex, index, key)
							syscall.Exit(-1)
						}
					}()
					response, err := uploadFileToQiniu(key)
					fmt.Printf("goroutineIndex:%d, index:%d error:%v response:%v \n", goroutineIndex, index, err, response)
					done <- true
				}
			}
		}(source, i)
	}

	wait.Wait()
}

func uploadFileToQiniu(key string) (response storage.PutRet, err error) {

	filePath := "/Users/senyang/Desktop/QiNiu/pycharm.dmg"
	filePath = "/Users/senyang/Desktop/QiNiu/UploadResource_49M.zip"
	filePath = "/Users/senyang/Desktop/QiNiu/Image/image.png"
	filePath = "/Users/senyang/Desktop/Upload.zip"

	token := "HwFOxpYCQU6oXoZXFOTh1mq5ZZig6Yyocgk3BTZZ:yZ82TCOe1jJiK4NWYJIs64VWzUY=:eyJzY29wZSI6ImtvZG8tcGhvbmUtem9uZTAtc3BhY2UiLCJkZWFkbGluZSI6MTYyNTgxODQxNiwgInJldHVybkJvZHkiOiJ7XCJjYWxsYmFja1VybFwiOlwiaHR0cDpcL1wvY2FsbGJhY2suZGV2LnFpbml1LmlvXCIsIFwiZm9vXCI6JCh4OmZvbyksIFwiYmFyXCI6JCh4OmJhciksIFwibWltZVR5cGVcIjokKG1pbWVUeXBlKSwgXCJoYXNoXCI6JChldGFnKSwgXCJrZXlcIjokKGtleSksIFwiZm5hbWVcIjokKGZuYW1lKX0ifQ=="

	config := &storage.Config{
		Zone: &storage.Region{
			SrcUpHosts: []string{"up.qiniu.com"},
		},
		Region:   nil,
		UseHTTPS: true,
	}
	ctx := context.Background()

	//uploader := storage.NewResumeUploader(config)
	uploader := storage.NewResumeUploaderV2(config);

	err = uploader.PutFile(ctx, &response, token, key, filePath, nil)
	return
}

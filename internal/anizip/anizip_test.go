package anizip

import (
	"fmt"
	"sync"
	"testing"
)

func TestFetchAniZipMedia(t *testing.T) {

	media, err := FetchAniZipMedia("anilist", 1)

	if err != nil {
		t.Error("expected media, found error, ", err)
	}

	fmt.Println(media)

}

func TestFetchAniZipMediaC2(t *testing.T) {

	ids := []int{1, 21, 55, 1001}
	resultChan := make(chan *Media, len(ids))
	errorChan := make(chan error, len(ids))

	go func() {
		var wg sync.WaitGroup

		for _, id := range ids {
			wg.Add(1)
			go func(_id int) {
				defer wg.Done()
				res, err := FetchAniZipMedia("anilist", _id)
				if err != nil {
					errorChan <- err
				} else {
					resultChan <- res
				}
			}(id)
		}

		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	for aniZipData := range resultChan {
		fmt.Println("Received data:", aniZipData.GetTitle())
	}

	for err := range errorChan {
		fmt.Println("Error:", err)
	}

}

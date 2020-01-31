package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"

	"golang.org/x/sync/semaphore"
)

func main() {
	failed := make(chan string)
	urls := make([]string, 0)
	var wg sync.WaitGroup
	sem := semaphore.NewWeighted(50)

	file, _ := os.Open(".\\testUrls.txt")
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		urls = append(urls, scanner.Text())
	}

	wg.Add(len(urls))

	fmt.Printf("found %d urls\n", len(urls))

	processed := 0
	for _, url := range urls {
		go func(url string) {
			sem.Acquire(context.TODO(), 1)
			callAPI(url, failed, &wg)
			sem.Release(1)
			processed++

			fmt.Printf("\rProcessed: %d", processed)
		}(url)
	}

	go func() {
		errorsFile, err := os.Create(".\\error.txt")
		if err != nil {
			return
		}
		defer errorsFile.Close()

		fileWriter := bufio.NewWriter(errorsFile)
		for f := range failed {
			fileWriter.WriteString(f + "\n")
		}
	}()

	wg.Wait()
	close(failed)
}

func callAPI(url string, failed chan<- string, wg *sync.WaitGroup) {
	resp, err := http.Get(url)

	if err == nil {
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			failed <- url
		}
	} else {
		failed <- url
	}

	wg.Done()
}

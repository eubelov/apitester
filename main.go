package main

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

func main() {
	failed := make(chan string)
	urls := make([]string, 0)
	callTimes := make(chan time.Duration)

	var wg, writers sync.WaitGroup
	sem := semaphore.NewWeighted(10)

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

			callTimes <- callAPI(url, failed)

			sem.Release(1)
			processed++

			fmt.Printf("\rProcessed: %d", processed)

			wg.Done()
		}(url)
	}

	go func() {
		writers.Add(1)

		f, err := os.Create(".\\requestsDurations.txt")
		if err != nil {
			return
		}
		defer f.Close()

		fileWriter := bufio.NewWriter(f)
		defer fileWriter.Flush()

		for f := range callTimes {
			fileWriter.WriteString(strconv.FormatInt(f.Milliseconds(), 10) + "\n")
		}

		writers.Done()
	}()

	go func() {
		writers.Add(1)

		errorsFile, err := os.Create(".\\error.txt")
		if err != nil {
			return
		}
		defer errorsFile.Close()

		fileWriter := bufio.NewWriter(errorsFile)
		defer fileWriter.Flush()

		for f := range failed {
			fileWriter.WriteString(f + "\n")
		}

		writers.Done()
	}()

	wg.Wait()

	close(failed)
	close(callTimes)

	writers.Wait()
}

func callAPI(url string, failed chan<- string) time.Duration {
	start := time.Now()

	resp, err := http.Get(url)

	if err == nil {
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			failed <- url
		}
	} else {
		failed <- url
	}

	duration := time.Since(start)

	return duration
}

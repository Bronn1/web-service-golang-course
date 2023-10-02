package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// сюда писать код
func SingleHash(in, out chan interface{}) {
	md5Mutex := &sync.Mutex{}
	waitGroup := &sync.WaitGroup{}
	for data := range in {
		waitGroup.Add(1)
		go SingleHashWorker(fmt.Sprintf("%v", data), out, waitGroup, md5Mutex)
	}

	waitGroup.Wait()
}

func SingleHashWorker(data string, out chan interface{}, waitGroup *sync.WaitGroup, md5Mutex *sync.Mutex) {
	defer waitGroup.Done()
	hash1 := make(chan string)
	go func(data string, hash1 chan string) {
		hash1 <- DataSignerCrc32(data)
	}(data, hash1)

	hash2 := make(chan string)
	go func(data string, hash2 chan string) {
		md5Mutex.Lock()
		md5Hash := DataSignerMd5(data)
		md5Mutex.Unlock()
		hash2 <- DataSignerCrc32(md5Hash)
	}(data, hash2)

	out <- fmt.Sprintf("%v~%v", <-hash1, <-hash2)
}

func MultiHash(in, out chan interface{}) {
	waitGroup := &sync.WaitGroup{}
	for data := range in {
		waitGroup.Add(1)
		go MultiHashWorker(fmt.Sprintf("%v", data), out, waitGroup)
	}

	waitGroup.Wait()
}

func MultiHashWorker(data string, out chan interface{}, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	var result string
	multiHashResults := make([]string, 6)
	wg := &sync.WaitGroup{}
	mutex := &sync.Mutex{}
	for th := 0; th < 6; th++ {
		wg.Add(1)
		go func(th int, strData string, wg *sync.WaitGroup, multiHashResults []string) {
			defer wg.Done()
			crc32Hash := DataSignerCrc32(strData)
			mutex.Lock()
			multiHashResults[th] = crc32Hash
			mutex.Unlock()
		}(th, strconv.Itoa(th)+data, wg, multiHashResults)
	}

	wg.Wait()
	for _, hash := range multiHashResults {
		result += hash
	}
	out <- result
}

func CombineResults(in, out chan interface{}) {
	calculatedHashes := []string{}
	for data := range in {
		calculatedHashes = append(calculatedHashes, fmt.Sprintf("%v", data))
	}
	sort.Strings(calculatedHashes)
	var result string
	for _, dataHash := range calculatedHashes {
		result += dataHash + "_"
	}
	out <- strings.TrimRight(result, "_")
}

func Pipelineworker(currJob job, wg *sync.WaitGroup, in, out chan interface{}) {
	defer func() {
		wg.Done()
		close(out)
	}()
	currJob(in, out)
}

func ExecutePipeline(jobs ...job) {
	pipelineWaitGroup := &sync.WaitGroup{}
	chanIn := make(chan interface{}, MaxInputDataLen)
	chanOut := make(chan interface{}, MaxInputDataLen)
	for _, currJob := range jobs {
		pipelineWaitGroup.Add(1)
		go Pipelineworker(currJob, pipelineWaitGroup, chanIn, chanOut)
		chanIn = chanOut
		chanOut = make(chan interface{}, MaxInputDataLen)
	}

	pipelineWaitGroup.Wait()
}

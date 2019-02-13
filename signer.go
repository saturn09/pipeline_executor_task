package main

import (
	"fmt"
	"strings"
	"sync"
)

// сюда писать код

func ExecutePipeline(jobs ...job) {
	actualInput := make(chan interface{})
	actualOutput := make(chan interface{})

	length := len(jobs)
	for i, currJob := range jobs {
		if length > i+1 {
			go currJob(actualOutput, actualInput)
		} else {
			go currJob(actualInput, actualOutput)
		}
	}
}

func SingleHash(in, out chan interface{}) {
	var strData, md5, nestedHash string
	strData = fmt.Sprint(<-out)
	mu := &sync.Mutex{}
	go func(mu *sync.Mutex) { // md5
		mu.Lock()
		out <- DataSignerMd5(strData) // Lock goroutine to avoid parallel access
		mu.Unlock()
	}(mu)

	go func() { // crc(md5)
		md5 = fmt.Sprint(<-out)
		out <- DataSignerCrc32(md5)
	}()

	go func() { // crc + crc(md5)
		nestedHash = fmt.Sprint(<-out)
		out <- DataSignerCrc32(strData) + "~" + nestedHash
	}()
}

func MultiHash(in, out chan interface{}) {
	var result []string
	data := fmt.Sprint(<-out)
	wg := &sync.WaitGroup{}
	for i := 0; i < 6; i ++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			out <- DataSignerCrc32(fmt.Sprint(i) + data)
		}(wg)
	}
	wg.Wait()
	for val := range out {
		result = append(result, val.(string))
	}
	out <- strings.Join(result, "")
}

func CombineResults(in, out chan interface{}) {
	//sort.Strings(results)
	//combinedResult := strings.Join(results, "_")
	fmt.Println("combinedResult")
	// Write total result to chan
}

func main() {
	jobs := []job{
		job(func(in, out chan interface{}) {
			for i := 0; i < 10; i ++ {
				out <- i
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
	}

	ExecutePipeline(jobs...)
}

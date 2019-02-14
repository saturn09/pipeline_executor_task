package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// сюда писать код

func ExecutePipeline(jobs ...job) {
	actualInput := make(chan interface{})
	actualOutput := make(chan interface{})

	for i, currJob := range jobs {
		if i == 1 {
			go currJob(actualInput, actualOutput)
		} else {
			go currJob(actualOutput, actualInput)
		}
	}
	time.Sleep(1 * time.Millisecond)
}

func SingleHash(in, out chan interface{}) {
	log.Println("logging singleHash")
	var strData, md5, nestedHash string
	strData = fmt.Sprint(<-in)
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

	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func(wg *sync.WaitGroup) { // crc + crc(md5)
		defer wg.Done()
		nestedHash = fmt.Sprint(<-out)
		out <- DataSignerCrc32(strData) + "~" + nestedHash
	}(wg)
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	log.Println("Logging multiHash")
	var result []string
	data := fmt.Sprint(<-in)
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
	log.Println("combinedResult")
	res := <-in
	out <- strings.Join([]string{res.(string)}, "_")
	//sort.Strings(results)
	//combinedResult := strings.Join(results, "_")
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

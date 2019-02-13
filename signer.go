package main

import (
	"fmt"
	"sync"
)

// сюда писать код

func ExecutePipeline(jobs ...job) {
	in := make(chan interface{})
	out := make(chan interface{})

	for i, currJob := range jobs {
		if len(jobs) > i + 1{
			go currJob(out, in)
		} else {
			go currJob(in, out)
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
	//var result []string
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

	//result = append(result, DataSignerCrc32(fmt.Sprint(i)+data))
	//return strings.Join(result, "")
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

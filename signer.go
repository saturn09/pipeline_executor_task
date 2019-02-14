package main

import (
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// сюда писать код

const multiHashLoops = 6

func ExecutePipeline(jobs ...job) {
	input := make(chan interface{})
	wg := &sync.WaitGroup{}

	for _, currJob := range jobs {
		wg.Add(1)

		output := make(chan interface{})
		go pipelineExecutor(currJob, input, output, wg)
		input = output
	}
	wg.Wait()
}

func pipelineExecutor(job job, in, out chan interface{}, wg *sync.WaitGroup) {

	defer wg.Done()
	defer close(out)

	job(in, out)
}

func SingleHash(in, out chan interface{}) {
	log.Println("logging singleHash")
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	for val := range in {
		wg.Add(1)
		go singleHashExecutor(val, out, wg, mu)
	}
	wg.Wait()
}

func singleHashExecutor(payload interface{}, out chan interface{}, wg *sync.WaitGroup, mu *sync.Mutex) {
	defer wg.Done()
	strData := strconv.Itoa(payload.(int))

	mu.Lock()
	md5 := DataSignerMd5(strData) // Lock goroutine to avoid parallel access
	mu.Unlock()

	crcChan := make(chan string)
	go parallelCrc(strData, crcChan)

	crc32Md5 := DataSignerCrc32(md5)
	crc32 := <-crcChan
	out <- crc32 + "~" + crc32Md5
}

func parallelCrc(strData string, crcChan chan string) {
	crcChan <- DataSignerCrc32(strData)
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for i := range in {
		wg.Add(1)

		go multiHashExecutor(wg, out, i.(string))
	}
	wg.Wait()
}

func multiHashExecutor(wg *sync.WaitGroup, out chan interface{}, val string) {
	defer wg.Done()

	innerWg := &sync.WaitGroup{}
	mu := &sync.Mutex{}
	container := make([]string, multiHashLoops)

	for i := 0; i < multiHashLoops; i ++ {
		innerWg.Add(1)

		strData := strconv.Itoa(i) + val

		go func(container []string, payload string, ix int, wg *sync.WaitGroup, mu *sync.Mutex) {
			defer wg.Done()
			payload = DataSignerCrc32(payload)
			mu.Lock()
			container[ix] = payload
			mu.Unlock()
		}(container, strData, i, innerWg, mu)
	}
	innerWg.Wait()

	joinResult := strings.Join(container, "")
	log.Printf("MultiHash result %s", joinResult)
	out <- joinResult
}

func CombineResults(in, out chan interface{}) {
	var results []string
	for val := range in {
		results = append(results, val.(string))
	}
	sort.Strings(results)
	result := strings.Join(results, "_")

	log.Printf("CombineResults \n%s\n", result)

	out <- result
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
		job(func(in, out chan interface{}) {
			data := <-in
			log.Println(data.(string))
		}),
	}

	ExecutePipeline(jobs...)
}

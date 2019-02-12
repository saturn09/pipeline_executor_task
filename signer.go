package main

import (
	"fmt"
	"log"
	"sort"
	"strings"
)

// сюда писать код

type jobGenerator func() string

// sync jobs handling
func ExecutePipeline(jobs ...jobGenerator) {
	var (
		calculatedHashes []string
		result           string
	)
	for _, job := range jobs {
		result = job()
		log.Println(result)
		calculatedHashes = append(calculatedHashes, result)
	}
	CombineResults(calculatedHashes)
}

func syncJob(input int) jobGenerator {
	return func() string {
		return MultiHash(SingleHash(input))
	}
}

func SingleHash(data int) string {
	var strData, md5, crc32, nestedHash string
	strData = fmt.Sprint(data)
	md5 = DataSignerMd5(strData)
	crc32 = DataSignerCrc32(strData)
	nestedHash = DataSignerCrc32(md5)
	return crc32 + "~" + nestedHash
}

func MultiHash(data string) string {
	var result []string
	for i := 0; i < 6; i ++ {
		result = append(result, DataSignerCrc32(fmt.Sprint(i)+data))
	}
	return strings.Join(result, "")
}

func CombineResults(results []string) string {
	sort.Strings(results)
	combinedResult := strings.Join(results, "_")
	fmt.Println(combinedResult)
	return combinedResult
}

func main() {
	var jobs []jobGenerator
	for i := 0; i < 10; i ++ {
		jobs = append(jobs, syncJob(i))
	}
	ExecutePipeline(jobs...)
}

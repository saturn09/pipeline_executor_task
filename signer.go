package main

import "fmt"

// сюда писать код

type jobFunc func(in, out chan interface{})

func ExecutePipeline(jobs...jobFunc) {
	
}

func SingleHash(data int) string {
	DataSignerCrc32(fmt.Sprint(data))
}

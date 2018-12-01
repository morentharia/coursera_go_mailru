package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

func ExecutePipeline(hashSignJobs ...job) {
	wg := &sync.WaitGroup{}
	wg.Add(len(hashSignJobs))
	in, out := make(chan interface{}), make(chan interface{})
	defer close(in)
	for _, jobItem := range hashSignJobs {
		go func(jobItem job, in, out chan interface{}) {
			defer wg.Done()
			defer close(out)
			jobItem(in, out)
		}(jobItem, in, out)
		in, out = out, make(chan interface{})
	}
	defer wg.Wait()
}

func myCrc32(data string) <-chan string {
	result := make(chan string, 1)
	go func(result chan<- string) {
		result <- DataSignerCrc32(data)
	}(result)
	return result
}

func SingleHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	mu := &sync.Mutex{}

	for input := range in {
		wg.Add(1)
		go func(input interface{}, out chan interface{}) {
			defer wg.Done()
			data := fmt.Sprintf("%v", input)

			mu.Lock()
			md5Data := DataSignerMd5(data)
			mu.Unlock()

			var left, right string
			leftCh, rightCh := myCrc32(data), myCrc32(md5Data)
			for i := 0; i < 2; i++ {
				select {
				case left = <-leftCh:
				case right = <-rightCh:
				}
			}
			out <- left + "~" + right
		}(input, out)
	}
	wg.Wait()
}

func MultiHash(in, out chan interface{}) {
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	for input := range in {
		wg.Add(1)
		go func(input interface{}, out chan interface{}) {
			defer wg.Done()
			data := input.(string)
			hashes := make([]string, 6)
			wgHashes := &sync.WaitGroup{}
			for i := 0; i < 6; i++ {
				wgHashes.Add(1)
				go func(i int) {
					defer wgHashes.Done()
					hash := DataSignerCrc32(fmt.Sprintf("%d%s", i, data))
					mu.Lock()
					defer mu.Unlock()
					hashes[i] = hash
				}(i)
			}
			wgHashes.Wait()
			out <- strings.Join(hashes, "")
		}(input, out)
	}
	wg.Wait()
}

func CombineResults(in, out chan interface{}) {
	res := []string{}
	for input := range in {
		res = append(res, input.(string))
	}
	sort.Strings(res)
	out <- strings.Join(res, "_")
}

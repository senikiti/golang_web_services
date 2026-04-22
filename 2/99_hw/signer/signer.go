package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// сюда писать код
type indexedResult struct {
	idx int
	val string
}
type Service struct {
	id   string
	sJob job
	next *Service
	in   chan interface{}
	out  chan interface{}
}

func (s Service) passToNext() {
	go func(s Service) {
		defer close(s.next.in)
		for v := range s.out {
			fmt.Println("Passing to next service job from", s.id)
			s.next.in <- v
		}
	}(s)
}

func (s Service) run() {
	fmt.Println("Running service job " + s.id)
	go func(s Service) {
		defer close(s.out)
		s.sJob(s.in, s.out)
	}(s)
}

func ExecutePipeline(jobs ...job) {
	jobCount := len(jobs)
	serviceQueue := make([]*Service, jobCount)
	for i := jobCount - 1; i >= 0; i-- {
		serviceQueue[i] = &Service{id: "j" + strconv.Itoa(i), sJob: jobs[i], in: make(chan interface{}), out: make(chan interface{})}
		if i != jobCount-1 { // No need for a passToNext listener on a last service
			serviceQueue[i].next = serviceQueue[i+1]
			serviceQueue[i].passToNext() // prepare channel listeners that get value from service[i] out and put it to service[i+1] in
		}
	}
	for i := jobCount - 1; i >= 0; i-- {
		serviceQueue[i].run() // Initialize pipeline in a reverse order(last to first) and first service sends first value into pipe
	}
	// Listen to the last service channel out until last service is done and out is closed
	for v := range serviceQueue[jobCount-1].out {
		fmt.Println("Final value: ", v)
	}
}

func OneSingleHash(i int, md5v string) string {
	var data = strconv.Itoa(i)
	crcvChan, crcvMd5Chan := make(chan string, 1), make(chan string, 1)
	go func(s string) {
		crcvChan <- DataSignerCrc32(s)
	}(data)
	go func(s string) {
		crcvMd5Chan <- DataSignerCrc32(s)
	}(md5v)
	crcv := <-crcvChan
	crcMd5v := <-crcvMd5Chan
	fmt.Println("Returning SingleHash for ", data)
	return crcv + "~" + crcMd5v
}

func SingleHash(in, out chan interface{}) {
	var wg sync.WaitGroup
	for i := range in {
		int_i := i.(int)
		var data = strconv.Itoa(int_i)
		md5v := DataSignerMd5(data)
		wg.Add(1)
		go func(i int, md5v string) {
			defer wg.Done()
			osh := OneSingleHash(i, md5v)
			out <- osh
		}(int_i, md5v)
	}
	wg.Wait()
}

func OneMultiHash(sh string) string {
	fmt.Println("In OneMultiHash")
	count := 6
	indexedResults := make([]string, count)
	resChan := make(chan indexedResult, count)
	for j := 0; j < count; j++ {
		jstr := strconv.Itoa(j)
		go func(idx int, s string) {
			resChan <- indexedResult{idx: idx, val: DataSignerCrc32(s)}
		}(j, jstr+sh)
	}
	for i := 0; i < count; i++ {
		idxRes := <-resChan
		indexedResults[idxRes.idx] = idxRes.val
	}
	close(resChan)
	fmt.Println("Got result from parallel MH CRC: ", indexedResults)
	return strings.Join(indexedResults, "")
}

func MultiHash(in, out chan interface{}) {
	var wg sync.WaitGroup
	for sh := range in {
		fmt.Println("In MultiHash")
		wg.Add(1)
		go func(sh string) {
			defer wg.Done()
			out <- OneMultiHash(sh)
		}(sh.(string))
	}
	wg.Wait()
}

func CombineResults(in chan interface{}, out chan interface{}) {
	mh_strs := []string{}
	for stri := range in {
		mh_strs = append(mh_strs, stri.(string))
	}
	fmt.Println("Done combining, passing further to out")
	sort.Strings(mh_strs)
	out <- strings.Join(mh_strs, "_")
}

// func main() {
// 	res_str := ""
// 	inputData := []int{0, 1, 1, 2, 3, 5, 8}
// 	jobs := []job{
// 		job(func(in, out chan interface{}) {
// 			fmt.Println("In IntialJob")
// 			for _, i := range inputData {
// 				fmt.Println("Writing to out: ", i)
// 				out <- i
// 			}
// 		}),
// 		job(SingleHash),
// 		job(MultiHash),
// 		job(CombineResults),
// 		job(func(in, out chan interface{}) {
// 			fmt.Println("In FinalJob")
// 			res_str = (<-in).(string)
// 			fmt.Println("FinalJob done, resuming", res_str)
// 		}),
// 	}
// 	ExecutePipeline(jobs...)
// 	fmt.Println("Main done", res_str)
// }

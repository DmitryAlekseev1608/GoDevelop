package main

import (
	"fmt"
	"sync"
	"strconv"
	"sort"
	"strings"
)

func ExecutePipeline (inputData []int) {
	wg := &sync.WaitGroup{}
	in := make(chan interface{})
	out := make(chan interface{})

	wg.Add(1)
	go SingleHash(in, out, wg)
	in <- strconv.Itoa(inputData[0])
	res := <- out
	fmt.Println(res)
	wg.Wait()

	wg.Add(1)
	go MultiHash(in, out, wg)
	in <- res
	res1 := <- out
	wg.Wait()
	fmt.Println(res1)

	wg.Add(1)
	go SingleHash(in, out, wg)
	in <- strconv.Itoa(inputData[1])
	res = <- out
	fmt.Println(res)
	wg.Wait()

	wg.Add(1)
	go MultiHash(in, out, wg)
	in <- res
	res2 := <- out
	wg.Wait()
	fmt.Println(res2)

	wg.Add(1)
	go CombineResults(in, out, wg)
	in <- res1
	in <- res2
	fmt.Println("---1---")
	fmt.Println(len(out))
	close(in)
	res = <- out
	fmt.Println("---2---")
	wg.Wait()
	fmt.Println(res)

	return
}


func SingleHash(in, out chan interface{}, wg *sync.WaitGroup){
	defer wg.Done() 
	data := <- in
	var dataHashCrc32 = DataSignerCrc32(data.(string))
	var dataHashMd5 = DataSignerMd5(data.(string))
	var dataHashCrc32Md5 = DataSignerCrc32(dataHashMd5)
	result :=  dataHashCrc32 + "~" + dataHashCrc32Md5
	out <- result
}

func MultiHash(in, out chan interface{}, wg *sync.WaitGroup){
	defer wg.Done() 
	data := <- in
	var result = ""
	for th := 0; th<6; th++ {
		var dataHashCrc32 = DataSignerCrc32(strconv.Itoa(th) + data.(string))
		result += dataHashCrc32
	}
	out <- result
}

func CombineResults(in, out chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	code := make([]string, 0, 100)
	for data := range in {
		code = append(code, data.(string))
	}
	sort.Strings(code)
	var result = strings.Join(code, "_")
	out <- result
}

func main() {
	inputData := []int{0,1}
	ExecutePipeline(inputData)
}

package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

var wg sync.WaitGroup
var mu sync.Mutex

func ExecutePipeline(curJob ...job) {

	inChannels := make([]chan interface{}, len(curJob)+1)
	outChannels := make([]chan interface{}, len(curJob)+1)
	wgSub := make([]sync.WaitGroup, len(curJob)+1)
	inChannels[0] = make(chan interface{}, 1)

	for i, function := range curJob {
		outChannels[i] = make(chan interface{}, 1)
		inChannels[i+1] = outChannels[i]

		if i == 0 || i > 2 {
			wg.Add(1)
			go func(i int, function func(chan interface{}, chan interface{})) {
				defer wg.Done()
				defer close(inChannels[i+1])
				function(inChannels[i], outChannels[i])
				}(i, function)
		}

		
		if i == 2 || i == 1 {
			wg.Add(1)
			go func(i int, function func(chan interface{}, chan interface{})) {
				defer wg.Done()
				defer close(inChannels[i+1])
				for j := 0; j<10; j++ {
					wgSub[i].Add(1)
					go func(i int, function func(chan interface{}, chan interface{})) {
						defer wgSub[i].Done()
						function(inChannels[i], outChannels[i])
					}(i, function)
				}
				wgSub[i].Wait()
			}(i, function)
		}

	}
	
	wg.Wait()
}


func SingleHash(in, out chan interface{}) {
	for data := range in {
		var data_str = strconv.Itoa(data.(int))
		var wgF sync.WaitGroup
		var dataHashCrc32 string
		var dataHashMd5 string
		var dataHashCrc32Md5 string

		wgF.Add(1)
		go func() {
			defer wgF.Done()
			dataHashCrc32 = DataSignerCrc32(data_str)
		}()

		wgF.Add(1)
		go func() {
			defer wgF.Done()
			mu.Lock()
			dataHashMd5 = DataSignerMd5(data_str)
			mu.Unlock()
			dataHashCrc32Md5 = DataSignerCrc32(dataHashMd5)
		}()

		wgF.Wait()
		result := dataHashCrc32 + "~" + dataHashCrc32Md5
		out <- result
	}
}

func MultiHash(in, out chan interface{}) {
	for data := range in {
		var data_str = data.(string)
		var wgF sync.WaitGroup
		var muF sync.Mutex
		var result string

		var code = make([]string, 6)
		for th := 0; th < 6; th++ {
			code[th] = ""
		}

		for th := 0; th < 6; th++ {
			wgF.Add(1)
			go func(th int) {
				defer wgF.Done()
				var dataHashCrc32 = DataSignerCrc32(strconv.Itoa(th) + data_str)
				muF.Lock()
				code[th] = dataHashCrc32
				muF.Unlock()
			}(th)
		}
		wgF.Wait()
		result = strings.Join(code, "")
		out <- result
	}
}

func CombineResults(in, out chan interface{}) {
	code := make([]string, 0, 100)
	for data := range in {
		code = append(code, data.(string))
		sort.Strings(code)
	}
	var result = strings.Join(code, "_")
	out <- result
}

func ExecutePipeline(curJob ...job) {
	var wg sync.WaitGroup
	inChannels := make([]chan interface{}, len(curJob))
	outChannels := make([]chan interface{}, len(curJob))
	for i, function := range curJob {
		wg.Add(1)
		inChannels[i] = make(chan interface{}, 1)
		outChannels[i] = make(chan interface{}, 1)
		go func(i int, function func(chan interface{}, chan interface{})) {
			defer wg.Done()
			function(inChannels[i], outChannels[i])
		}(i, function)
	}
	var existData = false
	var existWork = false
	count := make([]bool, len(curJob)-1)

	for i := range count {
		count[i] = false
	}
	count[0] = true
	
	outerLoop:
	for {
		for i, _ := range outChannels {
			if len(outChannels[i]) != 0 {
				num := <- outChannels[i]
				inChannels[i+1] <- num
				count[i] = false
				if i != len(curJob)-2 {
					count[i+1] = true
				}
				fmt.Println(count)
			}
		}
		
		for i, ch := range outChannels {
			if len(ch) != 0 || len(inChannels[i]) != 0 {
				existData = true
				break
			}
		}


		for _, val := range count {
			if val {
				existWork = true
				break
			}
		}

		if !existData && !existWork {
			for i, ch := range outChannels {
				close(inChannels[i])
				close(ch)
			}
			wg.Wait()
			break outerLoop
		}

		existData = false
		existWork = false
	}
}
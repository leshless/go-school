package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

const (
	MaxChanLen = 100
	MaxWorkers = 10
)

func SingleHash(in, out chan interface{}) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	worker := func() {
		defer wg.Done()

		var localWG sync.WaitGroup

		var firstHash string 
		localWorker1 := func (data interface{}) {
			defer localWG.Done()
			
			dataString := fmt.Sprintf("%d", data.(int))

			firstHash = DataSignerCrc32(dataString)
		}
		
		var secondHash string
		localWorker2 := func (data interface{}) {
			defer localWG.Done()

			dataString := fmt.Sprintf("%d", data.(int))

			mu.Lock()
			md5Hash := DataSignerMd5(dataString)
			mu.Unlock()

			secondHash = DataSignerCrc32(md5Hash)
		}
		
		for data := range in {
			localWG.Add(2)
			go localWorker1(data)
			go localWorker2(data)

			localWG.Wait()

			hash := fmt.Sprintf("%s~%s", firstHash, secondHash)

			out <- hash
		}
	}

	
	wg.Add(MaxWorkers)
	for i := 0; i < MaxWorkers; i++ {
		go worker()
	}

	wg.Wait()
	close(out)
}

func MultiHash(in, out chan interface{}) {
	var wg sync.WaitGroup

	hashes := make([]string, 6)

	worker := func() {
		defer wg.Done()

		var localWG sync.WaitGroup

		localWorker := func(data interface{}, i int) {
			defer localWG.Done()

			dataString := fmt.Sprintf("%d%s", i, data.(string))
			hashes[i] = DataSignerCrc32(dataString)
		}
		
		for data := range in {
			localWG.Add(6)
			for i := 0; i < 6; i++ {
				go localWorker(data, i)
			}

			localWG.Wait()

			out <- strings.Join(hashes, "")
		}
	}
	
	wg.Add(MaxWorkers)
	for i := 0; i < MaxWorkers; i++ {
		go worker()
	}
	
	wg.Wait()
	close(out)
}

func CombineResults(in, out chan interface{}) {
	dataAcc := make([]interface{}, 0, MaxChanLen)

	for data := range in {
		dataAcc = append(dataAcc, data)
	}

	dataAccString := make([]string, 0, MaxChanLen)
	for _, data := range dataAcc {
		dataAccString = append(dataAccString, data.(string))
	}

	sort.Slice(dataAccString, func(i, j int) bool {
		return dataAccString[i] < dataAccString[j]
	})


	out <- strings.Join(dataAccString, "_")

	close(out)
}

func ExecutePipeline(jobs ...job) {
	inChan := make(chan interface{}, MaxChanLen)

	for _, job := range jobs{
		outChan := make(chan interface{}, MaxChanLen)

		go job(inChan, outChan)

		inChan = outChan
	}

	for _ = range inChan {
		// wait until the last channel is closed
	}
}
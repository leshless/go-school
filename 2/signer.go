package main

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

const (
	MaxWorkers = 7
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
}

func MultiHash(in, out chan interface{}) {
	var wg sync.WaitGroup
	
	worker := func() {
		defer wg.Done()
		
		var localWG sync.WaitGroup
		var localMu sync.Mutex

		hashes := make([]string, 6)

		localWorker := func(data interface{}, i int) {
			defer localWG.Done()

			dataString := fmt.Sprintf("%d%s", i, data.(string))
			
			hash := DataSignerCrc32(dataString)

			localMu.Lock()
			hashes[i] = hash
			localMu.Unlock()
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
}

func CombineResults(in, out chan interface{}) {
	dataAcc := make([]interface{}, 0, MaxInputDataLen)

	for data := range in {
		dataAcc = append(dataAcc, data)
	}

	dataAccString := make([]string, 0, MaxInputDataLen)
	for _, data := range dataAcc {
		dataAccString = append(dataAccString, data.(string))
	}

	sort.Slice(dataAccString, func(i, j int) bool {
		return dataAccString[i] < dataAccString[j]
	})


	out <- strings.Join(dataAccString, "_")
}

func ExecutePipeline(jobs ...job) {
	nJobs := len(jobs)

	var wg sync.WaitGroup
	wg.Add(nJobs)

	chans := make([]chan interface{}, nJobs + 1)
	for i := 0; i < nJobs + 1; i++ {
		chans[i] = make(chan interface{}, MaxInputDataLen)
	}

	for i, job := range jobs{
		i := i
		job := job

		go func(){
			defer wg.Done()
			defer close(chans[i+1])
			
			job(chans[i], chans[i+1])
		}()
	}

	wg.Wait()
}
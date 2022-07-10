// counting card numbers
package main

import (
	"crypto/sha256"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	//   binbin0123456789
	max      = 9999999999
	numSize  = 10
	salt     = "C6qpl4nCgYhg08vTXaQs"
	panCount = 7
)

var found int32
var finished int32
var position int

func main() {
	start := time.Now()

	iins := []string{"517955", "524602", "533736", "540616", "541435", "544123", "550583"}

	// read file
	in, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	// [hash]level
	winners := make(map[string]string)

	// parse CSV
	r := csv.NewReader(strings.NewReader(string(in)))

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		key := record[0]
		winners[key] = record[1]
	}

	// calculate
	var wg sync.WaitGroup

	mtx := &sync.RWMutex{}

	for _, iin := range iins {
		go worker(winners, iin, &wg, mtx)
		wg.Add(1)

	}

	// status
	for {
		time.Sleep(10 * time.Second)
		now := time.Now()
		elapsed := now.Sub(start)
		log.Printf("time elapsed: %s, hashes found: %d, position: %d/%d", elapsed, found, position, max)

		if finished == panCount {
			break
		}
	}
	wg.Wait()
	end := time.Now()
	elapsed := end.Sub(start)
	log.Printf("ALL DONE. time elapsed: %v, speed: %f hashes/s\n", elapsed, max*panCount/elapsed.Seconds())
}

func worker(winners map[string]string, iin string, wg *sync.WaitGroup, mtx *sync.RWMutex) {
	i := 0
	for i <= max {
		numStr := countDigits(i)
		padLen := numSize - numStr

		pan := iin + strings.Repeat("0", padLen) + strconv.Itoa(i)

		pan19 := pan + "000"

		calcHash := sha256.Sum256([]byte(salt + pan19))
		calcHashStr := fmt.Sprintf("%x", calcHash)

		level, present := winners[calcHashStr]
		if present {
			fmt.Printf("%s,%s,%s\n", calcHashStr, level, pan)
			atomic.AddInt32(&found, 1)
		}

		position = i

		i++
	}
	log.Printf("iin ok: %s", iin)
	atomic.AddInt32(&finished, 1)

	wg.Done()
}

func countDigits(number int) int {
	if number == 0 {
		return 1
	}
	count := 0
	for number != 0 {
		number /= 10
		count += 1
	}
	return count
}

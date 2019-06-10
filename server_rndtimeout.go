//Simple  Lazy and Very Random Server
package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"
	"os"

	"flag"
	"log"
	"runtime"
	"runtime/pprof"

)

// Переменные для формирования  cpu and memory profile
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile `file`")
var memprofile = flag.String("memprofile", "", "write memory profile to `file`")
var countTime int
var f os.File

func main() {

	////////////////////////////////block for cpuprofile///////////////
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		//		defer pprof.StopCPUProfile()
		defer f.Close()
	}

	http.HandleFunc("/", LazyServer)
	http.ListenAndServe(":1112", nil)

}

// sometimes really fast server, sometimes really slow server
func LazyServer(w http.ResponseWriter, req *http.Request) {
	//Рондомизируем генератор случайных чисел

	countTime++

	////////////////////////////////block for memprofile///////////////
	if countTime == 30000 {

		if *memprofile != "" {
			fm, err := os.Create(*memprofile)
			if err != nil {
				log.Fatal("could not create memory profile: ", err)
			}
			runtime.GC() // get up-to-date statistics
			if err := pprof.WriteHeapProfile(fm); err != nil {
				log.Fatal("could not write memory profile: ", err)
			}
			fm.Close()
			pprof.StopCPUProfile()
			fmt.Println("Save memprofile")
			if *cpuprofile != "" {
				f.Close()
			}

		}
	}

	rand.Seed(time.Now().UnixNano())
	headOrTails := rand.Intn(2)

	if headOrTails == 0 {
		time.Sleep(7 * time.Second)
		fmt.Fprintf(w, "МЕДЛЕННЫЙ ОТВЕТ Go! slow %v", headOrTails)
		fmt.Printf("МЕДЛЕННЫЙ ОТВЕТ Go! slow %v \n", headOrTails)
		return
	}

	fmt.Fprintf(w, "БЫСТРЫЙ ОТВЕТ Go! quick %v ", headOrTails)
	fmt.Printf("БЫСТРЫЙ ОТВЕТ Go! quick %v \n", headOrTails)

	return
}

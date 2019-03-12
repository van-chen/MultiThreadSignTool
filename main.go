package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func main() {
	_cmd := flag.String("exec", "", "Program to process files (Required)")
	_path := flag.String("dir", "", "Directory to search files (Required)")
	_mask := flag.String("mask", "", "Files masks {.dll,.exe,.sys} (Required)")
	_max := flag.Int("max", 100, "Number of simultaneously running programs")
	_tryCount := flag.Int("trycount", 10, "Number of retrys to process file")
	flag.Parse()
	usage :=
		`Typical usage: -dir="C:\test" -mask=".exe,.dll" -exec="signtool.exe" sign  /q /ac "Mycert.cer" <filepath>
	Where <filepath> will be replaced by real file from -dir according -mask `

	if flag.NArg() == 0 || *_cmd == "" || *_path == "" || *_mask == "" {
		flag.Usage()
		fmt.Println(usage)
		os.Exit(1)
	}
	start := time.Now()

	var wg sync.WaitGroup

	canRun := make(chan bool, *_max)

	fmt.Println("Start processing", *_path)
	filepath.Walk(*_path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("Error:", err)
			return err
		}
		if info.IsDir() == false && strings.Contains(*_mask, filepath.Ext(path)) {

			wg.Add(1)
			go func(path string, info os.FileInfo, wg *sync.WaitGroup, args []string) {
				defer wg.Done()
				canRun <- true
				temp := make([]string, len(args))
				copy(temp, args)
				for i, s := range temp {
					temp[i] = strings.Replace(s, "<filepath>", path, 1)
				}

				fmt.Printf("Start %s\n", path)
				for i := 1; i <= *_tryCount; i++ {
					if out, cmderr := exec.Command(*_cmd, temp...).CombinedOutput(); cmderr != nil {
						if i >= *_tryCount {
							fmt.Printf("Error %s \n%s\n", path, string(out))
						} else {
							fmt.Printf("Retry (%d) %s\n", i, path)
						}
					} else {
						fmt.Printf("Success %s \n%s\n", path, string(out))
						break
					}
				}

				<-canRun
			}(path, info, &wg, flag.Args())

		}
		return nil
	})

	wg.Wait()

	fmt.Println("Processing finished. Time ", time.Since(start))
}

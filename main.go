package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type Download struct {
	Url string
	TargetPath string
	TotalSection int
}
func main()  {
	startTime := time.Now()
	d := Download{
		Url: "https://www.dropbox.com/s/3uk5djzrr008lkc/1.%20Why%20choose%20Go%20%28%20golang%20%29%20for%20web%20development.mp4?dl=1",
		TargetPath: "final.mp4",
		TotalSection: 10,
	}
	err := d.Do()
	if err != nil{
		log.Fatalf("An Error Occured!")
	}
	fmt.Printf("Download Completed in %v seconds\n", time.Now().Sub(startTime).Seconds())
}

func (d Download) Do() error  {
	fmt.Println("Making Connection ....")
	r, err := d.getNewRequest("HEAD")
	if err != nil {
		return err
	}
	res, err := http.DefaultClient.Do(r)
	if err != nil {
		return err
	}
	fmt.Printf("Got %v\n", res.StatusCode)

	if res.StatusCode > 299 {
		return errors.New(fmt.Sprintf("Can't process, response is %v", res.StatusCode))
	}
	size, err := strconv.Atoi(res.Header.Get("Content-Length"))
	if err != nil {
		fmt.Printf("Size is %v bytes\n", size)
	}
	var section = make([][2]int, d.TotalSection)

	eachSize := size / d.TotalSection
	fmt.Printf("Each size is %v bytes\n", eachSize)
	//fmt.Println(section)

	for i := range section{
		if i == 0 {
			section[i][0] = 0
		} else {
			section[i][0] = section[i-1][1] + 1
		}

		if i < d.TotalSection-1 {
			section[i][1] = section[i][0] + eachSize
		} else {
			section[i][1] = size - 1
		}
	}
	fmt.Println(section)
	var waitGroup sync.WaitGroup
	for i, s := range section{
		waitGroup.Add(1)
		i := i
		s := s
		go func(){
			defer waitGroup.Done()
			err = d.downloadSection(i, s)
			if err !=nil {
				panic(err)
			}
		}()

	}
	waitGroup.Wait()
	err = d.mergeFiles(section)
	if err != nil {
		return err
	}
	return nil
}

func (d Download) getNewRequest(method string) (*http.Request, error)  {
	r, err := http.NewRequest(
		method,
		d.Url,
		nil,
		)
	if err != nil {
		r.Header.Set("User-Agent", "Download Manager")
	}

	return r, nil
}
func (d Download) downloadSection(i int, s [2]int) error  {
	r, err:= d.getNewRequest("GET")
	if err !=nil {
		return err
	}
	r.Header.Set("Range", fmt.Sprintf("bytes=%v-%v", s[0], s[1]))
	res, err := http.DefaultClient.Do(r)
	if err !=nil {
		return err
	}
	fmt.Printf("Downloaded %v bytes for section %v: %v\n", res.Header.Get("Content-Length"), i, s)
	b, err := ioutil.ReadAll(res.Body)
	if err !=nil {
		return err
	}
	err = ioutil.WriteFile(fmt.Sprintf("Section-%v.tmp", i), b, os.ModePerm)
	if err !=nil {
		return err
	}
	return nil
}

func (d Download) mergeFiles(section [][2]int) error  {
	f, err := os.OpenFile(d.TargetPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, os.ModePerm)
	if err !=nil {
		return err
	}
	defer f.Close()

	for i := range section{
		b, err := ioutil.ReadFile(fmt.Sprintf("section-%v.tmp", i))
		if err !=nil {
			return err
		}
		n, err := f.Write(b)
		fmt.Printf("%v bytes merged\n", n)
	}

	return nil
}
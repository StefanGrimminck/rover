package main

import (
	"bufio"
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Websites struct {
	Domain  string      `json:"domain"`
	Status  string		`json:"status"`
	Content string      `json:"body"`

}


var TIMEOUT = 60
var SUCCESS, FAIL, NOHOST, TIMEDOUT float64

func MakeRequest(url string,wg *sync.WaitGroup) {
	defer wg.Done()

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		MaxIdleConnsPerHost: 10,
		TLSHandshakeTimeout: time.Duration(TIMEOUT * 1000000000),
	}

	timeout := time.Duration(TIMEOUT * 1000000000)
	client := http.Client{
		Timeout: timeout,
		Transport: tr,
	}
	var err error
	var resp *http.Response

	if  strings.Contains(url, "http://") ||  strings.Contains(url, "https://") || strings.Contains(url, "www."){
		resp, err = client.Get(url)

	}else{
		resp, err = client.Get("http://www." + url)

	}
	if err == nil {
		SUCCESS++

		body, _ := ioutil.ReadAll(resp.Body)

		file, _ := json.Marshal(
			Websites{
				Domain:url,
				Status: resp.Status,
				Content: string(body),

			}, )

	//	fmt.Printf("%v\n", string(file))
		Use(file)



	} else if strings.Contains(string(err.Error()), "no such host"){
		NOHOST ++
	} else if strings.Contains(string(err.Error()),"request canceled while waiting") || strings.Contains(string(err.Error()),"Timeout"){
		TIMEDOUT ++
	}else{
		FAIL ++
		fmt.Println(err)
	}

}


func main() {

	start := time.Now()

	csvFile, _ := os.Open("websites.csv")
	reader := csv.NewReader(bufio.NewReader(csvFile))

	var wg sync.WaitGroup
	var sites []string


	for {
		domain, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}
		sites = append(sites, domain[0])

	}

	wg.Add(len(sites))

	for _, element := range sites {
		if len(element) > 0 {
			go MakeRequest(element, &wg)
		}
	}

	wg.Wait()

	elapsed := time.Since(start)
	fmt.Printf("Rover took: %s\n", elapsed)
	fmt.Printf("TOTAL:   \t%v\n",  len(sites))
	fmt.Printf("SUCCESS: \t%v, \t%.1f%% of total\n", SUCCESS, SUCCESS/float64(len(sites))*100)
	fmt.Printf("FAILED:  \t%v, \t%.1f%% of total\n", FAIL, FAIL/float64(len(sites))*100)
	fmt.Printf("NO HOST: \t%v, \t%.1f%% of total\n", NOHOST, NOHOST/float64(len(sites))*100)
	fmt.Printf("TIMED OUT: \t%v, \t%.1f%% of total\n", TIMEDOUT, TIMEDOUT/float64(len(sites))*100)



}

func Use(vals ...interface{}) {
	for _, val := range vals {
		_ = val
	}
}
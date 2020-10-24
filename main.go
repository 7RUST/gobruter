package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

var transport = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	DialContext: (&net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: time.Second / 2,
		DualStack: true,
	}).DialContext,
}

var httpClient = &http.Client{
	Transport: transport,
}

func main() {
	//Setup variables
	var wg sync.WaitGroup
	urls := make(chan string)
	dd := flag.String("d", "", "Specify domain")
	wordlist := flag.String("w", "", "Specify wordlist to use")
	threads := flag.Int("t", 20, "Specify threads to run")
	flag.Parse()
	domain := *dd
	if domain[len(domain)-1] == 47 {
		domain = domain[:len(domain)-1]
	}
	if domain == "" {
		fmt.Println("Please Specify a domain")
		os.Exit(0)
	}
	if *wordlist == "" {
		fmt.Println("Please Specify a wordlist")
		os.Exit(0)
	}
	if _, err := os.Stat(*wordlist); err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Wordlist dont exist")
			os.Exit(0)
		}
	}
	//Check if domain has protocol
	parsed, err := url.Parse(domain)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	if len(parsed.Scheme) == 0 {
		fmt.Println("Invalid url format. Try adding http or https at the beggining")
		os.Exit(0)
	}

	//Setup workers
	for i := 0; i < *threads; i++ {
		wg.Add(1)
		go workers(urls, &wg)
	}

	//Open File
	file, _ := os.Open(*wordlist)
	paths := bufio.NewScanner(file)

	for paths.Scan() {
		if paths.Text()[0] == 47 {
			urls <- domain + paths.Text()
		} else {
			urls <- domain + "/" + paths.Text()
		}
	}
	close(urls)
	wg.Wait()
}

func workers(cha chan string, wg *sync.WaitGroup) {
	for i := range cha {
		checkifalive(i)
	}
	wg.Done()
}

func checkifalive(s string) {
	resp, err := httpClient.Get(s)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		fmt.Println(s)
		return
	}
	return
}

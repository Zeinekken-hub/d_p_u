package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var notFoundBreeds []string

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	fmt.Print("input: ")
	text, err := bufio.NewReader(os.Stdin).ReadString('\n')
	checkError(err)
	data := strings.Split(text, ",")
	notFoundBreeds = make([]string, 0, len(data))
	url := "https://www.akc.org/dog-breeds/"
	start := time.Now()

	wg := &sync.WaitGroup{}

	for _, elem := range data {
		elem = strings.TrimSpace(elem)
		if elem == "" {
			continue
		}
		elem = strings.ToLower(elem)
		elem = strings.ReplaceAll(elem, "_", "-")
		elem = elem[1 : len(elem)-1]
		urlToGo := url + elem + "/"
		createFolder(elem)

		wg.Add(1)
		go parseImages(urlToGo, elem, elem, wg)

		fmt.Println()
	}

	wg.Wait()
	elapsed := time.Since(start)
	if len(notFoundBreeds) != 0 {
		log.Printf("Not found breeds %v\n", notFoundBreeds)
	}
	log.Printf("Execution took %s\n", elapsed)
	fmt.Println("Type any key to exit")
	fmt.Scanln()
}

func parseImages(url, elem, folderName string, wg1 *sync.WaitGroup) {
	defer wg1.Done()
	fmt.Printf("going to url %v\n", url)
	response, err := http.Get(url)
	checkError(err)
	defer response.Body.Close()

	document, err := goquery.NewDocumentFromReader(response.Body)
	checkError(err)

	wg := &sync.WaitGroup{}
	document.Find("img").Each(func(index int, element *goquery.Selection) {
		imgSrc, exists := element.Attr("data-src")
		if exists {
			if strings.Contains(strings.ToLower(imgSrc), elem) {
				wg.Add(1)
				go getImage(imgSrc, folderName, wg) // go
			}
		}
	})

	wg.Wait()

	ok, err := isEmpty(folderName)
	checkError(err)
	if ok {
		notFoundBreeds = append(notFoundBreeds, folderName)
		fmt.Println("Remove folder cause it empty")
		os.Remove(folderName)
	}
}

func getImage(urlImage, folder string, wg *sync.WaitGroup) {
	defer wg.Done()

	resp, err := http.Get(urlImage)
	checkError(err)
	defer resp.Body.Close()

	fileURL, err := url.Parse(urlImage)
	checkError(err)
	path := strings.Split(fileURL.Path, "/")
	name := path[len(path)-1]

	createImage(name, folder, resp)
}

func createImage(name, folder string, resp *http.Response) {
	file, err := os.Create(folder + "/" + name) // create in new folder
	checkError(err)
	size, err := io.Copy(file, resp.Body)
	checkError(err)
	defer file.Close()

	fmt.Printf("image %s was created with size %db\n", name, size)
}

func createFolder(name string) {
	_, err := os.Stat(name)
	fmt.Printf("Check if folder %v exists\n", name)

	if os.IsNotExist(err) {
		fmt.Printf("Try to create folder %v\n", name)
		err := os.MkdirAll(name, 0755)
		checkError(err)
	}
}

func isEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

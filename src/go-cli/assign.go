package main

import(
    "fmt"
    "flag"
	"net/url"
	"log"
	"net"
	"io/ioutil"
	"time"
	"unsafe"
	"strings"
	"bufio"
)


func showHelp() {
    fmt.Println(`
go-cli:  command to make http request to the requested url and print the response.

Usage: go-cli -h
Usage: go-cli --url
Usage: go-cli --profile

Options:
    -u, --url=url     	     Flag to input the url.
    -p, --profile=num    number of req make to url and return the stats
    -h, --help           prints this help info.
    `)
}

func findStatusCode(resp string) string{

	scanner := bufio.NewScanner(strings.NewReader(resp))
	scanner.Scan()
	line := scanner.Text()
	words := strings.Fields(line)
	return words[1]

}

func printResponseBody(response string){
	scanner := bufio.NewScanner(strings.NewReader(response))
	isBodyStart := false
	for scanner.Scan(){
		line :=scanner.Text()
		if line == ""{
			isBodyStart = true
		}


		if isBodyStart{
			fmt.Println(line)
		}
	}

}

func findMedianTime(timeTaken[] float64) float64{
	length := len(timeTaken)
	if length%2 != 0 {
		return timeTaken[length/2]
	} else {
		return (timeTaken[length/2]+timeTaken[(length-1)/2])/2
	}
} 

func setFlag(flag *flag.FlagSet) {
    flag.Usage = func() {
        showHelp()
    }
}

func main() {

	maxTime := -1000000.0
	minTime := 1000000.0
	maxSize := -1000000
	minSize := 1000000
	var timeTaken[] float64
	var codes[] string
	var urlString string
	var profile float64
	var sHelp bool
	profileFlagSet := false


	flag.StringVar(&urlString, "u", "", "")
	flag.StringVar(&urlString, "url", "", "")
	
	
    flag.Float64Var(&profile, "p", 0, "")
    flag.Float64Var(&profile, "profile", 0, "")
    
    flag.BoolVar(&sHelp, "h", false, "")
    flag.BoolVar(&sHelp, "help", false, "")

    setFlag(flag.CommandLine)

    flag.Parse()

    if sHelp {
       showHelp()
        return
    }

	if profile > 0 {
		profileFlagSet = true
	} else {
		profile = profile+1
	}


	if urlString != "" {

		u, err := url.Parse(urlString)
		if err != nil {
			log.Fatal(err)
		}
		
		for i:=0 ; i < int(profile) ; i++ {

			conn, err := net.Dial("tcp", u.Host+":80")
    		if err != nil {
        		log.Fatal(err)
			}
		
			start := time.Now()
			rt := fmt.Sprintf("GET %v HTTP/1.1\r\n", u.Path)
			rt += fmt.Sprintf("Host: %v\r\n", u.Host)
			rt += fmt.Sprintf("Connection: close\r\n")
			rt += fmt.Sprintf("\r\n")

			_, err = conn.Write([]byte(rt))
			if err != nil {
				log.Fatal(err)
			}

			resp, err := ioutil.ReadAll(conn)
			if err != nil {
				log.Fatal(err)
			}
			
			responseSize := int(unsafe.Sizeof(resp))

			if responseSize < minSize {
				minSize = responseSize
			}

			if responseSize > maxSize {
				maxSize = responseSize
			}


			response := string(resp)

			if profileFlagSet == false{
				printResponseBody(response)
			}

			
			statusCode :=findStatusCode(response)
			codes = append(codes,statusCode)

			elapsed := time.Since(start).Seconds()
			timeTaken = append(timeTaken,elapsed)

			if elapsed < minTime {
				minTime = elapsed
			}

			if elapsed > maxTime {
				maxTime = elapsed
			}

		    	conn.Close()

		}
	}

	if profileFlagSet {

		fmt.Println("number of request made ", profile)
		fmt.Println("fastest time ", minTime)
		fmt.Println("Slowest time ", maxTime)
		
		totalTime := 0.00
		successReq := 0.00
		var errorCodes[] string
		for i :=0;i < int (profile);i++ {
			totalTime += timeTaken[i]
			if codes[i][0] == '2' {
				successReq = successReq+1
			} else { 
				errorCodes = append(errorCodes, codes[i])
			}
		}

		fmt.Println("Mean Time ", totalTime/profile)
		fmt.Println("Median Time ", findMedianTime(timeTaken))
		fmt.Println("Percentage of success Request ", successReq/profile*100)
		fmt.Println("Error Codes if any:")
		for i := range errorCodes{
			fmt.Println(errorCodes[i])
		}
		fmt.Println("Minmum Response Size ",minSize)
		fmt.Println("Maximum Response Size ",maxSize)
	}

}
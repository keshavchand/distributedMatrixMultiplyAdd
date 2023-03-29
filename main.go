package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type matrix struct {
	r    int
	c    int
	nums []int
}

type Client struct {
	net.Conn
	l *sync.Mutex
}

func (c *Client) Printf(format string, a ...any) {
	c.l.Lock()
	defer c.l.Unlock()
	fmt.Fprintf(c.Conn, format, a...)
}

func registerToServer(out chan struct{}, serverUrl string, selfUrl string) {
	conn, err := net.Dial("tcp", serverUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	fmt.Fprintf(conn, "%s\n", selfUrl)
	close(out)

	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for _ = range t.C {
		_, err := fmt.Fprintf(conn, "HB\n")
		if err != nil {
			break
		}
	}

	// XXX: What to do if master goes belly up
}

func main() {
	serverUrl := "localhost:7080"
	selfUrl := "localhost:8080"
	const standalone = false

	conn, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if !standalone {
		done := make(chan struct{})
		go registerToServer(done, serverUrl, selfUrl)
		<-done
	}

	var RequestId uint64

	var mapLock sync.Mutex
	processes := make(map[uint64]MatrixResult)

	for {
		_client, err := conn.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		client := &Client{Conn: _client, l: &sync.Mutex{}}

		go func(conn *Client) {
			defer client.Close()

			scanner := bufio.NewScanner(client)
			// Request Format:
			// <request>
			for scanner.Scan() {
				req := scanner.Text()
				switch req {
				case "matstore":
					A, err := getMatrix(scanner)
					if err != nil {
						client.Printf("NOK: %v\n", err)
						continue
					}

					reqId := atomic.AddUint64(&RequestId, 1) - 1
					mapLock.Lock()
					processes[uint64(reqId)] = &MatStoreReq{A: A}
					mapLock.Unlock()

					client.Printf("%d\n", reqId)

				case "fetchres":
					if !scanner.Scan() {
						log.Println("Can't read data from client")
						continue
					}

					reqId, err := strconv.Atoi(scanner.Text())
					if err != nil {
						client.Printf("NOK: Not available\n")
						continue
					}

					mapLock.Lock()
					req, ok := processes[uint64(reqId)]
					mapLock.Unlock()
					if !ok {
						client.Printf("NOK: No Id %d present\n", reqId)
						continue
					}

					if !req.Completed() {
						client.Printf("NOK: In Processing\n")
						continue
					}

					result, err := req.Result()
					// TODO: sepereate function function
					func() {
						client.l.Lock()
						defer client.l.Unlock()
						writeOutputToClient(result, client)
					}()

				case "matadd":
					if !scanner.Scan() {
						log.Println("Can't read data from client")
						continue
					}

					reqId, err := strconv.Atoi(scanner.Text())
					if err != nil {
						client.Printf("NOK: Not available\n")
						continue
					}

					mapLock.Lock()
					req, ok := processes[uint64(reqId)]
					mapLock.Unlock()
					if !ok {
						client.Printf("NOK: No Id %d present\n", reqId)
						continue
					}

					if !req.Completed() {
						client.Printf("NOK: In Processing\n")
						continue
					}

					reqMatrix, err := req.Result()
					if err != nil {
						client.Printf("NOK: Result Matrix %v\n", err)
						continue
					}

					matrix, err := getMatrixes(scanner)
					if err != nil {
						client.Printf("NOK: Error getting matrics %v\n", err)
						continue
					}

					new_reqId := atomic.AddUint64(&RequestId, 1) - 1
					matAddReq := &MatAddReq{Res: reqMatrix, A: matrix}
					go matAddReq.Process()

					mapLock.Lock()
					processes[uint64(new_reqId)] = matAddReq
					mapLock.Unlock()
					client.Printf("%d\n", reqId)

				case "matmuladd":
					A, err := getMatrixes(scanner)
					if err != nil {
						log.Println("error matrix multiply add: ", err)
						continue
					}
					B, err := getMatrixes(scanner)
					if err != nil {
						log.Println("error matrix multiply add: ", err)
						continue
					}

					reqId := atomic.AddUint64(&RequestId, 1) - 1
					req := &MatMulAddReq{ReqId: reqId, A: A, B: B}
					go req.Process()

					mapLock.Lock()
					processes[reqId] = req
					mapLock.Unlock()

					client.Printf("%d\n", reqId)
				default:
					log.Println("Not Implemented ", req)
				}
			}
		}(client)
	}
}

func writeOutputToClient(result matrix, client net.Conn) {
	resp := bufio.NewWriter(client)
	defer resp.Flush()
	fmt.Fprintf(resp, "%d %d\n", result.r, result.c)
	for idx, i := range result.nums {
		if idx != 0 {
			resp.Write([]byte(" "))
		}
		fmt.Fprintf(resp, "%d", i)
	}
	resp.Write([]byte("\n"))
}

func getMatrix(scanner *bufio.Scanner) (matrix, error) {
	// r, c : <rows, cols>
	// A1 ...
	var A matrix
	if !scanner.Scan() {
		log.Println("No Item in scanner")
		return A, fmt.Errorf("no item in scanner")
	}

	if !scanner.Scan() {
		log.Println("No Data in Scanner")
		return A, fmt.Errorf("No data in scanner")
	}
	r1, c1, found := strings.Cut(scanner.Text(), " ")
	if !found {
		log.Println("Not found")
		return A, fmt.Errorf("Error parsing sizes")
	}

	r, _ := strconv.Atoi(r1)
	c, _ := strconv.Atoi(c1)
	log.Println("Getting Matrix:", r, c)

	if !scanner.Scan() {
		log.Println("No Data in Scanner")
		return A, fmt.Errorf("No data in scanner")
	}
	numbers := strings.Split(scanner.Text(), " ")
	if len(numbers) != r*c {
		log.Println("Length doesn't match the given amount")
		return A, fmt.Errorf("Length doesn't match the given amount")
	}

	nums := make([]int, r*c)
	for i, n := range numbers {
		var err error
		nums[i], err = strconv.Atoi(n)
		if err != nil {
			log.Println("parsing numbers: ", err)
			return A, fmt.Errorf("Parsing Number")
		}
	}

	A = matrix{r: r, c: c, nums: nums}
	return A, nil
}

func getMatrixes(scanner *bufio.Scanner) ([]matrix, error) {
	// n : <number of matricies>
	// r, c : <rows, cols>
	// A1 ...
	// r, c : <rows, cols>
	// A2 ...
	// r, c : <rows, cols>
	// An ...
	if !scanner.Scan() {
		log.Println("No Item in scanner")
		return nil, fmt.Errorf("no item in scanner")
	}

	numberOfMatrices, err := strconv.Atoi(scanner.Text())
	// TODO: Notify client of the err
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("matrix mul add %v", err)
	}

	A := make([]matrix, numberOfMatrices)
	for i := 0; i < numberOfMatrices; i++ {
		if !scanner.Scan() {
			log.Println("No Data in Scanner")
			return nil, fmt.Errorf("No data in scanner")
		}
		r1, c1, found := strings.Cut(scanner.Text(), " ")
		if !found {
			log.Println("Not found")
			return nil, fmt.Errorf("Error parsing sizes")
		}

		r, _ := strconv.Atoi(r1)
		c, _ := strconv.Atoi(c1)
		log.Println("Getting Matrix:", r, c)

		if !scanner.Scan() {
			log.Println("No Data in Scanner")
			return nil, fmt.Errorf("No data in scanner")
		}
		numbers := strings.Split(scanner.Text(), " ")
		if len(numbers) != r*c {
			log.Println("Length doesn't match the given amount")
			return nil, fmt.Errorf("Length doesn't match the given amount")
		}

		nums := make([]int, r*c)
		for i, n := range numbers {
			nums[i], err = strconv.Atoi(n)
			if err != nil {
				log.Println("parsing numbers: ", err)
				return nil, fmt.Errorf("Parsing Number")
			}
		}

		A[i] = matrix{r: r, c: c, nums: nums}
	}
	return A, nil
}

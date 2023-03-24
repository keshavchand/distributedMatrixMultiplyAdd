package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
)

type matrix struct {
	r    int
	c    int
	nums []int
}

func main() {
	conn, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal(err)
	}

	clientId := 0
	for {
		client, err := conn.Accept()
		if err != nil {
			log.Println(err)
			continue
		}

		clientId += 1
		go func(conn net.Conn) {
			defer client.Close()
			var W sync.Mutex

			scanner := bufio.NewScanner(client)
			// Request Format :
			// <request>
			for scanner.Scan() {
				req := scanner.Text()
				switch req {
				case "matmuladd":
					A, B, err := getMatrixForMultiplyAdd(scanner)
					if err != nil {
						log.Println("error matrix multiply add: ", err)
						continue
					}
					// Returns:
					// <rows, cols>
					// A1*B1 + A2*B2 + ... + An*Bn
					// TODO: run in goroutine
					result, err := MatrixMultiplyAndAdd(A, B)

					func() {
						W.Lock()
						defer W.Unlock()
						resp := bufio.NewWriter(client)

						defer resp.Flush()
						fmt.Fprintf(resp, "%d %d\n", result.r, result.c)
						for idx, i := range result.nums {
							if idx != 0 {
								resp.Write([]byte(" "))
							}
							fmt.Fprintf(resp, "%d", i)
						}
					}()
				default:
					log.Println("Not Implemented ", req)
				}
			}
		}(client)
	}
}

func getMatrixForMultiplyAdd(scanner *bufio.Scanner) ([]matrix, []matrix, error) {
	// n : <number of matricies>
	// r, c : <rows, cols>
	// A1 ...
	// r, c : <rows, cols>
	// A2 ...
	// r, c : <rows, cols>
	// An ...
	//
	// r, c : <rows, cols>
	// B1 ...
	// r, c : <rows, cols>
	// B2 ...
	// r, c : <rows, cols>
	// Bn ...
	//
	if !scanner.Scan() {
		log.Println("No Item in scanner")
		return nil, nil, fmt.Errorf("no item in scanner")
	}

	numberOfMatrices, err := strconv.Atoi(scanner.Text())
	// TODO: Notify client of the err
	if err != nil {
		log.Println(err)
		return nil, nil, fmt.Errorf("matrix mul add %v", err)
	}

	A := make([]matrix, numberOfMatrices)
	B := make([]matrix, numberOfMatrices)

	for _, Mat := range [][]matrix{A, B} {

		for i := 0; i < numberOfMatrices; i++ {
			if !scanner.Scan() {
				log.Println("No Data in Scanner")
				return nil, nil, fmt.Errorf("No data in scanner")
			}
			r1, c1, found := strings.Cut(scanner.Text(), " ")
			if !found {
				log.Println("Not found")
				return nil, nil, fmt.Errorf("Error parsing sizes")
			}

			r, _ := strconv.Atoi(r1)
			c, _ := strconv.Atoi(c1)
			log.Println("Getting Matrix:", r, c)

			if !scanner.Scan() {
				log.Println("No Data in Scanner")
				return nil, nil, fmt.Errorf("No data in scanner")
			}
			numbers := strings.Split(scanner.Text(), " ")
			if len(numbers) != r*c {
				log.Println("Length doesn't match the given amount")
				return nil, nil, fmt.Errorf("Length doesn't match the given amount")
			}

			nums := make([]int, r*c)
			for i, n := range numbers {
				nums[i], err = strconv.Atoi(n)
				if err != nil {
					log.Println("parsing numbers: ", err)
					return nil, nil, fmt.Errorf("Parsing Number")
				}
			}

			Mat[i] = matrix{r: r, c: c, nums: nums}
		}
	}
	return A, B, nil
}

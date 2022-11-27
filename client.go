package main
import (
	"os"
	"fmt"
	"sync"
	"strconv"
)

func createNode(i int, val float64, m sync.Map) {

	//m.Store(i, )
	fmt.Printf("%d, %f\n", i, val)
}
//func simulateDelay()
func main() {
	//TODO: ERROR CHECKING


	//run using
	//go run client.go [server to connect to] [N number of nodes] [number of faults] [list of N values between 0 and 1]
	args := os.Args
	//serverNo := args[1]

	N := args[2]
	f := args[3]
	values := args[4:]
	nodes := sync.Map{}

	for i, val := range values {
		value, err := strconv.ParseFloat(val,32)
		//fmt.Println(val)
		if err != nil{
			fmt.Printf("The %dth value that you supplied was not a float, please rerun the program", i+1)
			
		}
		createNode(i, value, nodes)
	}
	fmt.Println(N,f,values)
}

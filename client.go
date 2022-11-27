package main
import (
	"os"
	"fmt"
	"sync"
	"strconv"
)
type Map struct {
	syncMap sync.Map
}
func (m *Map) Load(i int) chan interface{} {
	val, ok := m.syncMap.Load(i)
	if ok {
		return val.(chan interface{})
	} else {
		return nil
	}
}
func (m *Map) Store(i int, value chan interface{}) {
	m.syncMap.Store(i, value)
}
func (m *Map) Exists(i int) bool {
	_, ok := m.syncMap.Load(i)
	return ok
}
func (m *Map) Delete(i int) {
	m.syncMap.Delete(i)
}
var nodes Map
type channel chan interface{float64}


func createNode(i int, val float64, m Map, waitsFor int) {
	dataChannel := make(channel, waitsFor)
	m.Store(i, dataChannel)
	fmt.Printf("%d, %f\n", i, val)
}


func findConcensus(i int, m Map, N int){
	for j:= 0; j < N; {
		channel := m.Load(j)
		if channel == nil{
			fmt.Println("channel loaded incorrectly")
		}
		channel <- i
	}
}
//TODO
//func simulateDelay()
func main() {
	//TODO: ERROR CHECKING


	//run using
	//go run client.go [server to connect to] [N number of nodes] [number of faults] [list of N values between 0 and 1]
	args := os.Args
	//serverNo := args[1]

	N, err:= strconv.Atoi(args[2])
	if err != nil {
		fmt.Println("The second value that you supplied was not an integer, please rerun the program")
	}
	f, err:= strconv.Atoi(args[3])
	if err != nil {
		fmt.Println("The third value that you supplied was not an integer, please rerun the program")
	}

	values := args[4:]
	if len(values) != N{
		fmt.Println("The length of the value list is not the same as the input N, please rerun the program. ")
	}
	

	for i, val := range values {
		value, err := strconv.ParseFloat(val,32)
		//fmt.Println(val)
		if err != nil{
			fmt.Printf("The %dth value that you supplied was not a float, please rerun the program\n", i+1)
			
		}

		createNode(i, value, nodes, N-f)
	}
	for i, _ := range values {
		go findConcensus(i)
	}
	fmt.Println(N,f,values)
}

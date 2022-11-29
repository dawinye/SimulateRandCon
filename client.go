package main
import (
	"fmt"
	"os"
	"strconv"
	"math"
	"net"
	"errors"
	//"reflect"
	"bytes"
	"encoding/gob"
	"sync"
	"time"
)
type Map struct {
	syncMap sync.Map
}


type Data struct {
	Val float64
	Round int
}

var errUnexpectedType = errors.New("Non-numeric type could not be converted to float")
func getFloatSwitchOnly(unk interface{}) (float64, error) {
	switch i := unk.(type) {
	case float64:
		return i, nil
	case float32:
		return float64(i), nil
	case int64:
		return float64(i), nil
	case int32:
		return float64(i), nil
	case int:
		return float64(i), nil
	case uint64:
		return float64(i), nil
	case uint32:
		return float64(i), nil
	case uint:
		return float64(i), nil
	default:
		return math.NaN(), errUnexpectedType
	}
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


//rounds is a regular map that will be accessed using mutexes because it doesnt
//fall under sync.Map's use cases
var rounds = make(map[int]int)
//nodes is a custom struct of type Map which makes it have channels as values
var nodes Map
//mutex to access values of rounds
var mutex = &sync.RWMutex{}
//i have no clue what this means
type channel chan interface{}


//initalize each node's channel and entries in both maps
func createNode(i int, val float64, waitsFor int) {

	dataChannel := make(channel, waitsFor)

	nodes.Store(i, dataChannel)
	mutex.Lock()
	rounds[i] = 1
	mutex.Unlock()
}
//send one node's value to another one
func sendValue(self int, to int, round int, val float64, c net.Conn){
	fmt.Println(self, to, round, val)

	if self == to {
		d := Data{val, round}
		//fmt.Println(d)
		var binBuf = new(bytes.Buffer)
		var gobobj = gob.NewEncoder(binBuf)
		gobobj.Encode(d)

		_, err  := c.Write(binBuf.Bytes())
		if err != nil{
			fmt.Println(err)
		}
		return

	}
	time.Sleep(1*time.Second)
	for {
		mutex.RLock()
		toRound, _ := rounds[to]
		mutex.RUnlock()
		if toRound == round {
			channel := nodes.Load(to)
			channel <- val
			break
		}
	}
	//fmt.Println("Sent ", val, " to ", to)
}


func findConsensus(i int, N int, f int, r int, initVal float64, c net.Conn) {
	//according to randomized to the consensus algorithm we learned,
	//one of the nodes that are included in n-f must be itself
	//therefore, we send that value first and then skip it in the for loop
	//d := data{initVal,r}
	self := nodes.Load(i)
	self <- initVal
	for j:= 0; j < N; j++{
		go sendValue(i, j, r, initVal,c)
	}
	time.Sleep(1*time.Second)
	//infinite for loop to see if N-f messages are in the channel,
	var sum float64
	for {
		if len(self) == cap(self) {
			for j := 0; j < N-f; j++ {
				newVal := <- self
				v, _ := getFloatSwitchOnly(newVal)
				sum += v
			}
			total := float64(N-f)
			sum /= total
			//fmt.Println(sum)
			break
		}
	}
	//lock the mutex to increase this node's round by 1, then unlock
	mutex.Lock()
	rounds[i] += 1
	mutex.Unlock()
	//fmt.Println(rounds)
	fmt.Printf("New average for node %d is : %f\n",i,  sum)
	//run recursively to begin next round

	findConsensus(i, N, f, r+1, sum,c)
}
//TODO
//func simulateDelay()
func main() {
	//TODO: ERROR CHECKING

	
	//run using
	//go run client.go [server to connect to] [N number of nodes] [number of faults] [list of N values between 0 and 1]
	args := os.Args
	port := args[1]
	c, err := net.Dial("tcp4", "127.0.0.1:"+port)
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
	
	//iterate through all the values to initialize the nodes
	for i, val := range values {
		value, err := strconv.ParseFloat(val,32)
		//fmt.Println(val)
		if err != nil{
			fmt.Printf("The %dth value that you supplied was not a float, please rerun the program\n", i+1)
			
		}

		createNode(i, value, N-f)
	}
	//iterate through all the values since now that the nodes have been created,
	//approx consensus can begin
	for i, val := range values {
		
		value, err := strconv.ParseFloat(val,32)
		if err != nil{
			fmt.Printf("The %dth value that you supplied was not a float, please rerun the program\n", i+1)
		}
		go findConsensus(i, N, f, 1, value,c)
	}
	//sleeping to give time for the go routines in the above line time to run, will replace with a signal from the server on when to stop
	time.Sleep(10 * time.Second)
}

package main
import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
	"reflect"
)
type Map struct {
	syncMap sync.Map
}


type data struct {
	val float64
	round int
}

func ChanToSlice(ch interface{}) interface{} {
	chv := reflect.ValueOf(ch)
	slv := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(ch).Elem()), 0, 0)
	for {
		v, ok := chv.Recv()
		if !ok {
			return slv.Interface()
		}
		slv = reflect.Append(slv, v)
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
	rounds[i] = 1
	fmt.Printf("%d, %f\n", i, val)
}
//send one node's value to another one
func sendValue(to int, round int, val float64){
	d := data{val, round}
	fmt.Println(d)
	for {
		mutex.RLock()
		toRound, _ := rounds[to]
		if toRound == round {
			channel := nodes.Load(to)
			channel <- d
			break
		}
		mutex.RUnlock()
	}
}


func findConsensus(i int, N int, f int, r int, initVal float64){
	fmt.Printf("here")
	//according to randomized to the consensus algorithm we learned,
	//one of the nodes that are included in n-f must be itself
	//therefore, we send that value first and then skip it in the for loop
	d := data{initVal,r}
	self := nodes.Load(i)
	self <- d
	fmt.Println(self)
	for j:= 0; j < N; {

		if j == i{
			continue
		}
		//making this a goroutine spawns an infinite amount of print statements for some reason
		//leaving it like this stops the program since its stuck on something
		sendValue(j, r, initVal)
	}

	//infinite for loop to see if N-f messages are in the channel,
	var newVal float64
	for {
		if cap(self) == N-f {
			//stole some code to convert channel interface to a slice interface of type data
			sl := ChanToSlice(self).([]data)

			//calculate the average
			for _, d := range sl {
				newVal += d.val
			}
			total := float64(N-f)
			newVal /= total

			break
		}
	}
	//lock the mutex to increase this node's round by 1, then unlock
	mutex.Lock()
	rounds[i] += 1
	mutex.Unlock()
	fmt.Printf("%f", newVal)

	//run recursively to begin next round
	findConsensus(i, N, f, r+1, newVal)

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
		go findConsensus(i, N, f, 1, value)
	}
	//sleeping to give time for the go routines in the above line time to run, will replace with a signal from the server on when to stop
	time.Sleep(10 * time.Second)
	fmt.Println(N,f,values)
}

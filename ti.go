package main
import "fmt"
func rec(i int) int {
	fmt.Println(i)
	return rec(i-2)+i
}
func main() {
	for i:= 0; i < 5; i++ {
		fmt.Println(i)
		go rec(10+i)
	}
}

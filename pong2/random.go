package pong

import (
	"fmt"
	"math/rand"
)





func Hero(){

min:=10
max:=30

for i:=0;i<max;i++{
pass:=rand.Intn(max-min)-min
fmt.Println(pass)
}


}
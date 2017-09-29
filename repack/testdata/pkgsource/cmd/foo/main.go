package main

import (
	"fmt"

	"example.com/user/pkgsource/util"
	"example.com/user/pkgsource/util/sub"
)

func main() {
	fmt.Println(util.Broadcast())
	fmt.Println(sub.Fireworks())
}

package pkgsource

import (
	"fmt"

	"example.com/user/pkgsource/util" // trailing comment
	subalias "example.com/user/pkgsource/util/sub"
)

func ROOT() {
	fmt.Println(util.Broadcast())
	fmt.Println(subalias.Fireworks())
}

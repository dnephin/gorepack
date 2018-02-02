package pkgsource // import "vanity.fake/newsy"

import (
	"fmt"

	"vanity.fake/newsy/util" // trailing comment
	subalias "vanity.fake/newsy/util/sub"
)

func ROOT() {
	fmt.Println(util.Broadcast())
	fmt.Println(subalias.Fireworks())
}

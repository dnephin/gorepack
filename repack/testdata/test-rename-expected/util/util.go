//Package util provides things for foo
package util

import "vanity.fake/newsy/util/sub"

func Broadcast() string {
	return "BROADCAST"
}

func Fire() string {
	return sub.Fireworks()
}

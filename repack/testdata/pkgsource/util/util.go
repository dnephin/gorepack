//Package util provides things for foo
package util

import "example.com/user/pkgsource/util/sub"

func Broadcast() string {
	return "BROADCAST"
}

func Fire() string {
	return sub.Fireworks()
}

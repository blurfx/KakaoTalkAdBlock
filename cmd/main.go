//go:build windows

package main

import "kakaotalkadblock/internal"
import _ "kakaotalkadblock/winres"

func main() {
	internal.Run()
}

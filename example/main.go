//go:build darwin

package main

/*
#cgo CFLAGS: -mmacosx-version-min=10.13 -ObjC -fobjc-arc
#cgo LDFLAGS: -mmacosx-version-min=10.13 -framework Cocoa

void objCMain() {
	//NSLog(@"%@", [Greeter greetCiccio]);
}
*/
import "C"

func main() {
	C.objCMain()
}

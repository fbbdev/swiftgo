// Copyright (c) 2024 Fabio Massaioli
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package main

/*
typedef const char cchar_t;
*/
import "C"

//export greetGo
func greetGo(person *C.cchar_t) *C.char {
	return C.CString("Hej " + C.GoString(person) + "!")
}

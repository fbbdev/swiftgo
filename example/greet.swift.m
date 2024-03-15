//go:build darwin
// Copyright (c) 2024 Fabio Massaioli
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

import SwiftGo.GreetObjC

func greet(person: String) -> String {
    return greetObjC(person)
}

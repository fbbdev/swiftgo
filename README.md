# swiftgo

_SwiftGo_ is a wrapper for the standard Go tool that adds support for embedded
Swift code when targeting darwin systems.

> NOTE: This software package is in an early stage of development
> and most features advertised below are not implemented yet.

## Usage

```
swiftgo <go command> [go arguments]
```

SwiftGo reads its configuration from the following environment variables:

  - `SWIFTGO_GOTOOL` - Go tool binary (default: `$GOROOT/bin/go`, `go`)
  - `SWIFTGO_SWIFTC` - Swift compiler binary (default: `xcrun swiftc`, `swiftc`)
  - `SWIFTGO_SWIFTFLAGS` - Swift compiler flags (default: `-g -O`)

If a package contains files with the extension `.swift.m` and the current build
context has `GOOS=darwin`, Swiftgo will compile them as Swift code instead of
Objective-C; otherwise, it will report an error.

The use of the alternative extension `.swift.m`, is necessary to have the Go tool recognize swift files and take them into account during incremental rebuilds.

SwiftGo finds all header files in the package with extension `.h` and makes
them available to Swift code as importable modules: for example, the directive

```swift
import SwiftGo.HEADER
```

imports all declarations contained in `HEADER.h` from the module directory,
if such a file is present. As a special case, the directive

```swift
import SwiftGo.CgoExports
```

imports all declarations — if any — exported by Cgo files in the package.

A Swift to Objective-C bridging header will also be generated automatically
and included in the build, so that all Objective-C code may refer to Swift
classes marked with `@objc`/`@objcMembers` attributes.

Finally, SwiftGo scans each Go package for additional C flags specified
by `#cgo CFLAGS` directives and forwards them to the Swift compiler as follows:

  - if the `-mmacosx-version-min=<VERSION>` flag is present, it is used to select a target for the Swift compiler;
  - every additional header/framework search path is passed on as a module/framework search path.

## License

MIT License

Copyright (c) 2024 Fabio Massaioli

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

## Borrowed code

This module embeds some code from the Go software distribution, licensed under the following terms:

Copyright (c) 2009 The Go Authors. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are
met:

   * Redistributions of source code must retain the above copyright
notice, this list of conditions and the following disclaimer.
   * Redistributions in binary form must reproduce the above
copyright notice, this list of conditions and the following disclaimer
in the documentation and/or other materials provided with the
distribution.
   * Neither the name of Google Inc. nor the names of its
contributors may be used to endorse or promote products derived from
this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
"AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
(INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

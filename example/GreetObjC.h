//go:build darwin
// Copyright (c) 2024 Fabio Massaioli
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

#pragma once

#ifdef __cplusplus
extern "C" {
#endif

#import <Foundation/Foundation.h>

extern NSString *greetObjC(NSString *person);

#ifdef __cplusplus
} // extern "C"
#endif

//go:build darwin

#include <stdlib.h>
#include "_cgo_export.h"

#include "GreetObjC.h"

NSString *greetObjC(NSString *person) {
	char *cresult = greetGo([person UTF8String]);
	NSString *result = @(cresult);
	free(cresult);
	return result;
}

int main() {
    NSLog(@"%@", [Greeter greetCiccio]);
    return 0;
}

#import <Foundation/Foundation.h>


int MoveToTrash(const char* path) {
    @autoreleasepool {
        NSString *nsPath = [NSString stringWithUTF8String:path];
        NSURL *url = [NSURL fileURLWithPath:nsPath];
        NSError *error = nil;
        
        // The macOS “Portal” API
		// It will handle sandbox permission, cross-partition operation
		// and generate .DS_Store recovery data.
        BOOL success = [[NSFileManager defaultManager] 
                        trashItemAtURL:url 
                        resultingItemURL:nil 
                        error:&error];
        
        return success ? 0 : 1;
    }
}
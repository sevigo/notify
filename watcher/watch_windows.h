#ifndef WIN_H_
#define WIN_H_

#include <stdlib.h>

void Setup();

void WatchDirectory(char *dir);

void StopWatching(char *dir);

// this is a call-back function from the go code
extern void goCallbackFileChange(char *path, char *file, int action);

#endif // WIN_H_
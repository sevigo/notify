#ifndef LIN_H_
#define LIN_H_

#include <stdlib.h>

void goCallbackFileChange(char* root, char* path, char* file, int action);
void WatchDirectory(char* root, char* dir);

#endif
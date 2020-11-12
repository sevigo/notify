#include "watch_linux.h"

#include <stdio.h>
#include <sys/inotify.h>
#include <limits.h>
#include <unistd.h>
#include <dirent.h>
#include <string.h>
#include <pthread.h>

#define BUF_LEN (10 * (sizeof(struct inotify_event) + 1024 + 1))

void WatchDirectory(char* root, char* dir) {
	int inotifyFd, wd, j;
  	char buf[BUF_LEN] __attribute__ ((aligned(8)));
  	ssize_t numRead;
  	char *p;
  	struct inotify_event *event;

  	inotifyFd = inotify_init();
  	if (inotifyFd == -1) {
		printf("[CGO] [ERROR] inotify_init()");
		exit(-1);
   	}

   	wd = inotify_add_watch(inotifyFd, dir, IN_CLOSE_WRITE | IN_DELETE);
   	if (wd == -1) {
		printf("[CGO] [ERROR] WatchDirectory(): inotify_add_watch()");
		exit(-1);
	}

  	printf("[CGO] [INFO] WatchDirectory(): watching %s", dir);
  	for (;;) {
    	numRead = read(inotifyFd, buf, BUF_LEN);
    	if (numRead == 0) {
			printf("[CGO] [ERROR]: read() from inotify fd returned 0!");
			exit(-1);
		}

    	if (numRead == -1) {
			printf("[CGO] [ERROR] read() from inotify fd returned -1");
			exit(-1);
		}

    	for (p = buf; p < buf + numRead; ) {
			event = (struct inotify_event *) p;
			printf("[CGO] [INFO] file was changed: %s %s\n", dir, event->name);
			goCallbackFileChange(root, dir, event->name, event->mask);
			p += sizeof(struct inotify_event) + event->len;
    	}
  	}
}

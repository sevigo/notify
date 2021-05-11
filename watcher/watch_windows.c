#include "watch_windows.h"
#include <windows.h>
#include <stdio.h>
#include <stdlib.h>

#define BUFFER_SIZE 1024
#define INSTANCES 2
#define PIPE_TIMEOUT 5000
#define BUFSIZE 4096

// Create the name pipe by the pipe name;
HANDLE stopWatchHandle;
HANDLE eventToChild;
int totalWatchers = 0;
const char *const PIPE_NAME = "\\\\.\\pipe\\stopWatchPipe";

void Setup()
{
	printf("[CGO] [INFO] Setup\n");
	eventToChild = CreateEvent(NULL, TRUE, FALSE, NULL);
}

void StopWatching(char *dir)
{
	printf("[CGO] [INFO] StopWatching %s\n", dir);
	HANDLE pipe;

	stopWatchHandle = CreateNamedPipe(PIPE_NAME,				   // pipe name
									  PIPE_ACCESS_DUPLEX,		   // read/write access
									  PIPE_TYPE_MESSAGE |		   // message-type pipe
										  PIPE_READMODE_MESSAGE |  // message-read mode
										  PIPE_WAIT,			   // blocking mode
									  INSTANCES,				   // number of instances
									  BUFFER_SIZE * sizeof(TCHAR), // output buffer size
									  BUFFER_SIZE * sizeof(TCHAR), // input buffer size
									  PIPE_TIMEOUT,				   // client time-out
									  NULL);					   // default security attributes

	if (stopWatchHandle == INVALID_HANDLE_VALUE)
	{
		// https://docs.microsoft.com/en-us/windows/desktop/debug/system-error-codes--0-499-
		printf("[CGO] [ERROR] Setup(): CreateNamedPipe failed with %d\n", GetLastError());
		return;
	}

	// notify the child process
	if (!SetEvent(eventToChild))
	{
		printf("[CGO] [ERROR] StopWatching(): SetEvent failed (%d)\n", GetLastError());
		return;
	}

	while (TRUE)
	{
		pipe = CreateFile(PIPE_NAME, // pipe name
						  GENERIC_READ | GENERIC_WRITE,
						  0,			 // no sharing
						  NULL,			 // default security attributes
						  OPEN_EXISTING, // opens existing pipe
						  FILE_ATTRIBUTE_NORMAL,
						  NULL); // no template file

		if (pipe != INVALID_HANDLE_VALUE)
		{
			break;
		}

		// Exit if an error other than ERROR_PIPE_BUSY occurs.
		if (GetLastError() != ERROR_PIPE_BUSY)
		{
			printf("[CGO] [ERROR] StopWatching(): Could not open pipe. %d\n", GetLastError());
			return;
		}

		// All pipe instances are busy, so wait for 20 seconds.
		if (!WaitNamedPipe(PIPE_NAME, 20000))
		{
			printf("[CGO] [ERROR] StopWatching(): Could not open pipe: 20 second wait timed out.");
			return;
		}
	}

	// The pipe connected; change to message-read mode.
	DWORD dwMode = PIPE_READMODE_MESSAGE;
	if (!SetNamedPipeHandleState(pipe, &dwMode, NULL, NULL))
	{
		printf("[CGO] [ERROR] StopWatching(): SetNamedPipeHandleState faild: %d\n", GetLastError());
		return;
	}

	printf("[CGO] [INFO] StopWatching(): Have total %d directory watchers\n", totalWatchers);
	// Write the data to the named pipe
	DWORD writtenSize;
	DWORD cbToWrite = (lstrlen(dir) + 1) * sizeof(TCHAR);
	for (int i = 0; i <= totalWatchers; i++)
	{
		if (!WriteFile(pipe, dir, cbToWrite, &writtenSize, NULL))
		{
			printf("[CGO] [ERROR] StopWatching(): WriteFile failed (%d)\n", GetLastError());
		}
	}

	CloseHandle(pipe);
}

// Install https://sourceforge.net/projects/mingw-w64/ to compile (x86_64-8.1.0-posix-seh-rt_v6-rev0)
// For the API documentation see:
// https://msdn.microsoft.com/de-de/library/windows/desktop/aa365261(v=vs.85).aspx
// https://docs.microsoft.com/en-us/windows/desktop/api/fileapi/nf-fileapi-findfirstchangenotificationa
void WatchDirectory(char *dir)
{
	printf("[CGO] [INFO] WatchDirectory(): %s\n", dir);
	totalWatchers++;
	size_t count;
	DWORD waitStatus;
	DWORD dw;
	char buffer[BUFFER_SIZE];
	HANDLE handle;
	OVERLAPPED ovlEventHandle = {0};

	// FILE_NOTIFY_CHANGE_FILE_NAME  – File creating, deleting and file name changing
	// FILE_NOTIFY_CHANGE_DIR_NAME   – Directories creating, deleting and file name changing
	// FILE_NOTIFY_CHANGE_ATTRIBUTES – File or Directory attributes changing
	// FILE_NOTIFY_CHANGE_SIZE       – File size changing
	// FILE_NOTIFY_CHANGE_LAST_WRITE – Changing time of write of the files
	// FILE_NOTIFY_CHANGE_SECURITY   – Changing in security descriptors
	handle = FindFirstChangeNotification(
		dir,  // directory to watch
		TRUE, // do watch subtree
		FILE_NOTIFY_CHANGE_LAST_WRITE | FILE_NOTIFY_CHANGE_FILE_NAME | FILE_NOTIFY_CHANGE_DIR_NAME);

	ovlEventHandle.hEvent = CreateEvent(NULL, TRUE, FALSE, NULL);
	HANDLE handles[] = {ovlEventHandle.hEvent, eventToChild};

	if (handle == NULL)
	{
		printf("[CGO] [ERROR] WatchDirectory(): Unexpected NULL from CreateFile for directroy [%s]\n", dir);
		ExitProcess(GetLastError());
	}

	if (handle == INVALID_HANDLE_VALUE)
	{
		printf("[CGO] [ERROR] WatchDirectory(): FindFirstChangeNotification function failed for directroy [%s] with error [%s]\n", dir, GetLastError());
		ExitProcess(GetLastError());
	}

	ReadDirectoryChangesW(handle, buffer, sizeof(buffer), FALSE, FILE_NOTIFY_CHANGE_LAST_WRITE, NULL, &ovlEventHandle, NULL);
	while (TRUE)
	{
		waitStatus = WaitForMultipleObjects(
			INSTANCES, // number of event objects
			handles,   // array of event objects
			FALSE,	   // does not wait for all
			INFINITE); // waits indefinitely

		switch (waitStatus)
		{
		case WAIT_OBJECT_0:
			// printf("[CGO] [INFO] A file was created, renamed, or deleted\n");
			GetOverlappedResult(
				handle,			 // pipe handle
				&ovlEventHandle, // OVERLAPPED structure
				&dw,			 // bytes transferred
				FALSE);			 // does not wait

			char fileName[MAX_PATH] = "";

			FILE_NOTIFY_INFORMATION *fni = NULL;
			DWORD offset = 0;
			do
			{
				fni = (FILE_NOTIFY_INFORMATION *)(&buffer[offset]);

				if (fni->Action == 0)
				{
					printf("[CGO] [INFO] file notification is empty, exignore\n");
					break;
				}

				wcstombs_s(&count, fileName, sizeof(fileName), fni->FileName, (size_t)fni->FileNameLength / sizeof(WCHAR));

				// FILE_ACTION_ADDED=0x00000001: The file was added to the directory.
				// FILE_ACTION_REMOVED=0x00000002: The file was removed from the directory.
				// FILE_ACTION_MODIFIED=0x00000003: The file was modified. This can be a change in the time stamp or attributes.
				// FILE_ACTION_RENAMED_OLD_NAME=0x00000004: The file was renamed and this is the old name.
				// FILE_ACTION_RENAMED_NEW_NAME=0x00000005: The file was renamed and this is the new name.
				printf("[CGO] [INFO] file=[%s] action=[%d]\n", fileName, fni->Action);

				goCallbackFileChange(dir, fileName, fni->Action);
				memset(fileName, '\0', sizeof(fileName));
				offset += fni->NextEntryOffset;
			} while (fni->NextEntryOffset != 0);

			ResetEvent(ovlEventHandle.hEvent);
			if (ReadDirectoryChangesW(handle, buffer, sizeof(buffer), FALSE, FILE_NOTIFY_CHANGE_LAST_WRITE, NULL, &ovlEventHandle, NULL) == 0)
			{
				printf("[CGO] [INFO] Reading Directory Change\n");
			}
			break;

		case WAIT_OBJECT_0 + 1:
			printf("[CGO] [INFO] WatchDirectory(): Stop watching signal for [%s]\n", dir);

			DWORD numRead;
			if (!ReadFile(stopWatchHandle, buffer, BUFFER_SIZE * sizeof(TCHAR), &numRead, NULL))
			{
				printf("[CGO] [ERROR] read from pipe: ReadFile failed (%d)\n", GetLastError());
			}

			if (numRead > 0)
			{
				int i = strcmp(buffer, dir);
				if (i == 0)
				{
					CloseHandle(handle);
					totalWatchers--;
					ResetEvent(eventToChild);
					goto EndWhile;
				}
			}

			ResetEvent(eventToChild);
			break;

		case WAIT_TIMEOUT:
			printf("\nNo changes in the timeout period.\n");
			break;

		default:
			printf("\n ERROR: Unhandled status.\n");
			ExitProcess(GetLastError());
			break;
		}
	}
EndWhile:;
	printf("[CGO] [INFO] WatchDirectory(): stop watching %s\n", dir);
}

# Display Configuration Query Tool

This Go program provides two methods for querying display configuration information on a Windows system: one using CGO (cgo) and the other using syscall.

## Prerequisites

- This program is intended to run on a Windows system.
- Ensure that you have a valid Go environment set up.

## How to Use

1. Clone the repository to your local machine:

   ```bash
   git clone https://github.com/nmarandi/QueryDisplayConfig.git
   ```

2. Navigate to the directory containing the code:

   ```bash
   cd QueryDisplayConfig
   ```

3. Set the CGO_ENABLED environment variable to 1:

   ```bash
   go env -w CGO_ENABLED=1
   ```

4. Build and run the program:

   ```bash
   go run main.go
   ```

## Overview

The program uses the Windows API to query and retrieve display configuration information. It defines two methods, one using CGO and the other using syscall, to achieve the same goal.

### CGO Method

The CGO method leverages the `user32.dll` library and C bindings to interact with the Windows API. It utilizes the `QueryDisplayConfig` and `GetDisplayConfigBufferSizes` functions to obtain information about active display paths and virtual modes.

### Syscall Method

The syscall method directly calls the Windows API functions (`QueryDisplayConfig` and `GetDisplayConfigBufferSizes`) without using CGO. It achieves the same result as the CGO method but demonstrates an alternative approach using pure Go.

## Code Structure

- `main.go`: The main entry point of the program that executes the display configuration query using both CGO and syscall methods and Contains the necessary data structures and constants for interacting with the Windows API. Defines functions (`queryDisplayConfigCGO` and `queryDisplayConfigSyscall`) for querying and printing display configuration information.

## Output

The program prints information about active display paths, including source and target details, as well as display modes (resolution and refresh rate) for each path.

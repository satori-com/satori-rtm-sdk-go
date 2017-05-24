# Example

For tutorial purposes, we subscribe to the same channel that we publish a
message to. So we receive our own published message. This allows end-to-end
illustration of data flow with just a single client.

## Before you start
Make sure that you have `go` installed: https://golang.org/doc/install  
Also please make sure that your Go-Workspace is properly configured: https://golang.org/doc/code.html#Workspaces  
and your **GOPATH** environment variable is point to your Go-Workspace.


Hint: Run `make run` from the console to make auto-configuration for the Go-Workspace:

    $ git clone git@github.com:satori-com/satori-rtm-sdk-go.git
    $ cd satori-rtm-sdk-go/tutorial/
    $ make run

Check the **primer.go** source code to get more information

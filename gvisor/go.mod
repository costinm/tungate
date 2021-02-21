module github.com/costinm/tungate/gvisor

go 1.16

//replace gvisor.dev/gvisor => ../gvisor
//replace github.com/costinm/ugate => ../ugate

require (
	github.com/bazelbuild/rules_go v0.25.1 // indirect
	github.com/costinm/tungate v0.0.0-20210221162225-dbcc36f47d74
	github.com/costinm/ugate v0.0.0-20210106052904-4da1a58a92e6
	github.com/songgao/water v0.0.0-20200317203138-2b4b6d7c09d8 // indirect
	golang.org/x/sys v0.0.0-20210220050731-9a76102bfb43 // indirect
	gvisor.dev/gvisor v0.0.0-20201215175918-b0f23fb7e0cf

)

module github.com/costinm/tungate/gvisor

go 1.16

//replace github.com/google/netstack => github.com/costinm/netstack v0.0.0-20190601172006-f6e50d4d2856

//replace github.com/google/netstack => ../netstack

//replace gvisor.dev/gvisor => ../gvisor

//replace github.com/costinm/ugate => ../ugate

require (
	github.com/costinm/tungate v0.0.0-20210106054017-3c4979c12690
	github.com/costinm/ugate v0.0.0-20210106052904-4da1a58a92e6
	gvisor.dev/gvisor v0.0.0-20201215175918-b0f23fb7e0cf

)

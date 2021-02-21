module github.com/costinm/tungate/netstack

go 1.16

replace github.com/google/netstack => github.com/costinm/netstack v0.0.0-20190601172006-f6e50d4d2856

//replace github.com/google/netstack => ../netstack
//replace github.com/costinm/ugate => ../ugate

require (
	github.com/costinm/tungate v0.0.0-20210106054017-3c4979c12690
	github.com/costinm/ugate v0.0.0-20210106052904-4da1a58a92e6
	github.com/google/netstack v0.0.0-00010101000000-000000000000
)

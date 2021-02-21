module github.com/costinm/tungate/netstack

go 1.16

replace github.com/google/netstack => github.com/costinm/netstack v0.0.0-20210221225206-58e23cf0bc03

//replace github.com/google/netstack => ../../netstack

//replace github.com/costinm/ugate => ../../ugate

require (
	github.com/costinm/tungate v0.0.0-20210221162225-dbcc36f47d74
	github.com/costinm/ugate v0.0.0-20210221155556-10edd21fadbf
	github.com/google/netstack v0.0.0-00010101000000-000000000000
	github.com/songgao/water v0.0.0-20200317203138-2b4b6d7c09d8 // indirect
)

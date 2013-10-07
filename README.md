go-socks-server
===============

Socks5 server in go. This was written in idiomatic Go with 2 goroutines per connection. One for handling inbound and another for 
outbound traffic. 

### Usage ###

    socks [optional binding address. Default is 127.0.0.1:8085]
    
### Example ###

    socks 127.0.0.1:1234

### Not Implemented ###

* Authentication
* IPV6
* Port Binding
* UDP    
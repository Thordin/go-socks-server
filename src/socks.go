package main 

import (
	"net"
	"encoding/binary"
	"io"
	"fmt"
	"time"
	"os"
	"runtime"
	"runtime/debug"
)

type address_enum byte

const (
	IPV4 = 1
	DOMAIN = 3
	IPV6 = 4
)

type connection_request struct {
	Version byte
	Command byte
	Reserved byte
	Address_type address_enum
}

func status() {
	var stats runtime.MemStats
	for {
		debug.FreeOSMemory()
		runtime.ReadMemStats(&stats)
		fmt.Printf("HeapAlloc %d HeapSys %d HeapRelease %d Goroutines %d\n", 
					stats.HeapAlloc, stats.HeapSys, stats.HeapReleased, runtime.NumGoroutine())
		time.Sleep(10*time.Second)
	}
}

func main() {
	var address = "127.0.0.1:8085"
	if len(os.Args) > 1 {
		address = os.Args[1]
	}
	
	fmt.Println("Listening on", address);

	listen,err := net.Listen("tcp4", address)
	if err != nil {
		panic(err)
	}	
	
	go status()
	
	for {
		con,err := listen.Accept()
		if err != nil {
			panic(err)
		}
		con.SetDeadline(time.Now().Add(1*time.Second))
		go auth(con)
	}
}

func auth(con net.Conn) {
	var version, method_count byte
	var err error 
	
	err = binary.Read(con, binary.LittleEndian, &version)
	if err != nil { 
		con.Close()
		return
	}
	if version != 5 {
		con.Close()
		return
	}
	err = binary.Read(con, binary.LittleEndian, &method_count)
	if err != nil {
		con.Close()
		return
	}
	
	methods := make([]byte, method_count)
	_,err = io.ReadFull(con, methods)
	if err != nil {
		con.Close()
		return
	}
	
	_,err = con.Write([]byte{5,0})
	if err != nil { 
		con.Close()
		return
	}
	
	request := &connection_request{}
	err = binary.Read(con, binary.LittleEndian, request)
	if err != nil {
		con.Close()
		return
	}
	
	if request.Version != 5 || request.Command != 1 {
		con.Close()
		return
	}
	
	var address string
	var port int16
	
	switch request.Address_type {
		case IPV4:
			ip := make([]byte, 4)
			err = binary.Read(con, binary.LittleEndian, &ip)
			if err != nil {
				fmt.Println(err)
				con.Close()
				return
			}
			err = binary.Read(con, binary.BigEndian, &port)
			if err != nil {
				fmt.Println(err)
				con.Close()
				return
			}
			address = net.IP(ip).String()
		case DOMAIN:
			var size byte
			err = binary.Read(con, binary.LittleEndian, &size)
			if err != nil {
				fmt.Println(err)
				con.Close()
				return
			}
			domain := make([]byte, size)
			_,err := io.ReadFull(con, domain)
			if err != nil {
				fmt.Println(err)
				con.Close()
				return
			}
			address = string(domain)
		default:
			fmt.Println("Invalid address type", request.Address_type)
			con.Close()
			return
	}
	
	err = binary.Read(con, binary.BigEndian, &port)
	if err != nil {
		fmt.Println(err)
		con.Close()
		return
	}	
	
	out,err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", address, port), 5 * time.Second)
	fmt.Printf("Connecting %s:%d\n", address, port)
	
	if err != nil {
		con.Close()
		return
	}
	
	_,err = con.Write([]byte{5,0,0,1,0,0,0,0,0,0})
	if err != nil { 
		con.Close()
		return
	}
	
	now := time.Now().Add(time.Second*60*2)
	con.SetDeadline(now)
	out.SetDeadline(now)
	
	go inbound(con, out)
	go outbound(con, out)
}

func inbound(in net.Conn, out net.Conn) {
	_,err := io.Copy(in, out)
	if err != nil {
		in.Close()
		out.Close()
	}
}

func outbound(in net.Conn, out net.Conn) {
	_,err := io.Copy(out, in)
	if err != nil {
		in.Close()
		out.Close()
	}
}
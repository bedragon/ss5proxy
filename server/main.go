package main
import(
        "fmt"
        "net"
        "log"
        //"sync"
    )


type clientConn struct {
    id int16
    conn net.Conn
    isBusy bool
}

type dataChunk struct{
    clientConnID int16
    data []byte;
}


func doConn(conn net.Conn){
    buf := make([]byte, 1024)
    defer conn.Close()
    for {
        n, err := conn.Read(buf)
        if n <= 0 || err != nil {
            break
        }
        fmt.Println(n, buf)
		//buf[0] = 0x78
		//buf[1] = byte(thisClientConn.id >> 8)
		//buf[2] = byte(thisClientConn.id & 0xff)
		//buf[3] = byte(n >> 8)
		//buf[4] = byte(n & 0xff)
		//gClientSendChannel <- buf
        //conn.Write([]byte("HTTP/1.1 200 OK\r\nServer:Apache Tomcat/5.0.12\r\nDate:Mon,6Oct2003 13:23:42 GMT\r\nContent-Length:1\r\n\r\n"))
        //conn.Write([]byte("a"))
        //conn.Close()
        
    }
    fmt.Println("xixi")
}
func start() {
    tcpAddr, err := net.ResolveTCPAddr("tcp", ":9188")
    if err != nil {
        log.Fatal(err)
    }
    listen, err := net.ListenTCP("tcp", tcpAddr)
    if err != nil {
        log.Fatal(err)
    }
    for {
        conn, err := listen.Accept()
        if err != nil {
            continue
        }
        go doConn(conn)
    }
}
func main() {
    start()
}

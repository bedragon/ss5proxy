package main
import(
        "fmt"
        "net"
        "log"
        "sync"
    )

const MAX_CONN = 512

type clientConn struct {
    id int16
    conn net.Conn
    isBusy bool
}

type dataChunk struct{
    clientConnID int16
    data []byte;
}

var gClientConn []clientConn
var gClientConnIndex int16
var gClientConnNum int16
var gClientConnAllocLock sync.Mutex
var gClientSendChannel = make(chan []byte, 1)

func serverRecvData(conn net.Conn){
	fmt.Println("serverRecvData start")
	for {
		data := <- gClientSendChannel
        length := int16(data[3]) << 8 | int16(data[4])
        fmt.Println("serverRecvData send byte : ", data[5:])
        conn.Write(data[:length + 5])
	}
}

func runServer(){
	addr, err := net.ResolveTCPAddr("tcp", "10.95.97.205:9188")
	if err != nil {
		log.Fatal(err)
	}
	conn,err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}
	initGClientConn()
	go serverRecvData(conn)
	var header [5]byte
	var id,readLength int16
	for {
		_ , err := conn.Read(header[:])
		if err != nil {
            fmt.Println("runServer fatal")
			log.Fatal(err)
		}
		flag := byte(header[0])
		id = int16(header[1]) << 8 | int16(header[2])
		length := int16(header[3]) << 8 | int16(header[4])
		if (flag != 0x78) {
			log.Fatal("runServer recv data flag is error", header)
			continue
		}
		if length == 0 {
			log.Println("id:%d client recv data length:0", header)
			closeClientConn(id)
			continue	
		}
		buf := make([]byte, length)
        readLength = 0
        for readLength != length {
            tmpLength, err := conn.Read(buf[readLength:])
            if err != nil {
                log.Fatal(err)
                continue
            }
            readLength = int16(tmpLength) + readLength
        }
        fmt.Println("runServer recv data :", id, length,readLength, header, buf)
		sendClientConn(id, buf)
	}
	   
}

func sendClientConn(id int16, buf []byte) {
    fmt.Println("sendClientConn id is:" ,id, gClientConn[id].conn)
	gClientConn[id].conn.Write(buf)	
}

func closeClientConn(id int16){
	gClientConn[id].conn.Close()
	gClientConn[id].isBusy = false
}

func doConn(conn net.Conn){
    fmt.Println("xixi")
    thisClientConn := allocClientConn(conn)
    thisClientConn.conn = conn
    fmt.Println("doConn alloc id:", thisClientConn.id, gClientConn[thisClientConn.id].conn)
    buf := make([]byte, 10000)
    defer conn.Close()
    for {
        n, err := conn.Read(buf[5:])
        if n <= 0 || err != nil {
            break
        }
		//var sendData []byte
		buf[0] = 0x78
		buf[1] = byte(thisClientConn.id >> 8)
		buf[2] = byte(thisClientConn.id & 0xff)
		buf[3] = byte(n >> 8)
		buf[4] = byte(n & 0xff)
        fmt.Println("doConn send id : %d, length : %d", thisClientConn.id, n)
        //fmt.Println(buf)
		gClientSendChannel <- buf[:n+5]
        //conn.Write([]byte("HTTP/1.1 200 OK\r\nServer:Apache Tomcat/5.0.12\r\nDate:Mon,6Oct2003 13:23:42 GMT\r\nContent-Length:1\r\n\r\n"))
        //conn.Write([]byte("a"))
        //conn.Close()
           
    }
    fmt.Println("doConn close id:", thisClientConn.id)
}
func initGClientConn(){
    gClientConn = make([]clientConn, MAX_CONN);
    gClientConnIndex = 0;
    gClientConnNum = MAX_CONN;
}

func allocClientConn(conn net.Conn) clientConn {
    gClientConnAllocLock.Lock()
    for gClientConn[gClientConnIndex].isBusy {
        gClientConnIndex = (gClientConnIndex + 1) % gClientConnNum
    }
    gClientConn[gClientConnIndex].id = gClientConnIndex
    gClientConn[gClientConnIndex].conn = conn
    gClientConn[gClientConnIndex].isBusy = true
    gClientConnAllocLock.Unlock()
    return gClientConn[gClientConnIndex]
}
func start() {
    tcpAddr, err := net.ResolveTCPAddr("tcp", ":8188")
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
	go runServer()
    start()
}

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
		fmt.Println("aaa serverRecvData")
		conn.Write(data)
	}
}

func runServer(){
	addr, err := net.ResolveTCPAddr("tcp", ":9188")
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
	var id int16
	for {
		_ , err := conn.Read(header[:])
		if err != nil {
			log.Fatal(err)
		}
		flag := string(header[0])
		id = int16(header[1]) << 8 | int16(header[2])
		length := int(header[3]) << 8 | int(header[4])
		if (flag != "x") {
			log.Fatal("runServer recv data flag is error")
			continue
		}
		if length == 0 {
			log.Println("id:%d client recv data length:0")
			closeClientConn(id)
			continue	
		}
		buf := make([]byte, length)
		_, err = conn.Read(buf)
		if err != nil {
			log.Fatal(err)
			continue
		}
		sendClientConn(id, buf)
	}
	   
}

func sendClientConn(id int16, buf []byte) {
	gClientConn[id].conn.Write(buf)	
}

func closeClientConn(id int16){
	gClientConn[id].conn.Close()
	gClientConn[id].isBusy = false
}

func doConn(conn net.Conn){
    fmt.Println("xixi")
    thisClientConn := allocClientConn()
    buf := make([]byte, 1024)
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
		fmt.Println(n,"hahbbbb", buf)
		gClientSendChannel <- buf[:n+5]
        //conn.Write([]byte("HTTP/1.1 200 OK\r\nServer:Apache Tomcat/5.0.12\r\nDate:Mon,6Oct2003 13:23:42 GMT\r\nContent-Length:1\r\n\r\n"))
        //conn.Write([]byte("a"))
        //conn.Close()
        
    }
}
func initGClientConn(){
    gClientConn = make([]clientConn, MAX_CONN);
    gClientConnIndex = 0;
    gClientConnNum = MAX_CONN;
}

func allocClientConn() clientConn {
    gClientConnAllocLock.Lock()
    for gClientConn[gClientConnIndex].isBusy {
        gClientConnIndex = (gClientConnIndex + 1) % gClientConnNum
    }
    gClientConn[gClientConnIndex].id = gClientConnIndex
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

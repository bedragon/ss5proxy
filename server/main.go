package main
import(
        "fmt"
        "net"
        "log"
        "sync"
    )


type clientConn struct {
    id int16
    conn net.Conn
    isBusy bool
}

type proxyConn struct{
    clientConnID int16
    ss5Conn net.Conn
    isBusy bool
    m sync.Mutex
}

func (oneProxyConn *proxyConn) close(){
    oneProxyConn.ss5Conn.Close()
    oneProxyConn.isBusy = false
    oneProxyConn = nil
}

func (oneProxyConn *proxyConn) new(id int16) *proxyConn{
    oneProxyConn.clientConnID = id
    addr, err := net.ResolveTCPAddr("tcp", ":1080")
    if err != nil {
        log.Fatal(err)
        return nil
    }
    oneProxyConn.ss5Conn, err = net.DialTCP("tcp", nil, addr)
    if err != nil {
        log.Fatal(err)
        return nil
    }
    oneProxyConn.isBusy = true
    return oneProxyConn
}

func (oneProxyConn *proxyConn) runSs5Conn(clientConn net.Conn) {
    fmt.Println("new runSs5Conn id : ", oneProxyConn.clientConnID)
    defer oneProxyConn.close()
    buf := make([]byte, 1024)
    for {
        length, _ := oneProxyConn.ss5Conn.Read(buf[5:])
        buf[0] = 0x78
        buf[1] = byte(oneProxyConn.clientConnID >> 8)
        buf[2] = byte(oneProxyConn.clientConnID & 0xff)
        buf[3] = byte(length >> 8)
        buf[4] = byte(length & 0xff)
        clientConn.Write(buf[:length + 5])
        if (length == 0){
            fmt.Println("runSs5Conn send byte is zero : ", buf[:length + 5])
            break
        }
        //fmt.Println("runSs5Conn send id : ", oneProxyConn.clientConnID, buf[:length + 5])
    }

}


func doConn(conn net.Conn){
    proxyConnArr := make([]proxyConn, 512)
    header := make([]byte, 5)
    defer conn.Close()
    var readLength int16
    for {
        n, err := conn.Read(header)
        if n <= 0 || err != nil {
            log.Fatal(err)
        }
        fmt.Println("doConn recv byte : ", header[0:5])
        flag := byte(header[0])
        id := int16(header[1]) << 8 | int16(header[2])
        length := int16(header[3]) << 8 | int16(header[4])
        fmt.Println("new doConn id : %d, length : %d", id, length)
        if (flag != 0x78) {
            log.Fatal("runServer recv data flag is error")
            continue
        }
        if length == 0 {
            log.Println("id:%d client recv data length:0")
            //closeClientConn(id)
            continue
        }
        buf := make([]byte, length)
        readLength = 0
        for readLength < length {
            fmt.Println("for read ",readLength, length)
            bufLength, err := conn.Read(buf[readLength:])
            if err != nil {
                log.Fatal(err)
                continue
            }
            readLength = readLength + int16(bufLength)
                
        }
        fmt.Println("recv byte : ", buf)
        if proxyConnArr[id].isBusy == false {
            newOneProxyConn := proxyConnArr[id].new(id)
            go newOneProxyConn.runSs5Conn(conn)
        }
        proxyConnArr[id].ss5Conn.Write(buf)
        
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

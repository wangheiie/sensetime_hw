package main

import (

    "bufio"
    "fmt"
    "flag"
    "io/ioutil"
    "log"
    "math/rand"
    "net"
    "os"
    //"path/filepath"
    "regexp"
    "strconv"
    "strings"
    "time"

)

var (
    host = flag.String("h","127.0.0.1","listen host")
    port = flag.Int("p",666,"listen port")
    dir = flag.String("d","/opt/myftp/data","data path")
    curdir string
    user = "myftp"
    password = "myftp"
)

type Conn struct{
    conn             net.Conn
    dataConn         net.Conn
    logger           Logger
    //requestUser      string
    //user             string
}

type Logger interface {

    Print(message interface{})

}

type FtpLogger struct{}

func (logger *FtpLogger) Print(message interface{}){
    log.Printf("%s", message)
}

func newConn(tcpconn net.Conn) *Conn{
    c := new(Conn)
    c.conn = tcpconn
    //c.dataConn = 
    return c
}

func ValidArgs() (h string, p int, acpt bool) {
        h = *host
        p = *port
        
	host_ok, _ := regexp.MatchString(
		"^(25[0-5]|2[0-4]\\d|[0-1]?\\d?\\d)(\\.(25[0-5]|2[0-4]\\d|[0-1]?\\d?\\d)){3}$", h)
	if !host_ok {
		fmt.Println("Invalid Host")
		return "", 0, false
	}
	if p<=0 || p > 65535 {
		fmt.Println("Invalid Port")
		return "", 0, false
	}
	return h, p, true
}

func CurDir(path string) (string,bool){

    path_ok ,_ := PathExists(path)
    if !path_ok {
        fmt.Println("No such directory.............")
        return "",false
    }
    return path,true
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func CheckPasswd(reqU string, passwd string) (string,bool){
        if reqU == user{
            if passwd == password{
                return reqU,true
            }else{
                //conn.Write("Password checked wrong")
                return "",false
            }
        }else{
            //conn.Write("Username checked wrong")
            return "",false
        }
}


func Server(conn *Conn){
    fmt.Println("Connection success............")
    conn.logger.Print("Conncetion success.")
    ddir := *dir
    _ ,dir_ok := CurDir(ddir)
    if !dir_ok {
        conn.conn.Write("Please confirm your directory correct and retry starting............")
        conn.logger.Print("Input Wrong dir.")
        return
    }
    buf := make([]byte, 4096)
    for {
        length, err := conn.conn.Read(buf)
        if err != nil {
            conn.conn.Write("Error Reading Client", err.Error())
            conn.logger.Print("Reading Client error.")
            return
        }
        line := string(buf[:length])
        ExecCommand(line,conn)
    }
    conn.logger.Print("Connection close.")
    conn.conn.Close()
    fmt.Println("Connection Close.............")
}

func ExecCommand(line string,conn *Conn){
    command, param := parsecmd(line)
    cmd := strings.ToUpper(command)
    switch {

        case cmd == "CWD":
            cwddir, err := CurDir(param)
            if err == false {
                curdir = cwddir
                os.Chdir(curdir)
                msg := fmt.Sprintln("%d %s", 250, "Directory change to "+cwddir)
                conn.conn.Write([]byte(msg))
                conn.logger.Print(msg)
            }else{
                
                msg := fmt.Sprintln("%d %s", 550, "Directory change failed. ")
                conn.conn.Write([]byte(msg))
                conn.logger.Print(msg)
            }

        case cmd == "LIST":
            listpath,err := os.Stat(param)
            if err != nil {
                conn.conn.Write([]byte("Path stat error"))
                return
            }
            if listpath == nil || !listpath.IsDir() {
                conn.conn.Write([]byte("Input is not a dir"))
                return
            }
            lsdir, err := ioutil.ReadDir(param)
            if err != nil {
                conn.conn.Write([]byte("Read dir error"))
            }
            listinfo := "id  filename  mode  size\n"
            for k,v := range lsdir{
                listinfo += fmt.Sprintf("%d  %s  %s  %d\n",k ,v.Name() ,v.Mode().String(), v.Size())
            }
            conn.conn.Write([]byte(listinfo))

        case cmd == "PASS":
            pass_ok,err := CheckPasswd(user,param)
            if err != true {
                conn.conn.Write([]byte("550 Password checking error"))
                conn.logger.Print("550 Password checking error")
                return
            }
            if pass_ok != "" {
                //conn.user = conn.requestUser
                //conn.requestUser = ""
                conn.conn.Write([]byte("230 password accept"))
            }else{
                conn.conn.Write([]byte("530 Wrong username or password"))
            }

        case cmd == "PASV":
            //pasvip := host
            minPort, maxPort := 30000,32000
            pasvPort := minPort + rand.Intn(maxPort - minPort)
            laddr, err := net.ResolveTCPAddr("tcp", *host+":"+strconv.Itoa(pasvPort))
            if err != nil {
                fmt.Println("TCP addr build error")
                return
            }
            //var pasvSocket net.Listener
            pasvListener, err := net.ListenTCP("tcp", laddr)
            pasvSocket,_ := pasvListener.Accept()
            conn.dataConn = pasvSocket
            //addr := pasvSocket.Addr()
            //fmt.Println(addr)
            if err != nil {
                fmt.Println("Pasv tcp listen error")
                return
            }
            port1 := pasvPort/256
            port2 := pasvPort - port1*256
            ipQua := strings.Split(*host, ".")
            pasvIPPort := fmt.Sprintf("(%s,%s,%s,%s,%d,%d)",ipQua[0],ipQua[1],ipQua[2],ipQua[3],port1,port2)
            conn.conn.Write([]byte("227 Pasv mode "+pasvIPPort))
 
        case cmd == "PORT":
            conn.conn.Write([]byte("Just for fun. Don`t worry, be happy."))

        case cmd == "PWD":
            curdir,_ = os.Getwd()
            conn.conn.Write([]byte("257 Current directory is "+curdir+"."))

        case cmd == "RETR":
            target_ok ,_ := PathExists(param)
            if !target_ok {
                conn.conn.Write([]byte("Path error"))
                return
            }
            retrFile, err := os.Open(param)
            if err != nil {
                fmt.Println("RETR file open err")
                return
            }
            retrBytes := make([]byte,4096)
            retrReader := bufio.NewReader(retrFile)
            //retrFileInfo,_ := os.Stat(param)
            //retrlen := retrFileInfo.Size()
            defer retrFile.Close()
            conn.conn.Write([]byte("150 Data transfer starting..........."))
            for {
                readlen,_ := retrReader.Read(retrBytes)
                if readlen == 0 {
                    conn.conn.Write([]byte("Transfer completed."))
                    return
                }
                
                _, err = conn.dataConn.Write([]byte(retrBytes))
                if err != nil {
                    conn.conn.Write([]byte("550 Transfer error."))
                    return
                }
            }
            
        case cmd == "STOR":
            target_ok ,_ := PathExists(param)
            if !target_ok {
                conn.conn.Write([]byte("Path error"))
                return
            }
            storFile, err := os.OpenFile(param, os.O_WRONLY|os.O_CREATE, 0666)
            if err != nil {
                fmt.Println("File store error")
                return
            }
            defer storFile.Close()
            storWriter := bufio.NewWriter(storFile)
            storBytes := make([]byte,4096)
            for {
                
                buflen, err := conn.dataConn.Read(storBytes)
                if err != nil {
                    fmt.Println("Read connection error.")
                    return
                }
                storlen, err := storWriter.Write(storBytes)
                if storlen == buflen {
                    storWriter.Flush()
                    fmt.Println("Transimission completed.")
                }
            }
            //storFile(conn, param) 
            //conn.Write([]byte())
            
        case cmd == "USER":
            if user == param {
                msg := fmt.Sprintln("%d %s", 331, "Username ready")
                conn.conn.Write([]byte(msg))
            }else{
                conn.conn.Write([]byte("Wrong username. Retry please.")) 
            }
    }
}

func parsecmd(line string) (string, string){
    params := strings.SplitN(strings.Trim(line, "\r\n"), " ", 2)
    if len(params) == 1 {
        return params[0],""
    }
    return params[0],strings.TrimSpace(params[1])
}

func main(){

    flag.Parse()
    host,port,acpt := ValidArgs()
    mainLogFile,_ := os.Create("/var/log/only-myftp-"+time.Now().Format("20060102")+".log")
    mainLogger := log.New(mainLogFile, "[IIE]", log.LstdFlags)
    if !acpt {
        mainLogger.Println("Error host ip address or port. Only-Myftp exit.")
        fmt.Println("Error host ip address or port. Only-Myftp exit.")
        return
    }
    fmt.Println("Only-Myftp Server starting at .............", host+":",port)
    listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
    if err != nil {
        fmt.Println("Start listening error.............")
        fmt.Println(err)
        mainLogger.Println("Error starting listening.")
        return
    }
    for {
        tcpconn, err := listener.Accept()
        if err != nil {
            fmt.Println("Listening accept error.............")
            fmt.Println(err)
            mainLogger.Println("Listening accept error.")
            continue
        }
       conn := newConn(tcpconn)
        go Server(conn)
    }
}

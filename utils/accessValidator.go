package utils

import (
    "bytes"
    "io"
    "log"
    //"os"
    "strings"
    "fmt"
    "os/exec"
)

func Execute(output_buffer *bytes.Buffer, stack ...*exec.Cmd) (err error) {
    var error_buffer bytes.Buffer
    pipe_stack := make([]*io.PipeWriter, len(stack)-1)
    i := 0
    for ; i < len(stack)-1; i++ {
        stdin_pipe, stdout_pipe := io.Pipe()
        stack[i].Stdout = stdout_pipe
        stack[i].Stderr = &error_buffer
        stack[i+1].Stdin = stdin_pipe
        pipe_stack[i] = stdout_pipe
    }
    stack[i].Stdout = output_buffer
    stack[i].Stderr = &error_buffer

    if err := call(stack, pipe_stack); err != nil {
        //log.Fatalln(string(error_buffer.Bytes()), err)
    }
    return err
}

func call(stack []*exec.Cmd, pipes []*io.PipeWriter) (err error) {
    if stack[0].Process == nil {
        if err = stack[0].Start(); err != nil {
            return err
        }
    }
    if len(stack) > 1 {
        if err = stack[1].Start(); err != nil {
             return err
        }
        defer func() {
            if err == nil {
                pipes[0].Close()
                err = call(stack[1:], pipes[1:])
            }
        }()
    }
    return stack[0].Wait()
}

/*func main() {
        fmt.Println("-----------------------------Start ------------------------")
 checkIpAccess("54.239.31.91")
}*/

func CheckIpAccess(ip string, port string) bool {
    var b bytes.Buffer

    fmt.Println("starting check ip")
    fmt.Println("ip to check: " + ip)
    fmt.Println("port to check: " + port)
    if err := Execute(&b,
        exec.Command("netstat", "-tn", "2>/dev/null"),
	      exec.Command("grep", ":" + port),
	      // exec.Command("grep", ":9090"),
	      // exec.Command("grep", ":58730"),
        exec.Command("awk", "{print $5}"),
	      exec.Command("cut", "-d:", "-f1"),
	      exec.Command("sort"),
	      exec.Command("uniq"),
    ); err != nil {
        log.Fatalln(err)
    }

    // io.Copy(os.Stdout, &b)

        // fmt.Println(b.Bytes())
  str:= string(b.Bytes())
  fmt.Println("filter result")
  fmt.Println(str)
  fmt.Println("-----------------------------End ------------------------")

 	// str := "apple orange durian pear"

 	// strArray := strings.Fields(str)
 	//fmt.Println(strArray[1:3])
 	strArray := strings.Split(str, "\n")
	// fmt.Println(strArray)

    _, exist := Find(strArray, ip)
  fmt.Println("EXIST")
  fmt.Println(exist)
 	return exist
 	//fmt.Println(strArray[1:3])

}



func Find(slice []string, val string) (int, bool) {
    for i, item := range slice {
        if item == val {
            return i, true
        }
    }
    return -1, false
}



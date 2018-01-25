package main

import (
	"fmt"
	"io/ioutil"
	"golang.org/x/crypto/ssh"
	"time"
	"bytes"
	"net"
	"strings"
)

const (
	username = "admin"
	password = "123456"
	network_ipaddress       = "10.10.10.100"
	port_number     = 22
	commands     = "display  cpu;display mem;display current-configuration"   //多条命令用 ;  号隔开
)

var (
	ciphers_list = []string{"aes128-gcm@openssh.com", "arcfour256", "aes128-ctr", "aes192-ctr", "aes256-ctr", "arcfour128", "aes128-cbc", "3des-cbc", "aes192-cbc", "aes256-cbc"}
)       //加密方法,可以继续添加 避免对设备不支持

func main() {
	shell_channel, err := connect(username, password, network_ipaddress, "" , port_number, ciphers_list)
	if err != nil {
		fmt.Println("连接创建错误 : ",err)
		return
	}
	defer session.Close()
    cmd_list := strings.Split(commands, ";")
    stdinBuf, err := shell_channel.StdinPipe()
    if err != nil {
        fmt.Println("通道无法进行操作 : ",err)
        return
    }

    var outbt, errbt bytes.Buffer
    shell_channel.Stdout = &outbt

    shell_channel.Stderr = &errbt
    err = shell_channel.Shell()
    if err != nil {
        fmt.Println(err)
        return
    }
    for _, cmd := range cmd_list {
        cmd = cmd + "\n"
        stdinBuf.Write([]byte(cmd))
		for {
			time.Sleep(time.Second * 1)
			if strings.HasSuffix(outbt.String(),"---- More ----") {    //完整显示命令结果处理 ---- More ---- 方法
				stdinBuf.Write([]byte(" \n"))
			} else {
				break
			}
		}
        time.Sleep(time.Second * 1)
    }
    stdinBuf.Write(huawei_quit())
    session.Wait()
    fmt.Println(outbt.String() + errbt.String())

}


func huawei_quit() []byte {                    //华为交换机需要按空格继续
	quit_command := []byte("quit\n")
	return quit_command
}


//使用 https://xiaozhuanlan.com/topic/3490681275 描述方法
func connect(user, password, host, key string, port int, cipherList []string) (*ssh.Session, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		client       *ssh.Client
		config       ssh.Config
		session      *ssh.Session
		err          error
	)
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	if key == "" {
		auth = append(auth, ssh.Password(password))
	} else {
		pemBytes, err := ioutil.ReadFile(key)
		if err != nil {
			return nil, err
		}

		var signer ssh.Signer
		if password == "" {
			signer, err = ssh.ParsePrivateKey(pemBytes)
		} else {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(pemBytes, []byte(password))
		}
		if err != nil {
			return nil, err
		}
		auth = append(auth, ssh.PublicKeys(signer))
	}

	if len(cipherList) == 0 {
		config = ssh.Config{
			Ciphers: []string{"aes128-ctr", "aes192-ctr", "aes256-ctr", "aes128-gcm@openssh.com", "arcfour256", "arcfour128", "aes128-cbc", "3des-cbc", "aes192-cbc", "aes256-cbc"},
		}
	} else {
		config = ssh.Config{
			Ciphers: cipherList,
		}
	}

	clientConfig = &ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: 30 * time.Second,
		Config:  config,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	// connet to ssh
	addr = fmt.Sprintf("%s:%d", host, port)

	if client, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}

	// create session
	if session, err = client.NewSession(); err != nil {
		return nil, err
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		return nil, err
	}

	return session, nil
}

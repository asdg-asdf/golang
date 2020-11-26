package main


import (
        "os"
        "fmt"
        _ "bytes"
        _ "bufio"
        "io"
        "os/exec"
        "net"
        "strings"
        "path/filepath"
        "errors"
        "net/http"
        "path"
        "strconv"
        "time"
        _ "flag"
        "github.com/GeertJohan/go.rice"
    "io/ioutil"
        "regexp"
        "encoding/json"
)

var (
        deploy_type string
        ip string
        usesrname string
        password  string
        port int = 22
        local_ip string
        comfirm_string string
        comfirm_string_x string


)


// define data structure
type app_file struct {
        UI_PKG                     string
        boddy_JAR                 string
        CloudAPP_WEB_PKG               string
        WordFlow_PKG                    string
        auto_exec_SSH             string
        auto_exec_POWERVC         string
        auto_exec_POWER           string
        auto_exec_OPENSTACK       string
        auto_exec_VMWARE          string
        boddy_WEB                 string
        JDK                        string
        NGINX                      string
        TOMCAT                     string
}



func main() {

        fmt.Println(`This cloud deploy tools......
Program start......
This support CentOS or RedHat 7.3 and last
8Cpu 16Gmemory 200Gdisk
/ > 30G
/opt   >  30G
/home > 30G
need yum source is availability
port number :
    80  boddy  web
    8088  portal  web
    8081  WordFlow     
    61616 activemq     
    3306  mysql
    8443  portal  web https
Program will change files  list is :     /etc/hosts,   /etc/sysconfig/network,   /etc/selinux/config,    stop firewall,    disable selinux,   and add user to linux system.
Please press the any key to continue or Press Ctrl+C to Quit!!!...... [any key]`)
        fmt.Scanln()

        local_ip = get_local_ip()  //获取本机ip地址
        if local_ip == "" {
                local_ip = "127.0.0.1"
        }

        http.Handle("/", http.FileServer(rice.MustFindBox("media").HTTPBox()))  //go-rice 打包文件
		//go  1.16可内置包，需要升级代码，去掉 rice 包依赖
    go http.ListenAndServe(":10001", nil)

        //修改/etc/hosts文件函数
        /*host_context := "\n#cloud deploy tool config start\n" + local_ip + "\t\tportal01\tCloudAPPcloudportal\tactivemq01\tactivemq\tCloudAPPclouddb\tlocalhost\n127.0.0.1\t\tportal02\tactivemq02\n127.0.0.1\tdb01\n127.0.0.1\tdb02\n127.0.0.1\tdb03\n\n\n\n#HA host config\n#app01\n#10.10.10.8\tportal01\tactivemq01\tlocalhost\n#10.10.10.9\tportal02\tactivemq02\n#10.10.10.10\tdb01\n#10.10.10.11\tdb02\n#10.10.10.12\tdb03\n#app02\n#10.10.10.8\tportal01\tactivemq01\n#10.10.10.9\tportal02\tactivemq02\tlocalhost\n#10.10.10.10\tdb01\n#10.10.10.11\tdb02\n#10.10.10.12\tdb03\n"*/
        host_context := "# single config\n" + local_ip + "\tportal01\tactivemq01\tlocalhost\n" + local_ip + "\tportal02\tactivemq02\n"  + local_ip + "\tdb01\n" + local_ip + "\tdb02\n" + local_ip + "\tdb03\n\n" + "## ha config\n##app01\n#" + local_ip +    "\tportal01\tactivemq01\tlocalhost\n#" + local_ip + "\tportal02\tactivemq02\n#" + local_ip + "\tdb01\n#" + local_ip + "\tdb02\n#" + local_ip + "\tdb03\n##app02\n#" + local_ip + "\tportal01\tactivemq01\n#" + local_ip + "\tportal02\tactivemq02\tlocalhost\n#" + local_ip + "\tdb01\n#" + local_ip + "\tdb02\n#" + local_ip + "\tdb03\n\n#powervc ipaddress\n\n#huawei openstack ipaddress\n\n#HuaWei openstack  az1\n"

        //判断 /etc/hosts 文件是否已经包含  host_context 内容
        host_file, _ := os.Open("/etc/hosts")
        host_file_context, _ := ioutil.ReadAll(host_file)
        host_file.Close()
        if strings.Index(string(host_file_context), "portal01") == -1 {
                //添加 host_contxt 内容到 /etc/hosts
                add_txt_to_file("/etc/hosts", host_context)

                //修改 /etc/sysconfig/network
                fmt.Println("\nchange /etc/sysconfig/network ......")
                fmt.Println(tracefile("HOSTNAME=portal01","/etc/sysconfig/network",`HOSTNAME.*`))
                //os.Remove("/etc/sysconfig/network")
                fmt.Println("\nchange /etc/selinux/config ......")
                fmt.Println(tracefile("SELINUX=disabled","/etc/selinux/config",`SELINUX=.*`))
                // 关闭系统防火墙和selinux服务
                system_service()

                //重启动linux

                fmt.Print("\n------------------------------------------\n--  This Linux OS will reboot, Please Input 'y' to Reboot OR Any Key to Quit!!!\nPlease input : ")
                fmt.Scanln(&comfirm_string)
                if comfirm_string == "y" {
                        execCommand("/sbin/shutdown", []string{"-r", "now"})
                        time.Sleep(time.Second * 10)
                } else {
                        fmt.Println("Cloud Installer Quit. Please reboot by manual......")
                        os.Exit(1)
                }
        } else {
                fmt.Println("/etc/hosts check ok . Continue......")
        }

        deploy_tools_url := "http://" + local_ip + ":10001/doc"
        portal_usl := "http://" + local_ip + ":8088/CloudAPP-web"
        boddy_url := "http://" + local_ip
        Deploy_url := "http://" + local_ip + ":10001/"
        Doc_url := "http://" + local_ip + ":10001/doc"
        scripts_url := "http://" + local_ip + ":10001/scripts"
        media_url := "http://" + local_ip + ":10001"

        //循环菜单开始
        beging:
    for {
                fmt.Print("\n\n\n\n", time.Now().Format("2006-01-02 15:04:05"), "*********************************************************************\n*********************************************\n")
                fmt.Println("Document url :", deploy_tools_url, "\nDeploy software Download url : ", Deploy_url)
                fmt.Println("departments menu")
                fmt.Print("Input deploy type : \n10. single machine deploy all departments\n20. deploy url info\n33. start CloudAPP cloud\n44. stop CloudAPP cloud\n70. delete all departments\n99. Quit\nPlease select : ")
                deploy_type := "100"
                fmt.Scanln(&deploy_type)
                deploy_type_int,err:=strconv.Atoi(deploy_type) 
                if err != nil {
                        fmt.Println("\nLog=>",time.Now().Format("2006-01-02 15:04:05"),"**************************You input is :",deploy_type,"   *************Input Error!!!*************\nPlease input again.")
                        goto beging
                }
                if deploy_type_int == 10 {
                        fmt.Println("input : ",deploy_type_int)
                        single_deploy(local_ip)
                } else if deploy_type_int == 20 {
                        fmt.Println("\nDeploy url list is :\nAdmin url :", portal_usl, "\nboddy url :", boddy_url, "\n")
                        fmt.Println("\nDoc url is : ",Doc_url,"\nscripts url is : ", scripts_url, "\nall media url is : ", media_url , "\n")
                        goto beging
                } else if deploy_type_int == 33 {
                        stop_server()
                        run_server()
                        goto beging
                } else if deploy_type_int == 44 {
                        stop_server()
                        goto beging
                } else if deploy_type_int == 70 {
                        clear_all()
                        goto beging
                } else if deploy_type_int == 99 {
                        fmt.Println("\n... This cloud deploy tools Quit! ...\n")
                        os.Exit(0)
                } else {
                        fmt.Println("\nLog=>",time.Now().Format("2006-01-02 15:04:05"),"You input is :",deploy_type_int,"   *************Input Error!!!*************\nPlease input again.")
            }
    }
}


//单机部署脚本
func single_deploy(local_ip string) error {
        fmt.Println("deploy all departments")


        //创建CloudAPP,auto_exec,boddy,WordFlow用户
        user_list := []string{"CloudAPP","WordFlow","boddy","activemq","CloudAPPauto"}
        for username := range user_list {
                name := []string{user_list[username]}
                //fmt.Println(name)
                if adduser(name) {
                        fmt.Println("    create user : ",name,"    ---> Success")
                } else {
                        fmt.Println("    create user : ",name,"    --->  fail,please check")
                        os.Exit(1)
                }
        }

        // files_name 需要下载的文件
        files_name := []string{
                "/home/activemq.tar.gz",
                "/home/CloudAPPauto.tar.gz",
                "/home/WordFlow.tar.gz",
                "/home/CloudAPP.tar.gz",
                "/home/boddy.tar.gz",
                "/opt/jdk1.8.0_131.tar.gz",
                "/home/CloudAPP-boddy-ui.tar.gz",
        }
        for _, file := range files_name {
                download_file(local_ip,file)
        }

        //安装mysql
        install_mysql(local_ip)

        //修改目录权限
        dir_name_list := []string{
                "/home/CloudAPPauto",
                "/home/CloudAPP",
                "/home/WordFlow",
                "/home/boddy",
                "/home/activemq",
        }
        for _, dir_name := range dir_name_list {
                err := os.Chmod(dir_name,0777)
                if err != nil {
                        fmt.Println("chmod -R 777 ",dir_name,err,"\t\t\t\t\t\t\t\t---> fail,please check")
                } else {
                        fmt.Println("chmod -R 777 ",dir_name,"\t\t\t\t\t\t\t\t---> Success")
                }
                username := path.Base(dir_name)
                chown(username,dir_name)
        }

        os.Chmod("/opt/jdk1.8.0_131",0777)
        os.Chmod("/home/CloudAPP-boddy-ui",0777)
        chown("root","/opt/jdk1.8.0_131")
        chown("boddy","/home/CloudAPP-boddy-ui")

        //部署 nginx 和 boddy-ui
        if deploy_nginx_boddy_ui() {
                fmt.Println("nginx and boddy-ui deploy ---> Success")
        } else {
                fmt.Println("nginx and boddy-ui deploy ---> Fail......")
        }

        //获取部署包名称
        cloud_conf := "http://" + local_ip + ":10001/opt/cloud_pkg.conf"
        get_url_file(cloud_conf,"/opt/cloud_pkg.conf")
        data, err := ioutil.ReadFile("/opt/cloud_pkg.conf")
        if err != nil {
                        fmt.Print(err)
        }
        // json data
        var obj app_file
        // unmarshall it
        err = json.Unmarshal(data, &obj)
        if err != nil {
                        fmt.Println("error:", err)
        }
        // can access using struct now
        //fmt.Printf(obj.CloudAPP_WEB_PKG);


        //移动 /opt/CloudAPP_web.war 到部署目录 /home/CloudAPP/tomcat-web/webapps/
        CloudAPP_web := "http://" + local_ip + ":10001/opt/" + obj.CloudAPP_WEB_PKG
        if get_url_file(CloudAPP_web, "/home/CloudAPP/tomcat-web/webapps/CloudAPP-web.war") {
                fmt.Println("download  /home/WordFlow/tomcat-web/webapps/CloudAPP-web.war ---->  Success")
        } else {
                fmt.Println("download  /home/WordFlow/tomcat-web/webapps/CloudAPP-web.war ---->  Fail......")
        }
        //移动 /opt/WordFlow.war 到部署目录 /home/WordFlow/tomcat-WordFlow/webapps/
        CloudAPP_WordFlow := "http://" + local_ip + ":10001/opt/" + obj.WordFlow_PKG
        if get_url_file(CloudAPP_WordFlow, "/home/WordFlow/tomcat-WordFlow/webapps/CloudAPP-WordFlow.war") {
                fmt.Println("download  /home/WordFlow/tomcat-WordFlow/webapps/CloudAPP-WordFlow.war  ----> Success")
        } else {
                fmt.Println("download  /home/WordFlow/tomcat-WordFlow/webapps/CloudAPP-WordFlow.war  ----> Fail......")
        }

        // boddy ui 更新下载 /opt/dist.zip 文件
        boddy_ui_setup_url := "http://" + local_ip + ":10001/opt/" + obj.UI_PKG
        if get_url_file(boddy_ui_setup_url, "/opt/html.tar.gz") {
                fmt.Println("\ndownload ", boddy_ui_setup_url, " to /opt/html.tar.gz ---> update Success")
        } else {
                fmt.Println("\ndownload ", boddy_ui_setup_url, " to /opt/html.tar.gz ---> update Fail......")
        }

        // 从/opt 更新boddy租户ui页面  #/usr/bin/tar -xzvf /opt/html.tar.gz  -C /usr/share/nginx/
        execCommand("/usr/bin/tar", []string{"-xzvf", "/opt/html.tar.gz", "-C", "/usr/share/nginx/"})

        // boddy 后台 更新下载 /home/boddy/CloudAPP-boddy-0.0.1-SNAPSHOT.jar 文件
        boddy_jar_setup_url := "http://" + local_ip + ":10001/opt/" + obj.boddy_JAR
        boddy_jar_setup_path := "/home/boddy/CloudAPP-boddy-0.0.1-SNAPSHOT.jar"
        if get_url_file(boddy_jar_setup_url, boddy_jar_setup_path) {
                fmt.Println("\ndownload ", boddy_jar_setup_url, " to ", boddy_jar_setup_path, "---> update Success")
        } else {
                fmt.Println("\ndownload ", boddy_jar_setup_url, " to ", boddy_jar_setup_path, " ---> update Fail......")
        }

        // vmware_auto_exec 更新下载
        vmware_auto_exec := "http://" + local_ip + ":10001/opt/" + obj.auto_exec_VMWARE
        vmware_auto_exec_path :="/home/CloudAPPauto/CloudAPP-auto_exec-vmware/lib/CloudAPP-auto_exec-vmware-0.0.5.jar"
        if get_url_file(vmware_auto_exec, vmware_auto_exec_path) {
                fmt.Println("\ndownload ", vmware_auto_exec, " to ",vmware_auto_exec_path , "---> update Success")
        } else {
                fmt.Println("\ndownload ", vmware_auto_exec, " to ",vmware_auto_exec_path , "---> update Fail......")
        }

        // ssh_auto_exec 更新下载
        ssh_auto_exec := "http://" + local_ip + ":10001/opt/" + obj.auto_exec_SSH
        ssh_auto_exec_path := "/home/CloudAPPauto/CloudAPP-auto_exec-ssh/lib/CloudAPP-auto_exec-ssh-0.0.1.jar"
        if get_url_file(ssh_auto_exec, ssh_auto_exec_path) {
                fmt.Println("\ndownload ", ssh_auto_exec, " to ", ssh_auto_exec_path , "---> update Success")
        } else {
                fmt.Println("\ndownload ", ssh_auto_exec, " to ", ssh_auto_exec_path , "---> update Fail......")
        }

        // openstack_auto_exec 更新下载
        openstack_auto_exec := "http://" + local_ip + ":10001/opt/" + obj.auto_exec_OPENSTACK
        openstack_auto_exec_path := "/home/CloudAPPauto/CloudAPP-auto_exec-openstack/CloudAPP-auto_exec-openstack-1.0.jar"
        if get_url_file(openstack_auto_exec, openstack_auto_exec_path) {
                fmt.Println("\ndownload ", openstack_auto_exec, " to ", openstack_auto_exec_path, "---> update Success")
        } else {
                fmt.Println("\ndownload ", openstack_auto_exec, " to ", openstack_auto_exec_path, "---> update Fail......")
        }

        // powervc_auto_exec 更新下载
        powervc_auto_exec := "http://" + local_ip + ":10001/opt/" + obj.auto_exec_POWERVC
        powervc_auto_exec_path := "/home/CloudAPPauto/CloudAPP-auto_exec-powervc/CloudAPP-auto_exec-powervc-1.0.jar"
        if get_url_file(powervc_auto_exec, powervc_auto_exec_path) {
                fmt.Println("\ndownload ", powervc_auto_exec, " to ", powervc_auto_exec_path , "---> update Success")
        } else {
                fmt.Println("\ndownload ", powervc_auto_exec, " to ", powervc_auto_exec_path , "---> update Fail......")
        }

        // power_auto_exec 更新下载
        power_auto_exec := "http://" + local_ip + ":10001/opt/" + obj.auto_exec_POWER
        power_auto_exec_path := "/home/CloudAPPauto/CloudAPP-auto_exec-power/lib/CloudAPP-auto_exec-power-0.0.1.jar"
        if get_url_file(power_auto_exec, power_auto_exec_path) {
                fmt.Println("\ndownload ", power_auto_exec, " to ", power_auto_exec_path , "---> update Success")
        } else {
                fmt.Println("\ndownload ", power_auto_exec, " to ", power_auto_exec_path , "---> update Fail......")
        }

        os.Chmod("/CloudAPP/CloudAPPauto",0777)
        return nil
}

func Exist(filename string) bool {
        _, err := os.Stat(filename)
        return err == nil || os.IsExist(err)
}

//关闭系统进程
func system_service() error {
        if Exist("/usr/bin/systemctl") {
                //redhat 7.x
                if execCommand("/usr/sbin/setenforce",[]string{"0"}) {
                        fmt.Println("\t\t\t\t\t\t\t\t---> Success")
                } else {
                        fmt.Println("\t\t\t\t\t\t\t\t--->  fail,please check")
                }
                //关闭防火墙
                execCommand("/usr/bin/systemctl",[]string{"stop","firewalld.service"})
                if execCommand("/usr/bin/systemctl",[]string{"disable","firewalld.service"}) {
                        fmt.Println("\t\t\t\t\t\t\t\t---> Success")
                } else {
                        fmt.Println("\t\t\t\t\t\t\t\t---> fail,please check")
                }
                //关闭 iptable
                execCommand("/usr/bin/systemctl",[]string{"stop","iptables.service"})
                if execCommand("/usr/bin/systemctl",[]string{"disable","iptables.service"}) {
                        fmt.Println("\t\t\t\t\t\t\t\t---> Success")
                } else {
                        fmt.Println("\t\t\t\t\t\t\t\t---> fail,please check")
                }
        } else {
                // redhat 6.x
                if execCommand("/usr/sbin/setenforce",[]string{"0"}) {
                        fmt.Println("\t\t\t\t\t\t\t\t---> Success")
                } else {
                        fmt.Println("\t\t\t\t\t\t\t\t--->  fail,please check")
                }
                //关闭防火墙
                execCommand("/sbin/chkconfig",[]string{"iptables","off"})
                execCommand("/sbin/chkconfig",[]string{"ip6tables","off"})
                if execCommand("/sbin/service",[]string{"iptables","stop"}) {
                        fmt.Println("\t\t\t\t\t\t\t\t---> Success")
                } else {
                        fmt.Println("\t\t\t\t\t\t\t\t---> fail,please check")
                }
        }
        return nil
}

//获取本机ip地址
func get_local_ip() string {
        addrs, err := net.InterfaceAddrs()
        if err != nil {
                fmt.Println(err)
                os.Exit(1)
        }
        var local_ip string
        for _, address := range addrs {
                // 检查ip地址判断是否回环地址
                if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
                        if ipnet.IP.To4() != nil {
                                local_ip = ipnet.IP.String()
                                return ipnet.IP.String()
                        }
                }
        }
        return local_ip
}

//创建linux用户
func adduser(params []string) bool {
        //fmt.Println("adduser func",params)
        if execCommand("/usr/sbin/adduser",params) {
                return true
        } else {
                return false
        }
}


//下载文件方法
func download_file(ip_hostname,file_name string) error {
        file_url := "http://" + ip_hostname + ":10001" + file_name
        if get_url_file(file_url,file_name) {
                fmt.Println("\ndownload ",file_url," to ", file_name,"   ---> Success")
                if !tarfile(file_name) {
                        fmt.Println("tar -xzf ",file_name," fail!!!")
                        os.Exit(1)
                } else {
                        //删除tar.gz文件
                        os.Remove(file_name)
                        return nil
                }
        } else {
                fmt.Println("\ndownload ",file_url," to ", file_name,"   ---> fail,please check")
                os.Exit(1)
        }
        return nil
}

func tarfile(tarfilename string) bool {
        dir_name := path.Dir(tarfilename)
        cmd_args := []string{"-xzf",tarfilename,"-C",dir_name}
        if execCommand("/bin/tar",cmd_args) {
                fmt.Println("\t\t\t\t\t\t\t\t---> Success")
                return true
        } else {
                fmt.Println("\t\t\t\t\t\t\t\t---> fail,please check")
                return false
        }
}

//下载文件到指定目录
func get_url_file(url , local_file string) bool{  
    res, err := http.Get(url)  
    if err != nil {  
                fmt.Println(err)
                return false
    }  
    f, err := os.Create(local_file)  
    if err != nil {  
                fmt.Println(err) 
                return false
    }  
        io.Copy(f, res.Body) 
        return true 
}  


//组件部署脚本
func departmen_deploy_menu(number int) {
        fmt.Println("deploy ",number," departments")
}

//命令执行函数
func execCommand(commandName string, params []string) bool {
    //函数返回一个*Cmd，用于使用给出的参数执行name指定的程序
    cmd := exec.Command(commandName, params...)
    //显示运行的命令
    fmt.Print("\nRun cmd : ",cmd.Args)
    out, err := cmd.CombinedOutput()  
    if err != nil {  
                fmt.Println(err)  
                return false
    }  
        fmt.Println(string(out))  
        return true
}


//添加内容到文件末尾函数
func add_txt_to_file(path_filename string, text_context string) error {
        // 以只写的模式，打开文件
        host_file, _ := os.Open(path_filename)
        host_file_context, _ := ioutil.ReadAll(host_file)
        if strings.Index(string(host_file_context), text_context) != -1 {
                return nil
        } 
        f, err := os.OpenFile(path_filename, os.O_WRONLY, 0644)
        defer f.Close()
        if err != nil {
                fmt.Println("path_filename file create failed. err: " + err.Error())
        } else {
           // 查找文件末尾的偏移量
                n, _ := f.Seek(0, os.SEEK_END)
                // 从末尾的偏移量开始写入内容
                _, err = f.WriteAt([]byte(text_context), n)
        }
        return err
}

func chown(user,path string) {
        user_group := user+":"+user
        path_list := []string{"-R",user_group,path}
        if execCommand("/bin/chown",path_list) {
                fmt.Println("/bin/chown -R ",user_group," ",path,"\t\t------> Success")
        } else {
                fmt.Println("/bin/chown -R ",user_group," ",path,"\t\t------> fail,please check")
        }
        // path_list = []string{"-R","755",path}
        // if execCommand("/bin/chmod",path_list) {
        //      fmt.Println("/bin/chmod -R 755 ",path,"\t\t------> Success")
        // } else {
        //      fmt.Println("/bin/chmod -R 755 ",path,"\t\t------> fail,please check")
        // }
}


//获得执行文件路径函数
func getCurrentPath() (string, error) {
    file, err := exec.LookPath(os.Args[0])
    if err != nil {
        return "", err
    }
    path, err := filepath.Abs(file)
    if err != nil {
        return "", err
    }
    i := strings.LastIndex(path, "/")
    if i < 0 {
        i = strings.LastIndex(path, "\\")
    }
    if i < 0 {
        return "", errors.New(`error: Can't find "/" or "\".`)
    }
    return string(path[0 : i+1]), nil
}

//安装mysql rpm包 
func install_mysql(ip_hostname string) error{
        if Exist("/usr/bin/systemctl") {
                // redhat 7.x
                //卸载 mysql-libs 包  rpm -ev mysql-libs-5.1.73-3.el6_5.x86_64   --nodeps #yum erase mariadb-libs
                execCommand("/usr/bin/yum", []string{"erase", "-y", "mariadb-libs.x86_64"})
                execCommand("/usr/bin/yum", []string{"erase", "-y", "mariadb-libs.i686"})
                execCommand("/usr/bin/yum", []string{"erase", "-y", "mysql-community-common"})
                execCommand("/bin/rpm", []string{"-e", "mariadb-libs-*", "--nodeps"})
                //先卸载 MySQL-server和MySQL-client rpm包
                execCommand("/bin/rpm", []string{"-e", "MySQL-server"})
                execCommand("/bin/rpm", []string{"-e", "MySQL-client"})
                //安装mysql server 和 client
                execCommand("/usr/bin/yum", []string{"install","-y", "perl-Module-Install.noarch"})
                rpm_pkg_names := []string{"/mysql/MySQL-client-5.6.40-1.el7.x86_64.rpm", "/mysql/MySQL-server-5.6.40-1.el7.x86_64.rpm"}
                for _, rpm_pkg_name := range rpm_pkg_names {
                        rpm_pkg_url := "http://" + ip_hostname + ":10001" + rpm_pkg_name
                        rpm_pkg_url_slice := []string{"-i", "-v", "-h", rpm_pkg_url}
                        if execCommand("/bin/rpm", rpm_pkg_url_slice) {
                                fmt.Println("/bin/rpm -i -v -h ", rpm_pkg_url, "\t\t\t\t\t\t\t\t---> Success")
                        } else {
                                fmt.Println("/bin/rpm -i -v -h ", rpm_pkg_url, "\t\t\t\t\t\t\t\t---> fail,please check")
                                os.Exit(1)
                        }
                }
                //下载并导入数据库文件
                CloudAPP_init_file := "http://" + ip_hostname + ":10001/mysql/CloudAPP_init.sql"
                if get_url_file(CloudAPP_init_file, "/tmp/CloudAPP_init.sql") {
                        fmt.Println("\ndownload ", CloudAPP_init_file, " to /tmp/CloudAPP_init.sql ---> Success")
                } else {
                        fmt.Println("\ndownload ", CloudAPP_init_file, " to /tmp/CloudAPP_init.sql ---> Fail......")
                }
                //配置 mysql
                // /usr/my.cnf 内容
                var my_cnf_context string
                my_cnf_context = `[mysqld]
character-set-server=utf8
sql_mode=NO_ENGINE_SUBSTITUTION,STRICT_TRANS_TABLES
lower_case_table_names=1
max_allowed_packet=256M
wait_timeout=30000

[client]
default-character-set=utf8

[mysql]
default-character-set=utf8
`
                //写入 /usr/my.cnf 文件
                my_conf_err := ioutil.WriteFile("/usr/my.cnf", []byte(my_cnf_context), 0)
                if my_conf_err != nil {
                        fmt.Println("/usr/my.cnf write\t\t\t\t\t\t\t\t----> fail,please check. ", my_conf_err)
                } else {
                        fmt.Println("/usr/my.cnf write\t\t\t\t\t\t\t\t----> Success")
                }

                //配置脚本 下载 /mysql/mysql_setup.sh 文件
                mysql_setup_url := "http://" + ip_hostname + ":10001/mysql/mysql_setup.sh"
                if get_url_file(mysql_setup_url, "/tmp/mysql_setup.sh") {
                        fmt.Println("\ndownload ", mysql_setup_url, " to/tmp/mysql_setup.sh ---> Success")
                } else {
                        fmt.Println("\ndownload ", mysql_setup_url, " to /tmp/mysql_setup.sh ---> Fail......")
                }
                //执行/tmp/mysql_setup.sh配置脚本
                fmt.Println("\n\nImport SQL_File to mysqldatabase, Please Wait about 10 minutes......\n")
                fmt.Println("In the RedHat7 or CentOS7 environment, Import SQL_File to mysqldatabase step. \nIf it takes more than 10 minutes,\nPlease open another console and restart mysql with the [ service mysql restart ] or [ systemctl restart mysql ] command \nto keep the installer running.")
                if execCommand("/usr/bin/bash", []string{"/tmp/mysql_setup.sh", "&"}) {
                        fmt.Println(time.Now().Format("2006-01-02 15:04:05"), " import CloudAPP_init.sql to CloudAPPdb  ----> Success")
                } else {
                        fmt.Println(time.Now().Format("2006-01-02 15:04:05"), " import CloudAPP_init.sql to CloudAPPdb ---->  Fail......")
                }
        } else {
                // redhat 6.x
                //卸载 mysql-libs 包  rpm -ev mysql-libs-5.1.73-3.el6_5.x86_64   --nodeps
                execCommand("/usr/bin/yum", []string{"erase", "-y", "mysql-libs-5*"})
                execCommand("/bin/rpm", []string{"-e", "mysql-libs-5*", "--nodeps"})
                //先卸载 MySQL-server和MySQL-client rpm包
                execCommand("/bin/rpm", []string{"-e", "MySQL-server"})
                execCommand("/bin/rpm", []string{"-e", "MySQL-client"})
                //安装mysql server 和 client
                rpm_pkg_names := []string{"/mysql/MySQL-client-5.6.24-1.el6.x86_64.rpm", "/mysql/MySQL-server-5.6.24-1.el6.x86_64.rpm"}
                for _, rpm_pkg_name := range rpm_pkg_names {
                        rpm_pkg_url := "http://" + ip_hostname + ":10001" + rpm_pkg_name
                        rpm_pkg_url_slice := []string{"-i", "-v", "-h", rpm_pkg_url}
                        if execCommand("/bin/rpm", rpm_pkg_url_slice) {
                                fmt.Println("/bin/rpm -i -v -h ", rpm_pkg_url, "\t\t\t\t\t\t\t\t---> Success")
                        } else {
                                fmt.Println("/bin/rpm -i -v -h ", rpm_pkg_url, "\t\t\t\t\t\t\t\t---> fail,please check")
                                os.Exit(1)
                        }
                }

                //下载并导入数据库文件
                CloudAPP_init_file := "http://" + ip_hostname + ":10001/mysql/CloudAPP_init.sql"
                if get_url_file(CloudAPP_init_file, "/tmp/CloudAPP_init.sql") {
                        fmt.Println("\ndownload ", CloudAPP_init_file, " to /tmp/CloudAPP_init.sql ---> Success")
                } else {
                        fmt.Println("\ndownload ", CloudAPP_init_file, " to /tmp/CloudAPP_init.sql ---> Fail......")
                }

                //配置 mysql
                // /usr/my.cnf 内容
                var my_cnf_context string
                my_cnf_context = `[mysqld]
character-set-server=utf8
sql_mode=NO_ENGINE_SUBSTITUTION,STRICT_TRANS_TABLES
lower_case_table_names=1

[client]
default-character-set=utf8

[mysql]
default-character-set=utf8
`
                //写入 /usr/my.cnf 文件
                my_conf_err := ioutil.WriteFile("/usr/my.cnf", []byte(my_cnf_context), 0)
                if my_conf_err != nil {
                        fmt.Println("/usr/my.cnf write\t\t\t\t\t\t\t\t----> fail,please check. ", my_conf_err)
                } else {
                        fmt.Println("/usr/my.cnf write\t\t\t\t\t\t\t\t----> Success")
                }

                //配置脚本 下载 /mysql/mysql_setup.sh 文件
                mysql_setup_url := "http://" + ip_hostname + ":10001/mysql/mysql_setup.sh"
                if get_url_file(mysql_setup_url, "/tmp/mysql_setup.sh") {
                        fmt.Println("\ndownload ", mysql_setup_url, " to/tmp/mysql_setup.sh ---> Success")
                } else {
                        fmt.Println("\ndownload ", mysql_setup_url, " to /tmp/mysql_setup.sh ---> Fail......")
                }

                //执行/tmp/mysql_setup.sh配置脚本
                fmt.Println("\n\nImport SQL_File to mysqldatabase, Please Wait about 10 minutes......")
                if execCommand("/bin/bash", []string{"/tmp/mysql_setup.sh", "&"}) {
                        fmt.Println(time.Now().Format("2006-01-02 15:04:05"), " import CloudAPP_init.sql to CloudAPPdb  ----> Success")
                } else {
                        fmt.Println(time.Now().Format("2006-01-02 15:04:05"), " import CloudAPP_init.sql to CloudAPPdb ---->  Fail......")
                }
        }
        return nil
}


//部署nginx
func deploy_nginx_boddy_ui() bool {
        //部署nginx
        //if execCommand("/usr/bin/yum",[]string{"install","-y","/home/CloudAPP-boddy-ui/nginx-1.14.0-1.el6.ngx.x86_64.rpm"}) {
        //      fmt.Println("yum install -y /home/CloudAPP-boddy-ui/nginx-1.14.0-1.el6.ngx.x86_64.rpm ---> Success")
        //} else {
        //      fmt.Println("yum install -y /home/CloudAPP-boddy-ui/nginx-1.14.0-1.el6.ngx.x86_64.rpm ---> Fail.please check")
        //      return false
        //}
        if Exist("/usr/bin/systemctl") {
                // redhat 7.x
                execCommand("/usr/bin/rpm", []string{"-e", "nginx-1.*"})
                if execCommand("/usr/bin/rpm", []string{"-ivh", "/home/CloudAPP-boddy-ui/nginx-1.8.1-1.el7.ngx.x86_64.rpm"}) {
                        fmt.Println("/usr/bin/rpm -ivh /home/CloudAPP-boddy-ui/nginx-1.8.1-1.el7.ngx.x86_64.rpm ---> Success")
                } else {
                        fmt.Println("/usr/bin/rpm -ivh /home/CloudAPP-boddy-ui/nginx-1.8.1-1.el7.ngx.x86_64.rpm ---> Fail......")
                }
                if execCommand("/usr/bin/rm", []string{"-rf", "/usr/share/nginx/html"}) {
                        fmt.Println("/usr/bin/rm -rf /usr/share/nginx/html ----> Success")
                } else {
                        fmt.Println("/usr/bin/rm -rf /usr/share/nginx/html ---->  Fail......")
                }
                if execCommand("/usr/bin/mv", []string{"/home/CloudAPP-boddy-ui/html", "/usr/share/nginx/"}) {
                        fmt.Println("/usr/bin/mv /home/CloudAPP-boddy-ui/html /usr/share/nginx/    ----> Success")
                } else {
                        fmt.Println("/usr/bin/mv /home/CloudAPP-boddy-ui/html /usr/share/nginx/    ----> Fail......")
                }
                if execCommand("/usr/bin/mv", []string{"/home/CloudAPP-boddy-ui/nginx.conf", "/etc/nginx/nginx.conf"}) {
                        fmt.Println("/usr/bin/mv /home/CloudAPP-boddy-ui/nginx.conf /etc/nginx/nginx.conf ----> Success")
                } else {
                        fmt.Println("/usr/bin/mv /home/CloudAPP-boddy-ui/nginx.conf /etc/nginx/nginx.conf ----> Fail......")
                }
        } else {
                //redhat 6.x
                execCommand("/bin/rpm", []string{"-e", "nginx-1.*"})
                if execCommand("/bin/rpm", []string{"-ivh", "/home/CloudAPP-boddy-ui/nginx-1.8.1-1.el6.ngx.x86_64.rpm"}) {
                        fmt.Println("/bin/rpm -ivh /home/CloudAPP-boddy-ui/nginx-1.8.1-1.el6.ngx.x86_64.rpm ---> Success")
                } else {
                        fmt.Println("/bin/rpm -ivh /home/CloudAPP-boddy-ui/nginx-1.8.1-1.el6.ngx.x86_64.rpm ---> Fail......")
                }
                if execCommand("/bin/rm", []string{"-rf", "/usr/share/nginx/html"}) {
                        fmt.Println("/bin/rm -rf /usr/share/nginx/html ----> Success")
                } else {
                        fmt.Println("/bin/rm -rf /usr/share/nginx/html ---->  Fail......")
                }
                if execCommand("/bin/mv", []string{"/home/CloudAPP-boddy-ui/html", "/usr/share/nginx/"}) {
                        fmt.Println("/bin/mv /home/CloudAPP-boddy-ui/html /usr/share/nginx/    ----> Success")
                } else {
                        fmt.Println("/bin/mv /home/CloudAPP-boddy-ui/html /usr/share/nginx/    ----> Fail......")
                }
                if execCommand("/bin/mv", []string{"/home/CloudAPP-boddy-ui/nginx.conf", "/etc/nginx/nginx.conf"}) {
                        fmt.Println("/bin/mv /home/CloudAPP-boddy-ui/nginx.conf /etc/nginx/nginx.conf ----> Success")
                } else {
                        fmt.Println("/bin/mv /home/CloudAPP-boddy-ui/nginx.conf /etc/nginx/nginx.conf ----> Fail......")
                }
        }
        return true
}

//清理环境
func clear_all() bool{
        //删除用户
        user_list := []string{"CloudAPP","WordFlow","boddy","activemq","CloudAPPauto"}
        for _, user := range user_list {
                args := []string{"-r",user}
                if execCommand("/usr/sbin/userdel",args) {
                        fmt.Println("userdel -r ",user,"\t\t\t\t\t\t\t\t---> Success")
                } else {
                        fmt.Println("userdel -r ",user,"\t\t\t\t\t\t\t\t---> fail,please check")
                }
        }

        //删除mysql
        mysql_server_client_rpms := []string{"-e","MySQL-server"}
        if execCommand("/bin/rpm",mysql_server_client_rpms) {
                fmt.Println("/bin/rpm -e MySQL-server\t\t\t\t\t\t\t\t---> Success")
        } else {
                fmt.Println("/bin/rpm -e MySQL-server\t\t\t\t\t\t\t\t---> fail,please check")
        }
        mysql_server_client_rpms = []string{"-e","MySQL-client"}
        if execCommand("/bin/rpm",mysql_server_client_rpms) {
                fmt.Println("/bin/rpm -e MySQL-client\t\t\t\t\t\t\t\t---> Success")
        } else {
                fmt.Println("/bin/rpm -e MySQL-client\t\t\t\t\t\t\t\t---> fail,please check")
        }

        //清理无关目录
        CloudAPP_dirs := []string{
                "/usr/lib64/mysql",
                "/var/lib/mysql",
                "/var/lib/mysql/mysql",
                "/home/activemq",
                "/home/CloudAPPauto",
                "/home/WordFlow",
                "/home/CloudAPP",
                "/home/boddy",
                "/home/CloudAPP-boddy-ui",
                "/usr/share/nginx/html",
                "/usr/share/nginx/html.bak",
        }
        for _, CloudAPP_dir := range CloudAPP_dirs {
                mysql_dir_err := os.RemoveAll(CloudAPP_dir)
                if mysql_dir_err != nil {
                        fmt.Println("/bin/rm -rf",CloudAPP_dir,"\t\t\t\t\t\t\t\t---> fail,please check")
                } else {
                        fmt.Println("/bin/rm -rf",CloudAPP_dir,"\t\t\t\t\t\t\t\t---> Success")
                }
        }

        //删除相关文件
        del_files_list := []string{
                // "/home/activemq.tar.gz",
                // "/home/CloudAPPauto.tar.gz",
                // "/home/WordFlow.tar.gz",
                // "/home/CloudAPP.tar.gz",
                // "/home/boddy.tar.gz",
                "/tmp/mysql_setup.sh",
                "/tmp/CloudAPP_init.sql",
                "/opt/html.tar.gz",
                "/tmp/kill_all_CloudAPP_process.sh",
                "/tmp/powervc_run_scripts.sh",
                "/tmp/openstack_run_scripts.sh",
                "/tmp/boddy_run_scripts.sh",
                "/opt/cloud_pkg.conf",
        }
        for _, file_name := range del_files_list {
                err := os.Remove(file_name)
                if err != nil {
                        fmt.Println("/bin/rm -rf ",file_name,"\t\t\t\t\t\t\t\t---> fail,please check")
                } else {
                        fmt.Println("/bin/rm -rf ",file_name,"\t\t\t\t\t\t\t\t---> Success")
                }
        }

        execCommand("/bin/rpm",[]string{"-e","nginx-1.*"})

        return true
}


//启动cloud
func run_server() error {
        //启动 activemq 
        args := []string{`-`,`activemq`,`-c`,`/home/activemq/apache-activemq-5.8.0/bin/activemq start`}
        if execCommand("/bin/su",args) {
                fmt.Println("start activemq    --->    Success")
        } else {
                fmt.Println("start activemq    --->    Fail......")
        }
        fmt.Println("log file is     --->    /home/activemq/apache-activemq-5.8.0/data/activemq.log")
        time.Sleep(time.Second * 5)

        //启动 CloudAPP-web 和 CloudAPP-WordFlow
        args = []string{`-`,`WordFlow`, `-c`, `/home/WordFlow/tomcat-WordFlow/bin/startup.sh`}
        if execCommand("/bin/su",args) {
                fmt.Println("start WordFlow    --->    Success")
        } else {
                fmt.Println("start WordFlow    --->    Fail......")
        }
        fmt.Println("log file is    --->    /home/WordFlow/tomcat-WordFlow/logs/catalina.out")
        time.Sleep(time.Second * 5)

        args = []string{`-`, `CloudAPP`, `-c`, `/home/CloudAPP/tomcat-web/bin/startup.sh`}
        if execCommand("/bin/su",args) {
                fmt.Println("start portal    --->    Success")
        } else {
                fmt.Println("start portal    --->    Fail......")
        }
        fmt.Println("log file is    --->    /home/CloudAPP/tomcat-web/logs/catalina.out")
        time.Sleep(time.Second * 5)

        //启动 auto_exec
        auto_exec_run := []string{
                `cd /home/CloudAPPauto/CloudAPP-auto_exec-power/bin/;./startsvr.sh ../conf/config.xml`,
                `cd /home/CloudAPPauto/CloudAPP-auto_exec-vmware/bin/;./startsvr.sh ../conf/config_vmware_bj.xml`,
                `cd /home/CloudAPPauto/CloudAPP-auto_exec-ssh/bin/;./startsvr.sh ../conf/config_ssh_bj.xml`,
        }
        for _, auto := range auto_exec_run {
                args_auto := []string{`-`, `CloudAPPauto`,`-c`,auto}
                if execCommand("/bin/su",args_auto) {
                        fmt.Println("/bin/su - CloudAPPauto -c ",args_auto," ---> Success")
                } else {
                        fmt.Println("/bin/su - CloudAPPauto -c ",args_auto," ---> fail,please check")
                }
                auto_path := "log file is    --->    /home/CloudAPPauto/" + strings.Split(auto,"/")[3] + "/logs/nohup.log"
                fmt.Println(auto_path)
                time.Sleep(time.Second * 5)
        }

        ioutil.WriteFile("/tmp/spring.log", []byte("...... starting ......"), 0777)
        os.Chmod("/tmp/spring.log",0777)

        //启动 powervc auto_exec
        powervc_run_scripts := `#!/bin/sh
/bin/su - CloudAPPauto -c "cd /home/CloudAPPauto/CloudAPP-auto_exec-powervc/;nohup  /opt/jdk1.8.0_131/bin/java -jar -Dspring.config.location=/home/CloudAPPauto/CloudAPP-auto_exec-powervc/application.properties CloudAPP-auto_exec-powervc-1.0.jar  > /home/CloudAPPauto/CloudAPP-auto_exec-powervc/jar.log 2>&1 & " > /dev/null
exit 0
`
        powervc_run_scripts_err := ioutil.WriteFile("/tmp/powervc_run_scripts.sh", []byte(powervc_run_scripts), 0755)
        if powervc_run_scripts_err != nil {
                fmt.Println("/tmp/powervc_run_scripts.sh  write\t\t\t\t\t\t\t\t----> Fail......")
        } else {
                fmt.Println("/tmp/powervc_run_scripts.sh  write\t\t\t\t\t\t\t\t----> Success")
        }
        exec.Command("/bin/sh","/tmp/powervc_run_scripts.sh").Run()
        time.Sleep(time.Second * 5)

        //启动 openstack auto_exec
        openstack_run_scripts := `#!/bin/sh
/bin/su - CloudAPPauto -c "cd /home/CloudAPPauto/CloudAPP-auto_exec-openstack/;nohup /opt/jdk1.8.0_131/bin/java -jar -Dspring.config.location=/home/CloudAPPauto/CloudAPP-auto_exec-openstack/application.properties CloudAPP-auto_exec-openstack-1.0.jar  > /home/CloudAPPauto/CloudAPP-auto_exec-openstack/jar.log 2>&1 &" > /dev/null
exit 0
`
        openstack_run_scripts_err := ioutil.WriteFile("/tmp/openstack_run_scripts.sh", []byte(openstack_run_scripts), 0755)
        if openstack_run_scripts_err != nil {
                fmt.Println("/tmp/openstack_run_scripts.sh  write\t\t\t\t\t\t\t\t----> Fail......")
        } else {
                fmt.Println("/tmp/openstack_run_scripts.sh  write\t\t\t\t\t\t\t\t----> Success")
        }
        exec.Command("/bin/sh","/tmp/openstack_run_scripts.sh").Run()
        time.Sleep(time.Second * 5)

        //启动 boddy 租户进程
        boddy_run_scripts := `#!/bin/sh
/bin/su - boddy -c "cd /home/boddy/;nohup /opt/jdk1.8.0_131/bin/java -jar -Dspring.config.location=/home/boddy/application.yml,/home/boddy/cloud-link.yml,/home/boddy/monitor.yml CloudAPP-boddy-0.0.1-SNAPSHOT.jar  > /home/boddy/jar.log 2>&1 &" > /dev/null
exit 0
`
        boddy_run_scripts_err := ioutil.WriteFile("/tmp/boddy_run_scripts.sh", []byte(boddy_run_scripts), 0755)
        if boddy_run_scripts_err != nil {
                fmt.Println("/tmp/boddy_run_scripts.sh  write\t\t\t\t\t\t\t\t----> Fail......")
        } else {
                fmt.Println("/tmp/boddy_run_scripts.sh  write\t\t\t\t\t\t\t\t----> Success")
        }
        exec.Command("/bin/sh","/tmp/boddy_run_scripts.sh").Run()
        time.Sleep(time.Second * 5)

        //启动 nginx  boddy-ui
        if execCommand("/usr/sbin/nginx",[]string{"-c","/etc/nginx/nginx.conf"}) {
                fmt.Println("nginx start    ---->    Success ")
        } else  {
                fmt.Println("nginx start    ---->    Fail......")
        }
        fmt.Println("log file is    --->    /var/log/nginx/error.log")
        time.Sleep(time.Second * 5)
        return nil
}


//启动关闭所有进程
func stop_server(){
        kill_shell := `#!/bin/sh
set -x
ps -ef | /bin/grep -v grep | /bin/grep -E "/opt/jdk1.8.0_131/bin/java|boddy|/home/activemq/apache-activemq-5.8.0|/home/WordFlow/tomcat-WordFlow|/home/CloudAPP/tomcat-web|/home/CloudAPPauto|nginx:" 
ps_list=$(ps -ef | /bin/grep -v grep | /bin/grep -E "/opt/jdk1.8.0_131/bin/java|boddy|/home/activemq/apache-activemq-5.8.0|/home/WordFlow/tomcat-WordFlow|/home/CloudAPP/tomcat-web|/home/CloudAPPauto|nginx:" | /bin/awk '{print $2}' | /usr/bin/xargs)
if [ "x${ps_list}" == "x" ];then
        exit 0
fi
echo "java ps id list is  :  ${ps_list}"
/bin/kill -9 ${ps_list}
`
        kill_script_err := ioutil.WriteFile("/tmp/kill_all_CloudAPP_process.sh", []byte(kill_shell), 0755)
        if kill_script_err != nil {
                fmt.Println("/tmp/kill_all_CloudAPP_process.sh write\t\t\t\t\t\t\t\t----> fail,please check. ",kill_script_err)
        } else {
                fmt.Println("/tmp/kill_all_CloudAPP_process.sh write\t\t\t\t\t\t\t\t----> Success")
        }
        //执行/tmp/kill_script_run.sh配置脚本
        kill_script_run := []string{"/tmp/kill_all_CloudAPP_process.sh"}
    if execCommand("/bin/bash",kill_script_run) {
                fmt.Println("/bin/bash /tmp/kill_all_CloudAPP_process.sh ---> Success")
        } else {
                fmt.Println("/bin/bash /tmp/kill_all_CloudAPP_process.sh ---> fail,please check")
        }
}


func FileExists(path string) bool {
        _, err := os.Stat(path)    //os.Stat获取文件信息
        if err != nil {
                if os.IsExist(err) {
                        return true
                }
                return false
        }
        return true
}



//修改文件内容
func tracefile(str_content,fullpathfilename,regstring string) bool {
       filebuff, err := ioutil.ReadFile(fullpathfilename)
       if err != nil {
          fmt.Println("ERROR : file " + fullpathfilename + " not exist!!!")
          return false
       }
       reg := regexp.MustCompile(regstring)
       rep := []byte(str_content)
       buff:= reg.ReplaceAllLiteral(filebuff, rep)
       //fmt.Println(string(buff))
       ioutil.WriteFile(fullpathfilename, buff, 0644)
       return true
}


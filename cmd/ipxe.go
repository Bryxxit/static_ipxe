package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"text/template"
)

type Node struct {
	Name     string
	Network  []Network `json:"network"`
	Hostname string
}

type Network struct {
	Type          string     `json:"type"`
	Name          string     `json:"name"`
	Netmask       string     `json:"netmask"`
	Ip            string     `json:"ip"`
	Gateway       string     `json:"gateway"`
	BootInterface bool       `json:"bootInterface"`
	Vlans         *[]Network `json:"vlans"`
}

type BootstrapPage struct {
	Host string
	ISO  *ISO
	Node *Node
}

type ISO struct {
	InitRD        string `json:"initrd"`
	VMLinuz       string `json:"vmlinuz"`
	//KernelArgs    string `json:"kernelArgs"`
	//ImageUrl      string `json:"imageUrl"`
	KickstartUrl   string `json:"kickstart_url"`
	//KickstartIp   string `json:"kickstart_ip"`
	//KickstartPort int    `json:"kickstart_port"`
}

//func main() {
//	iP := "/home/zzeac/Desktop/Go2/go-projects/ipxe/ipxe"
//	node := Node{
//		Name:     "testy",
//		Network:  []Network{},
//		Hostname: "testy",
//	}
//
//	nw := Network{
//		Type:          "static",
//		Name:          "qdqsd",
//		Ip:            "10.65.194.112",
//		Gateway:       "10.65.194.1",
//		Netmask:       "255.255.255.0",
//		BootInterface: true,
//	}
//
//	node.Network = append(node.Network, nw)
//	// needded kickstarturl, vmlinuz url, initrd url
//	i := ISO{
//
//	}
//}

func CompileIpxeIso(ipxePath string, compilePath string, host string, j Node, iso ISO,  dest string) error {

	// create temp dir
	tempDir, err := CreateTempFolder(compilePath)

	// if you want to keep bootstrap files
	err = os.MkdirAll(compilePath+"/backupScripts", os.ModePerm)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	// copy ipxe contents
	err = CopyIpxeDir(ipxePath+"/", compilePath+"/"+tempDir)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	// write the template file
	d1 := []byte(ipxe_template_content)
	err = ioutil.WriteFile(ipxePath+"/bootstrap.ipxe", d1, 0644)
	if err != nil {
		log.Println(err.Error())
		return err
	}


	// Create the template
	err = WriteTemplate(host, j, ipxePath+"/bootstrap.ipxe", compilePath+"/"+tempDir+"/ipxe/src/bootstrap.ipxe", iso)
	if err != nil {
		return err
	}
	err = WriteTemplate(host, j, ipxePath+"/bootstrap.ipxe", compilePath+"/backupScripts/"+j.Name, iso)
	if err != nil {
		return err
	}

	// compile
	err = compileCommandLocal(compilePath + "/" + tempDir + "/ipxe/src")
	if err != nil {
		log.Println(err.Error())
		return err

	}
	////upload to the correct datastore
	//err = UploadIso(j, compilePath+"/"+tempDir+"/ipxe/src/bin/ipxe.iso", vh)
	//if err != nil {
	//	log.Println(err.Error())
	//	return err
	//}

	oldLocation := compilePath+"/"+tempDir+"/ipxe/src/bin/ipxe.iso"
	newLocation := dest
	err = os.Rename(oldLocation, newLocation)
	if err != nil {
		log.Fatal(err)
	}

	//clear out and clean out
	err = removeCompileFolderCommandLocal(compilePath + "/" + tempDir)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return nil

}

func WriteTemplate(host string, node Node, src string, dst string, iso ISO) error {
	//t, err := template.ParseFiles("/home/tim/Desktop/Go/src/zira/bootstrap.ipxe")
	funcs := template.FuncMap{"add": add, "getNext": getNext, "bootString": bootString}
	t, err := template.New("bootstrap.ipxe").Funcs(funcs).ParseFiles(src)

	if err != nil {
		log.Println(err.Error())
		return err

	}
	//t.Funcs(funcs)
	f, err := os.Create(dst)
	if err != nil {
		log.Println(err.Error())
		return err

	}

	// Execute the template to the file.
	nr := BootstrapPage{
		host,
		&iso,
		&node,
	}
	err = t.Execute(f, nr)
	if err != nil {
		log.Println(err.Error())
		return err

	}

	// Close the file when done.
	f.Close()
	return nil
}

func add(x, y int) int {
	return x + y
}

func getBootInterface(node *Node, iso *ISO, host string) (string, int) {
	ipString := ""
	index := 0
	for i, nw := range node.Network {
		if nw.BootInterface == true {
			index = i
			if nw.Type == "static" {
				ipString = fmt.Sprintf("ip=%s::%s:%s:%s:eth%d:off", nw.Ip, nw.Gateway, nw.Netmask, node.Hostname, i)
			}
			if nw.Type == "dhcp" {
				ipString = "dhcp"
				index = i
			}
			break
		}
	}
	if ipString == "" {
		index = 0
		for i, nw := range node.Network {
			if nw.Type != "disabled" {
				index = i
				if nw.Type == "static" {
					ipString = fmt.Sprintf("%s::%s:%s:%s:eth%d:off", nw.Ip, nw.Gateway, nw.Netmask, node.Hostname, i)
				}
				if nw.Type == "dhcp" {
					ipString = "dhcp"
					index = i
				}
				break
			}
		}
	}

	return ipString, index
}

func bootString(node *Node, iso *ISO, host string) string {

	ipString, index := getBootInterface(node, iso, host)
	str := "sleep 5\n"
	//if iso.KickstartIp == "" {
	//	iso.KickstartIp = strings.Split(host, ":")[0]
	//}
	//if iso.KickstartPort == 0 {
	//	var err error
	//	iso.KickstartPort, err = strconv.Atoi(strings.Split(host, ":")[1])
	//	if err != nil {
	//		iso.KickstartPort = 8150
	//	}
	//}

	str += fmt.Sprintf("kernel %s biosdevname=0 net.ifnames=0 ksdevice=eth%d ks=%s BOOTIF=01-${netX/mac} %s", iso.VMLinuz, index, iso.KickstartUrl, ipString)
	str += "\ninitrd " + iso.InitRD + " || goto error\nboot\n"
	return str

}

func getNext(index int, length int) string {
	if index == length-1 {
		return "goto chain_boot"

	}
	return fmt.Sprintf("goto set_net%d", index+1)
}

func CopyIpxeDir(src, dst string) error {
	return copyCommandLocal(src, dst)
}

func copyCommandLocal(src, dst string) error {
	_, err := exec.Command("cp", "-rf", src, dst).CombinedOutput()
	if err != nil {
		log.Println(err)
	}
	return nil
}

func compileCommandLocal(dst string) error {

	cmd := exec.Command("make", "EMBED=bootstrap.ipxe", "bin/ipxe.iso")
	cmd.Dir = dst
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(err)
	}
	return nil
	//cmd := exec.Command("make", "EMBED=bootstrap.ipxe", "bin/ipxe.iso")
	//cmd.Dir = dst
	//stdout, err := cmd.StdoutPipe()
	//if err != nil {
	//	log.Println(err)
	//}
	//if err := cmd.Start(); err != nil {
	//	log.Println(err)
	//}
	//log.Println(stdout)
	//if err := cmd.Wait(); err != nil {
	//	log.Println(err)
	//}
	//return nil
}

func removeCompileFolderCommandLocal(dst string) error {

	cmd := exec.Command("rm", "-rf", dst)
	_, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(err)
	}
	return nil
}

func CreateTempFolder(dest string) (string, error) {

	return createTempFolderLocal(dest)

}
func createTempFolderLocal(dest string) (string, error) {

	randy := RandStringRunes(5)
	err := os.MkdirAll(dest+"/"+randy, os.ModePerm)
	if err != nil {
		log.Println(err)
		return randy, err
	}
	return randy, nil

}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	//rand.Seed(time.Now().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}


var ipxe_template_content = `
#!ipxe
set tries:uint32    0
set maxtries:uint32 10

:retry
isset ${ip} || goto set_net0
set dhcp_mac ${mac:hexhyp}
goto chain_boot


{{ if .Node -}}
 {{- if .Node.Network -}}
    {{- $length := len .Node.Network -}}
    {{/* Loop trough each interface */}}
    {{- $counter:=0 -}}
    {{- range .Node.Network -}}
        {{- if (eq .Type "static") -}}
:set_net{{$counter}}
isset ${net{{$counter}}/mac} && set net{{$counter}}/ip {{.Ip}} && set net{{$counter}}/netmask {{.Netmask}} && set net{{$counter}}/gateway {{.Gateway}} && ifopen net{{$counter}} || {{getNext $counter $length }}
echo net{{$counter}} has static ip
set dhcp_mac net{{$counter}}$${net{{$counter}}/mac:hexhyp}

        {{- else if (eq .Type "disabled") -}}
:set_net{{$counter}}
isset ${net{{$counter}}/mac} && ifclose net{{$counter}} || {{getNext $counter $length }}
echo net{{$counter}}  has been disabled
set dhcp_mac net{{$counter}}$${net{{$counter}}/mac:hexhyp}

        {{- else if (eq .Type "dhcp") -}}
:set_net{{$counter}}
isset ${net{{$counter}}/mac} && dhcp net{{$counter}} || {{getNext $counter $length }}
echo net{{$counter}} has DHCP
set dhcp_mac net{{$counter}}$${net{{$counter}}/mac:hexhyp}
            {{/* TODO implement vlan support */}}
        {{- end }}

{{$counter = (add $counter 1) -}}

    {{- end -}}
{{ else -}}
:set_net0
isset ${net0/mac} && dhcp net0 || goto set_net1
echo net0 has DHCP
set dhcp_mac net0$${net0/mac:hexhyp}

:set_net1
isset ${net1/mac} && dhcp net0 || goto set_net2
echo net1 has DHCP
set dhcp_mac net1$${net1/mac:hexhyp}

:set_net2
isset ${net2/mac} && dhcp net0 || goto set_net3
echo net2 has DHCP
set dhcp_mac net2$${net2/mac:hexhyp}

:set_net3
isset ${net3/mac} && dhcp net3 || goto set_net4
echo net3 has DHCP
set dhcp_mac net3$${net3/mac:hexhyp}

:set_net4
isset ${net4/mac} && dhcp net4 || goto chain_boot
echo net4 has DHCP
set dhcp_mac net4$${net4/mac:hexhyp}
 {{ end -}}
{{ end -}}


:chain_boot
ifstat
nstat
route
{{ bootString .Node .ISO .Host }}
exit

:error
iseq ${tries} ${maxtries} && goto failed
inc tries
sleep ${tries}
goto retry

:failed
echo failed to obtain DHCP data after ${tries} attempts, giving up.
sleep 60
reboot
`
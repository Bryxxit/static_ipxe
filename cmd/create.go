/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"gopkg.in/src-d/go-git.v4"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Creates an ipxe iso",
	Long: `Create an ipxe iso with a static ip embedded so you can deploy nodes without having to configure a dhcp`,
	Run: func(cmd *cobra.Command, args []string) {
		ipxe_url, _ := cmd.Flags().GetString("ipxeurl")
		dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
		name, _ := cmd.Flags().GetString("name")
		compile_path := dir + "/compile"
 		if err != nil {
			log.Fatal(err)
		}
		ipxe_dir := dir + "/ipxe"
		log.Println("Cloning or pulling latest ipxe git")
		err = cloneRepo(ipxe_url, ipxe_dir)
		if err != nil {
			log.Fatal(err)
		}
		ip, _ := cmd.Flags().GetString("ip")
		gateway, _ := cmd.Flags().GetString("gateway")
		netmask, _ := cmd.Flags().GetString("netmask")
		initrd, _ := cmd.Flags().GetString("initrd")
		vmlinuz, _ := cmd.Flags().GetString("vmlinuz")
		kickstart, _ := cmd.Flags().GetString("kickstarturl")
		// set params for compiling
		node := Node{
			Name:     name,
			Network:  []Network{},
			Hostname: name,
		}
		nw := Network{
			Type:          "static",
			Name:          name, // could set this name to interface name
			Ip:            ip,
			Gateway:       gateway,
			Netmask:       netmask,
			BootInterface: true,
		}
		node.Network = append(node.Network, nw)
		i := ISO{
			InitRD: initrd,
			VMLinuz:vmlinuz,
			KickstartUrl: kickstart,
		}

		println("Compiling iso this takes a while your iso should be available at: " +dir + "/" + name + ".iso" )
		err = CompileIpxeIso(ipxe_dir, compile_path, name, node, i, dir + "/" + name + ".iso")
		if err != nil {
			log.Println(err)
		}
	},
}

func IsDirectoryAbsentOrIsemptyDirectory(path string) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return true, err
	}
	test, err := IsEmpty(path)
	return test, err
}

func IsEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}


func cloneRepo(url string, dest string) error {
	// Check if dir is exmpy or exists
	test, _ := IsDirectoryAbsentOrIsemptyDirectory(dest)

	if test {
		_, err := git.PlainClone(dest, false, &git.CloneOptions{
			URL:      url,
			Progress: os.Stdout,
		})
		if err != nil {
			return err
		}

	} else {
		com := "cd " + dest + " && git pull "
		println(com)
		cmd := exec.Command("sh", "-c", com+"; echo 1>&2 stderr")
		stdoutStderr, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%s\n", stdoutStderr)

		return nil
	}

	return nil
}


func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().StringP("ipxeurl", "x", "https://github.com/ipxe/ipxe.git", "ipxe url")

	createCmd.Flags().StringP("name", "n", "", "Set hostname for the machine (required)")
	createCmd.Flags().StringP("initrd", "I", "", "Set the initrd url (required)")
	createCmd.Flags().StringP("vmlinuz", "V", "", "Set vmlinuz url (required)")
	createCmd.Flags().StringP("kickstarturl", "u", "", "Set kickstartulr (required)")
	createCmd.Flags().StringP("ip", "i", "", "Set ip (required)")
	createCmd.Flags().StringP("netmask", "m", "", "Set netmask (required)")
	createCmd.Flags().StringP("gateway", "g", "", "Set gateway (required)")
	createCmd.MarkFlagRequired("name")
	createCmd.MarkFlagRequired("initrd")
	createCmd.MarkFlagRequired("vmlinuz")
	createCmd.MarkFlagRequired("kickstarturl")
	createCmd.MarkFlagRequired("ip")
	createCmd.MarkFlagRequired("netmask")
	createCmd.MarkFlagRequired("gateway")


	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}



# static_ipxe
This is a simple script to compile ixpe isos with a static ip embedded.


## Installation

go install  github.com/Bryxxit/static_ipxe


To compile the iso you'll need to install some depedencies according to the official ipxe site
- gcc (version 3 or later)
- binutils (version 2.18 or later)
- make
- perl
- liblzma or xz header files
- mtools
- mkisofs (needed only for building .iso images)
- syslinux (for isolinux, needed only for building .iso images)

  yum install -y isolinux mkisofs mtools xz-devel perl make binutils gcc syslinux nfs-utils genisoimage

## Ussage
The package should now be available under your gopath
cd $GOPATH/bin
./static_ipxe create -n testy --initrd "/path/to/initd" --vmlinuz "/path/to/vmlinuz" --kickstarturl "url.org/kickstart" --netmask 255.255.255.0 --ip 10.10.10.10 --gateway 10.10.10.1

Your result should be a:
  - A folder containing ipxe git
  - A compile folder containing a backupscript folder with your bootstrap ipxe file
  - and finally your iso being hostname.iso
  


package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"log"
	"time"

	"github.com/pedroalbanese/tftp"
)

var (
	ftp    = flag.String("ftp", "client", "Server or Client.")
	iport  = flag.String("ipport", "", "Local Port/remote's side Public IP:Port.")
	path   = flag.String("file", "", "File to Download/Upload.")
	upload = flag.Bool("upload", false, "Upload file to server.")
	noup   = flag.Bool("noup", false, "Does not allow Upload files to server.")
)

func main() {
	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "TFTP (c) 2022 ALBANESE Research Lab")
		fmt.Fprintln(os.Stderr, "Trivial File Transfer Protocol Tool")
	desc := `TFTP Implements:
   RFC 1350 - The TFTP Protocol (Revision 2)
   RFC 2347 - TFTP Option Extension
   RFC 2348 - TFTP Blocksize Option
   RFC 2349 - TFTP Timeout Interval and Transfer Size Options`
		fmt.Fprintln(os.Stderr, "\n"+desc)
		fmt.Fprintln(os.Stderr, "\nUsage of", os.Args[0]+":")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *ftp == "server" {
		ipport := "69"
		if *iport != "" {
			ipport = *iport
		}
		// use nil in place of handler to disable read or write operations
		var s *tftp.Server
		if *noup {
			s = tftp.NewServer(readHandler, nil)
		} else {
			s = tftp.NewServer(readHandler, writeHandler)
		}
		s.SetTimeout(5 * time.Second)
		err := s.ListenAndServe(":"+ipport)
		if err != nil {
			fmt.Fprintf(os.Stdout, "server: %v\n", err)
			os.Exit(1)
		}
	} else if *ftp == "client" && *path != "" {
		ipport := "127.0.0.1:69"
		if *iport != "" {
			ipport = *iport
		}
		if *upload {
			c, err := tftp.NewClient(ipport)
			if err != nil {
				fmt.Fprintf(os.Stdout, "client: %v\n", err)
				os.Exit(1)
			}
			file, err := os.Open(*path)
			if err != nil {
				fmt.Fprintf(os.Stdout, "client: %v\n", err)
				os.Exit(1)
			}
			c.SetTimeout(5 * time.Second)
			rf, err := c.Send(*path, "octet")
			if err != nil {
				fmt.Fprintf(os.Stdout, "client: %v\n", err)
				os.Exit(1)
			}
			if err != nil {
				fmt.Fprintf(os.Stdout, "client: %v\n", err)
				os.Exit(1)
			}
			rf.(tftp.OutgoingTransfer).SetSize(15360)
			n, err := rf.ReadFrom(file)
			if err != nil {
				fmt.Fprintf(os.Stdout, "client: %v\n", err)
				os.Exit(1)
			}
		        raddr := rf.(tftp.OutgoingTransfer).RemoteAddr()
		        log.Println("RRQ from", raddr.String())
			fmt.Printf("%d bytes sent\n", n)
		} else {
			c, err := tftp.NewClient(ipport)
			if err != nil {
				fmt.Fprintf(os.Stdout, "client: %v\n", err)
				os.Exit(1)
			}
			wt, err := c.Receive(*path, "octet")
			if err != nil {
				fmt.Fprintf(os.Stdout, "client: %v\n", err)
				os.Exit(1)
			}
			file, err := os.Create(*path)
			if err != nil {
				fmt.Fprintf(os.Stdout, "client: %v\n", err)
				os.Exit(1)
			}
			// Optionally obtain transfer size before actual data.
			if n, ok := wt.(tftp.IncomingTransfer).Size(); ok {
				fmt.Printf("Transfer size: %d\n", n)
			}
			n, err := wt.WriteTo(file)
			if err != nil {
				fmt.Fprintf(os.Stdout, "client: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("%d bytes received\n", n)
		}
	}
}

func readHandler(filename string, rf io.ReaderFrom) error {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	rf.(tftp.OutgoingTransfer).SetSize(15360)
	n, err := rf.ReadFrom(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	file.Close()
        raddr := rf.(tftp.OutgoingTransfer).RemoteAddr()
        log.Println("RRQ from", raddr.String())
	fmt.Printf("%d bytes sent\n", n)
	return nil
}

// writeHandler is called when client starts file upload to server
func writeHandler(filename string, wt io.WriterTo) error {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	n, err := wt.WriteTo(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	file.Close()
	fmt.Printf("%d bytes received\n", n)
	return nil
}

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/Comcast/gots/packet"
	"github.com/Comcast/gots/pes"
	"github.com/Comcast/gots/psi"
)

func main() {
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Printf("dump [flags] [input file path]\n")
		flag.Usage()
		return
	}
	filePath := os.Args[1]

	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	reader := bufio.NewReader(file)
	_, err = packet.Sync(reader)
	if err != nil {
		panic(err)
	}

	pat, err := psi.ReadPAT(reader)
	if err != nil {
		panic(err)
	}
	printPAT(pat)

	var pmts []psi.PMT
	pm := pat.ProgramMap()
	for pn, pid := range pm {
		pmt, err := psi.ReadPMT(reader, pid)
		if err != nil {
			panic(err)
		}
		pmts = append(pmts, pmt)
		printPMT(pn, pmt)
	}

	var pkt packet.Packet
	var numPackets uint64
	fmt.Println("\nPacket  PID  PTS         DTS")
	for {
		if _, err := io.ReadFull(reader, pkt[:]); err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		numPackets++

		b, err := packet.PESHeader(&pkt)
		if err != nil {
			continue
		}
		ph, err := pes.NewPESHeader(b)
		if err != nil {
			panic(err)
		}
		pid := packet.Pid(&pkt)
		fmt.Printf("%6v  %3v  %10v  %10v\n", numPackets, pid, ph.PTS(), ph.DTS())
	}
}

func printPAT(pat psi.PAT) {
	fmt.Println("PAT")
	fmt.Printf("\tPMT PIDs %v\n", pat.ProgramMap())
	fmt.Printf("\tNumber of Programs %v\n", pat.NumPrograms())
}

func printPMT(pn uint16, pmt psi.PMT) {
	fmt.Printf("Program #%v PMT\n", pn)
	fmt.Printf("\tPIDs %v\n", pmt.Pids())
	fmt.Println("\tElementary Streams")

	for _, es := range pmt.ElementaryStreams() {
		fmt.Printf("\t\tPid %v: StreamType %v: %v\n", es.ElementaryPid(), es.StreamType(), es.StreamTypeDescription())

		for _, d := range es.Descriptors() {
			fmt.Printf("\t\t\t%+v\n", d)
		}
	}
}

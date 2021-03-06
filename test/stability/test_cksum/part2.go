// Copyright 2017 Intel Corporation.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"

	"github.com/intel-go/yanff/flow"
	"github.com/intel-go/yanff/packet"

	"github.com/intel-go/yanff/test/stability/test_cksum/testCommon"
)

var hwol bool

// Main function for constructing packet processing graph.
func main() {
	var inport, outport uint

	flag.BoolVar(&hwol, "hwol", false, "Use Hardware offloading for TX checksums calculation")
	flag.UintVar(&inport, "inport", 0, "Input port number")
	flag.UintVar(&outport, "outport", 1, "Output port number")
	flag.Parse()

	// Init YANFF system
	config := flow.Config{
		CPUCoresNumber: 16,
		HWTXChecksum:   hwol,
	}
	flow.SystemInit(&config)

	// Receive packets from zero port. Receive queue will be added automatically.
	inputFlow := flow.SetReceiver(uint8(inport))
	flow.SetHandler(inputFlow, fixPacket, nil)
	flow.SetSender(inputFlow, uint8(outport))

	// Begin to process packets.
	flow.SystemStart()
}

func fixPacket(pkt *packet.Packet, context flow.UserContext) {
	offset := pkt.ParseData()

	if !testCommon.CheckPacketChecksums(pkt) {
		println("TEST FAILED")
	}

	if offset < 0 {
		println("ParseL4 returned negative value", offset)
		println("TEST FAILED")
		return
	}

	ptr := (*testCommon.Packetdata)(pkt.Data)
	if ptr.F2 != 0 {
		fmt.Printf("Bad data found in the packet: %x\n", ptr.F2)
		println("TEST FAILED")
		return
	}

	ptr.F2 = ptr.F1

	if hwol {
		packet.SetPseudoHdrChecksum(pkt)
		packet.SetHWCksumOLFlags(pkt)
	} else {
		testCommon.CalculateChecksum(pkt)
	}
}

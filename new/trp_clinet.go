// Copyright 2012 The GoSNMP Authors. All rights reserved.  Use of this
// source code is governed by a BSD-style license that can be found in the
// LICENSE file.

/*
The developer of the trapserver code (https://github.com/jda) says "I'm working
on the best level of abstraction but I'm able to receive traps from a Cisco
switch and Net-SNMP".
Pull requests welcome.
*/

package main

import (
	// "flag"
	"fmt"
	"log"
	"net"

	// "os"

	// "path/filepath"

	g "github.com/gosnmp/gosnmp"
	// "encoding/hex"
)

func main() {
	// flag.Usage = func() {
	// 	fmt.Printf("Usage:\n")
	// 	fmt.Printf("   %s\n", filepath.Base(os.Args[0]))
	// 	flag.PrintDefaults()
	// }

	tl := g.NewTrapListener()
	tl.OnNewTrap = myTrapHandler
	tl.Params = g.Default
	// tl.Params.Logger = g.NewLogger(log.New(os.Stdout, "", 0))

	err := tl.Listen("127.0.0.1:163")
	if err != nil {
		log.Panicf("error in listen: %s", err)
	}
}

func myTrapHandler(packet *g.SnmpPacket, addr *net.UDPAddr) {
	log.Printf("got trapdata from %s\n", addr.IP)
	for _, v := range packet.Variables {
		switch v.Type {
		case g.OctetString:
			b := v.Value.([]byte)
			// decoded, err := hex.DecodeString(string(b))
			// if err != nil {
			// 	log.Fatal(err)
			// }
			fmt.Printf("OID: %s, string: %s\n", v.Name, b)

			sendTrap(v.Name, v.Type, string(v.Value.([]byte)))

		default:
			// log.Printf("trap: %+v\n", v)
			// sendTrap(v.Name, v.Type, string(v.Value.([]byte)))
		}
	}
}

func sendTrap(a string, b g.Asn1BER, c string) {
	// Default is a pointer to a GoSNMP struct that contains sensible defaults
	// eg port 164, community public, etc
	g.Default.Target = "127.0.0.1"
	g.Default.Port = 164
	g.Default.Version = g.Version2c
	g.Default.Community = "public"
	// g.Default.Logger = g.NewLogger(log.New(os.Stdout, "", 0))
	err := g.Default.Connect()
	if err != nil {
		log.Fatalf("Connect() err: %v", err)
	}
	defer g.Default.Conn.Close()

	pdu := g.SnmpPDU{
		Name:  a,
		Type:  b,
		Value: c,
	}

	trap := g.SnmpTrap{
		Variables: []g.SnmpPDU{pdu},
	}

	_, err = g.Default.SendTrap(trap)
	if err != nil {
		log.Fatalf("SendTrap() err: %v", err)
	} else {
		log.Print("Trap sent to SNMP server")
	}
}

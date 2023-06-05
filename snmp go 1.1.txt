package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gosnmp/gosnmp"

	"os"
)

// Struct to hold SNMP server details
type SNMPConfig struct {
	IP   string
	Port uint16
}

// Struct to hold SNMP trap information
type SNMPTrap struct {
	Oid       string `json:"oid"`
	Value     string `json:"value"`
	StartTime int64  `json:"start_time"`
}

type Params struct {
	info, value string
}

// Function to print log in the terminal
func doLog(p Params) {
	fmt.Println(p.info, p.value)
}

// Function to print error in the terminal
func doError(info string, err string) {
	log.Fatalf(info, err)
}

// Make connection to the database
var db, err = sql.Open("mysql", "user1:1234@tcp(localhost:3306)/snmp")

// Function to open UDP socket and listen for SNMP traps
func listenForSNMPTraps(addr string) {
	// Open UDP socket on the specified address
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		doError("Error creating UDP socket: %s\n", err.Error())
	}
	defer conn.Close()

	doLog(Params{info: "UDP server listening on", value: addr})

	// Buffer to store incoming packet data
	buffer := make([]byte, 1024)

	// Receive and process incoming packets
	for {
		// Read from the connection into the buffer
		numberOfBytes, _, err := conn.ReadFrom(buffer)
		if err != nil {
			doError("Error reading from UDP socket: %s\n", err.Error())
			continue
		}

		go func(packet []byte) {
			// Extract SNMP trap information from the received packet
			trap, err := extractSNMPTrapInfo(packet)
			if err != nil {
				doError("Error extracting SNMP trap information: %s\n", err.Error())
				return
			}

			// Retrieve SNMP server details from the database
			snmpConfig, err := getSNMPConfigFromDB()
			if err != nil {
				doError("Error retrieving SNMP server details: %s\n", err.Error())
				return
			}

			// Create an SNMP trap message
			trap.StartTime = time.Now().Unix()

			// Convert SNMP trap to JSON
			trapJSON, err := json.Marshal(trap)
			if err != nil {
				doError("Error converting SNMP trap to JSON: %s\n", err.Error())
				return
			}

			// Send the SNMP trap message to the SNMP server
			err = sendSNMPTrap(trapJSON, snmpConfig.IP, snmpConfig.Port)
			if err != nil {
				doError("Error sending SNMP trap: %s\n", err.Error())
				return
			}

			doLog(Params{info: "\n\nPacket received successfully!"})

			date := time.Now()
			formatedDate := date.Format("01-02-2006 15:04:05")

			// Insert packet's information into log table in the SNMP database
			insertQuery, err := db.Query("INSERT INTO log (action, data, date) VALUES ('New Packet Received',?,?)", string(packet), formatedDate)

			if err != nil {
				doError("Error while inserting data: %s", err.Error())
			}
			defer insertQuery.Close()

			doLog(Params{info: "Packet inserted into database successfully!"})

			// Writing packet's information with defined template into the log file
			logFile, err := os.OpenFile("logSample.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				doError("Error while proccessing log file: %s", err.Error())
			}

			logTemplate := formatedDate + " || Notice: New Packet Received! => " + string(packet)

			bytesNumber, err := logFile.WriteString(logTemplate)
			if err != nil {
				doError("Error while writing into the log file: %s", err.Error())
			}
			_ = bytesNumber
			doLog(Params{info: "Packet inserted into log file successfully!"})
			logFile.Sync()

		}(buffer[:numberOfBytes])
	}
}

// Function to extract SNMP trap information from the packet
func extractSNMPTrapInfo(packet []byte) (*SNMPTrap, error) {
	// Add your logic to extract the OID and value from the packet
	// and create an SNMPTrap struct with the extracted values
	oid := "1.3.6.1.2.1.1.3.0"
	value := "uptime"
	trap := &SNMPTrap{
		Oid:   oid,
		Value: value,
	}
	return trap, nil
}

// Function to retrieve SNMP server details from the database
func getSNMPConfigFromDB() (SNMPConfig, error) {
	// Retrieve SNMP server details from the genopts table
	var snmpConfig SNMPConfig
	err = db.QueryRow("SELECT snmp_server_ip, snmp_trap_port FROM genopts LIMIT 1").Scan(&snmpConfig.IP, &snmpConfig.Port)
	if err != nil {
		return SNMPConfig{}, fmt.Errorf("error retrieving SNMP server details: %s", err.Error())
	}

	return snmpConfig, nil
}

// Function to send SNMP trap to the SNMP server
func sendSNMPTrap(trapJSON []byte, ip string, port uint16) error {
	// Create a new SNMP client
	client := &gosnmp.GoSNMP{
		Target:    ip,
		Port:      port,
		Community: "public",
		Version:   gosnmp.Version2c,
		Timeout:   time.Duration(5) * time.Second,
	}

	// Connect to the SNMP server
	err := client.Connect()
	if err != nil {
		return err
	}
	defer client.Conn.Close()

	// Send the SNMP trap
	//  err = client.SendTrapBytes(trapJSON)
	//  if err != nil {
	//  return err
	//  }

	return nil
}

func main() {
	listenForSNMPTraps("127.0.0.1:222")
}

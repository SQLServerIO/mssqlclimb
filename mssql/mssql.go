package mssql

import (
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/jmoiron/sqlx"

	//needed for sqlx
	_ "github.com/sqlserverio/go-mssqldb"

	"github.com/urfave/cli"
)

//Connect sets up a database connection and create the import schema
func Connect(connStr string) (*sqlx.DB, error) {
	db, err := sqlx.Open("mssql", connStr)
	if err != nil {
		return db, err
	}

	err = db.Ping()
	if err != nil {
		return db, err
	}

	return db, nil
}

//ParseConnStr parses sql connection string from cli flags
func ParseConnStr(c *cli.Context) (connectionString string) {

	//split host from instance name
	serverInstance := strings.Split(c.GlobalString("host"), "\\")
	Server := serverInstance[0]

	//pull our host name for comparison
	hn, err := os.Hostname()
	if err != nil {
		fmt.Printf("Unable to get computers host name %s", err.Error())
		os.Exit(0)
	}
	//look to see if we are trying to connect to the localhost and if so use a valid ip address instead
	//this is here due to a bug in go-mssqldb where you can't connect via windows auth if you are local to the server
	if strings.Contains(strings.ToUpper(serverInstance[0]), strings.ToUpper(hn)) {
		iaddrs, err := net.InterfaceAddrs()
		if err != nil {
			return
		}
		for _, address := range iaddrs {
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				//hacky work around trying to connect locally and getting IPv6 local host address instead of IPv4
				if ipnet.IP.To4() != nil {
					if len(serverInstance) > 1 {
						Server = ipnet.IP.String() + "\\" + serverInstance[1]
					} else {
						Server = ipnet.IP.String()
					}
					return "server=" + Server + ";database=" + c.GlobalString("dbname") + ";user id=" + c.GlobalString("username") + ";password=" + c.GlobalString("pass") + ";connection timeout=3600;encrypt=disable"
				}
			}
		}
	}
	return "server=" + Server + ";database=" + c.GlobalString("dbname") + ";user id=" + c.GlobalString("username") + ";password=" + c.GlobalString("pass") + ";connection timeout=3600;encrypt=disable"
}

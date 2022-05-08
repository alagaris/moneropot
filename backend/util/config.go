package util

import (
	"fmt"
	"log"

	"github.com/namsral/flag"

	"gopkg.in/natefinch/lumberjack.v2"
)

type config struct {
	Bind          string
	MaintAddress  string
	FundAddress   string
	RpcUser       string
	RpcPass       string
	RpcAddress    string
	DaemonUser    string
	DaemonPass    string
	DaemonAddress string
	DataPath      string
	DbName        string
	Production    bool
	SMTPHost      string
	SMTPPort      string
	SMTPUser      string
	SMTPPass      string
	ContactEmail  string
	LogFile       string
	AdminKey      string
}

var (
	Config = config{}
)

func ParseArgs() {
	flag.StringVar(&Config.Bind, "bind", "localhost:5000", "address:port to bind server")
	flag.StringVar(&Config.MaintAddress, "maint-address", "9xgcCBjmPLvK49CRfJQk46DbHJGSErHJBAJ9dT9nV3FxGVgo5oyRpHiRsJEMq6a1UVfXTpQEhfj3nYJH7gxo15b9Q4u6NjW", "monero address to send 10% after drawing")
	flag.StringVar(&Config.FundAddress, "fund-address", "A2eCGjvYkowZbMtshUi7ki7QrvDPtiepgfsy9TQLSHXLBfgihU6z8ZBYFP83yZ86MxdMxTJyqUFARFHAgtaP9eFtJNC69Jk", "fund address to send 5% after drawing")
	flag.StringVar(&Config.RpcAddress, "rpc-address", "http://localhost:18082/json_rpc", "monero wallet rpc address")
	flag.StringVar(&Config.RpcUser, "rpc-user", "", "monero wallet rpc username")
	flag.StringVar(&Config.RpcPass, "rpc-pass", "", "monero wallet rpc password")
	flag.StringVar(&Config.DaemonAddress, "daemon-address", "http://localhost:28081/json_rpc", "monero daemon rpc address")
	flag.StringVar(&Config.DaemonUser, "daemon-user", "", "monero daemon rpc username")
	flag.StringVar(&Config.DaemonPass, "daemon-pass", "", "monero daemon rpc password")
	flag.StringVar(&Config.DataPath, "data-path", "./data", "db storage")
	flag.StringVar(&Config.DbName, "db-name", "data.db", "db filename")
	flag.StringVar(&Config.SMTPHost, "smtp-host", "smtp.privatemail.com", "SMTP host")
	flag.StringVar(&Config.SMTPPort, "smtp-port", "587", "SMTP port")
	flag.StringVar(&Config.SMTPUser, "smtp-user", "", "SMTP username")
	flag.StringVar(&Config.SMTPPass, "smtp-pass", "", "SMTP password")
	flag.StringVar(&Config.LogFile, "log-file", "", "Log file")
	flag.StringVar(&Config.AdminKey, "admin-key", "abc123", "Admin key for auth stuff")
	flag.StringVar(&Config.ContactEmail, "contact-email", "support@moneropot.org", "Contact email")
	flag.BoolVar(&Config.Production, "production", false, "running in production")
	flag.Parse()
	if Config.MaintAddress == "" {
		log.Fatal(fmt.Errorf("no maintenance address provided"))
	} else if len(Config.MaintAddress) != 95 {
		log.Fatal(fmt.Errorf("invalid maintenance address provided"))
	}

	if Config.LogFile != "" {
		log.SetOutput(&lumberjack.Logger{
			Filename:   Config.LogFile,
			MaxSize:    50,
			MaxBackups: 2,
			MaxAge:     30,
			Compress:   true,
		})
	}
}

package web

import (
	"log"
	"os"
)

// The only logging that the web package does is to ERROR in error conditions. Error conditions are panic conditions (eg, bad route configuration), as well as
// application panics (eg, dbz in an app handler).
// Applications can set web.ERROR = your own logger, if they wish.

// In terms of logging the requests / responses, see logger_middleware. That is a completely separate system.

var ERROR = log.New(os.Stderr, "ERROR ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

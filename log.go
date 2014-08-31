package web

import (
	"log"
	"os"
)

// ERROR is the error logger that will log errors in error conditions.
// Error conditions are panic conditions (eg, bad route configuration), as well as
// application panics (eg, division by zero in an app handler).
// Applications can set web.ERROR = your own logger, if they wish.
// In terms of logging the requests / responses, see logger_middleware. That is a completely separate system.
var ERROR = log.New(os.Stderr, "ERROR ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile)

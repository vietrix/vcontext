package common

import (
	"log"
	"os"
)

func NewLogger() *log.Logger {
	return log.New(os.Stderr, "vcontext: ", log.LstdFlags|log.LUTC)
}

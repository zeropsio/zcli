package serviceLogs

/**
  RFC format explained in https://blog.datalust.co/seq-input-syslog/

  RFC 5424 (default):
  <prio>version timestamp(ISO) hostname appName procId msgId [structuredData] message
  e.g. <165>1 2003-10-11T22:14:15.003Z mymachine.example.com evntslog - ID47 [exampleSDID@32473 iut="3"
    eventSource="Application" eventID="1011"] BOMAn application event log entry...

  RFC 3164:
  <prio>timestamp(Mmm dd hh:mm:ss) hostname tag: message
  e.g.<34>Oct 11 22:14:15 mymachine su: 'su root' failed for lonvick on /dev/pts/8
*/

import (
	"fmt"
	"strings"
	"time"
)

func getFullByRfc(logData []Data, rfc string) {
	if rfc == RFC3164 {
		for _, data := range logData {
			fmt.Printf("<%d>%s %s %s: %s\n",
				data.Priority,
				rfc3164TimeFormat(fixTimestamp(data.Timestamp)),
				data.Hostname,
				data.Tag,
				data.Message,
			)
		}
	} else {
		for _, data := range logData {
			fmt.Printf("<%d>1 %v %s %s %s %s - %s\n",
				data.Priority,
				fixTimestamp(data.Timestamp),
				data.Hostname,
				getVal(data.AppName),
				getVal(data.ProcId),
				getVal(data.MsgId),
				data.Message,
			)
		}
	}
}

// add missing 0 to have the same length for all timestamps
func fixTimestamp(timestamp string) string {
	if 27-len(timestamp) == 0 {
		return timestamp
	}
	splitVal := strings.Split(timestamp, ".")
	millis := strings.Split(splitVal[1], "Z")[0]

	for len(millis) < 6 {
		millis += "0"
	}
	return fmt.Sprintf("%s.%sZ", splitVal[0], millis)
}

func rfc3164TimeFormat(timestamp string) string {
	layout := "2006-01-02T15:04:05.000000Z"
	timeStamp, err := time.Parse(layout, timestamp)
	if err != nil {
		return err.Error()
	}
	return timeStamp.Format("Jan 02 15:04:05")
}

// get response log data or "-"
func getVal(outputVal string) string {
	val := "-"
	if outputVal != "" {
		val = outputVal
	}
	return val
}

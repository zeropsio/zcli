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
	"time"
)

func getFullByRfc(jsonData Response, RFC string) {
	for _, data := range jsonData.Items {
		if RFC == RFC3164 {
			fmt.Printf("<%d>%s %s %s: %s\n",
				data.Priority,
				rfc3164TimeFormat(data.Timestamp),
				data.Hostname,
				data.Tag,
				data.Message,
			)
		}

		fmt.Printf("<%d>1 %v %s %s %s %s - %s\n",
			data.Priority,
			data.Timestamp,
			data.Hostname,
			getVal(data.AppName),
			getVal(data.ProcId),
			getVal(data.MsgId),
			data.Message,
		)
	}
}

func rfc3164TimeFormat(timestamp string) string {
	layOut := "2006-01-02T15:04:05.000000Z"
	timeStamp, err := time.Parse(layOut, timestamp)
	if err != nil {
		fmt.Println(err)
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

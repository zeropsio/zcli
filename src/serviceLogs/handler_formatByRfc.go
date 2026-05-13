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

func (h *Handler) getFullByRfc(logData []Data, rfc string) error {
	for _, data := range logData {
		if rfc == RFC3164 {
			if _, err := fmt.Fprintf(h.out, "<%d>%s %s %s: %s\n",
				data.Priority,
				rfc3164TimeFormat(fixTimestamp(data.Timestamp)),
				data.Hostname,
				data.Tag,
				data.Message,
			); err != nil {
				return err
			}
			continue
		}
		// RFC5425
		if _, err := fmt.Fprintf(h.out, "<%d>1 %v %s %s %s %s - %s\n",
			data.Priority,
			fixTimestamp(data.Timestamp),
			data.Hostname,
			getVal(data.AppName),
			getVal(data.ProcId),
			getVal(data.MsgId),
			data.Message,
		); err != nil {
			return err
		}
	}
	return nil
}

// fixTimestamp pads fractional seconds to six digits so all timestamps
// have the same width. Accepts RFC 3339 timestamps with or without a
// fractional-second component; on parse failure returns the input
// unchanged (the downstream Fprintf will still produce *something*).
func fixTimestamp(timestamp string) string {
	t, err := time.Parse(time.RFC3339Nano, timestamp)
	if err != nil {
		return timestamp
	}
	return t.UTC().Format("2006-01-02T15:04:05.000000Z")
}

func rfc3164TimeFormat(timestamp string) string {
	t, err := time.Parse("2006-01-02T15:04:05.000000Z", timestamp)
	if err != nil {
		return err.Error()
	}
	return t.Format("Jan 02 15:04:05")
}

// get response log data or "-"
func getVal(outputVal string) string {
	val := "-"
	if outputVal != "" {
		val = outputVal
	}
	return val
}

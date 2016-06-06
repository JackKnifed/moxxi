package moxxiConf

import "fmt"
import "time"
import "net/http"

// standard merror methods within my application
type Err interface {
	error
	LogError(*http.Request) string
}

// Err - the type used within my application for error handling
type NewErr struct {
	Code    int
	value   string
	deepErr error
}

func UpgradeError(e error) Err {
	return NewErr{Code: ErrUpgradedError, deepErr: e}
}

// the function `Error` to make my custom errors work
func (e NewErr) Error() string {
	switch {
	case e.Code == ErrUpgradedError && e.value == "":
		return e.deepErr.Error()
	case e.deepErr == nil && e.value == "":
		return errMsg[e.Code]
	case e.deepErr == nil && e.value != "":
		return fmt.Sprintf(errMsg[e.Code], e.value)
	case e.value == "" && e.deepErr != nil:
		return fmt.Sprintf(errMsg[e.Code], e.deepErr)
	default:
		return fmt.Sprintf(errMsg[e.Code], e.value, e.deepErr)
	}
}

// the function `LogError` to print error log lines
func (e NewErr) LogError(r *http.Request) string {
	ts := time.Now()
	switch {
	case e.Code == ErrUpgradedError && e.value == "":
		return fmt.Sprintf("%s %s",
			ts.Format("02-Jan-2006:15:04:05-0700"),
			errMsg[e.Code])
	case e.deepErr == nil && e.value == "":
		return fmt.Sprintf("%s %s",
			ts.Format("02-Jan-2006:15:04:05-0700"),
			errMsg[e.Code])
	case e.deepErr == nil && e.value != "":
		return fmt.Sprintf("%s %s %s "+errMsg[e.Code],
			ts.Format("02-Jan-2006:15:04:05-0700"),
			r.RemoteAddr,
			r.RequestURI,
			e.value)
	case e.deepErr != nil && e.value == "":
		return fmt.Sprintf("%s %s %s "+errMsg[e.Code],
			ts.Format("02-Jan-2006:15:04:05-0700"),
			r.RemoteAddr,
			r.RequestURI,
			e.deepErr)
	default:
		return fmt.Sprintf("%s %s %s "+errMsg[e.Code],
			ts.Format("02-Jan-2006:15:04:05-0700"),
			r.RemoteAddr,
			r.RequestURI,
			e.value,
			e.deepErr)
	}
}

// assign a unique id to each error
const (
	ErrUpgradedError = 1 << iota
	ErrCloseFile
	ErrRemoveFile
	ErrFilePerm
	ErrFileUnexpect
	ErrBadHost
	ErrBadIP
	ErrNoRandom
	ErrNoHostname
	ErrNoIP
	ErrConfigBadHost
	ErrConfigBadRead
	ErrConfigBadExtract
	ErrConfigBadStructure
	ErrConfigBadValue
	ErrConfigBadTemplate
)

// specify the error message for each error
var errMsg = map[int]string{
	ErrUpgradedError:      "not actually an error message",
	ErrCloseFile:          "failed to close the file [%s] - %v",
	ErrRemoveFile:         "failed to remove file [%s] - %v",
	ErrFilePerm:           "permission denied to create file [%s] - %v",
	ErrFileUnexpect:       "unknown error with file [%s] - %v",
	ErrBadHost:            "bad hostname provided [%s]",
	ErrBadIP:              "bad IP provided [%s]",
	ErrNoRandom:           "was not given a new random domain - shutting down",
	ErrNoHostname:         "no provided hostname",
	ErrNoIP:               "no provided IP",
	ErrConfigBadHost:      "Bad hostname for handler [%s]",
	ErrConfigBadRead:      "error reading config file - %v",
	ErrConfigBadExtract:   "unable to decode %s portion of config - %v",
	ErrConfigBadStructure: "bad config file - %s of wrong type - %v",
	ErrConfigBadValue:     "bad config file - %s is incorrect - %v",
	ErrConfigBadTemplate:  "bad template at %s - %v",
}

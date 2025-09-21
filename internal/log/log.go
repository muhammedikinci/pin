package log

//go:generate mockgen -source $GOFILE -destination ../mocks/mock_log.go -package mocks
type Log interface {
	Println(v ...interface{})
	Printf(format string, v ...interface{})
}
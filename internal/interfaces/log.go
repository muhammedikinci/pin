package interfaces

//go:generate mockgen -source $GOFILE -destination ../mocks/mock_$GOFILE -package mocks
type Log interface {
	Println(v ...interface{})
	Printf(format string, v ...interface{})
}

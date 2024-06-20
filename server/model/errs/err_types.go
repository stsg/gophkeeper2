package errs

import "fmt"

type DbError struct {
	Err error
}

func (dbErr DbError) Error() string {
	return fmt.Sprintf("db error: %v", dbErr.Err)
}

type DbConnectionError struct {
	Err error
}

func (dbConnErr DbConnectionError) Error() string {
	return fmt.Sprintf("db connection error: %v", dbConnErr.Err)
}

type InternalError struct {
	Err error
}

func (intErr InternalError) Error() string {
	return fmt.Sprintf("internal error: %v", intErr.Err)
}

type TokenError struct {
	Err error
}

func (tknErr TokenError) Error() string {
	return fmt.Sprintf("token error: %v", tknErr.Err)
}

type FileProcessingError struct {
	Err error
}

func (flPrErr FileProcessingError) Error() string {
	return fmt.Sprintf("file processing error: %v", flPrErr.Err)
}

type StreamError struct {
	Err error
}

func (strErr StreamError) Error() string {
	return fmt.Sprintf("stream error: %v", strErr.Err)
}

package server

import "errors"

var (
	ErrApplicationAlreadyInitialised = errors.New("the application is already initialised")
	ErrMissingDatabasePath           = errors.New("please set the database path")
)

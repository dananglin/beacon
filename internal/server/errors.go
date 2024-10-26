package server

import "errors"

var ErrApplicationAlreadyInitialised = errors.New("the application is already initialised")

// Copyright 2015 The go-simplechain Authors
// This file is part of the go-simplechain library.
//
// The go-simplechain library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-simplechain library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-simplechain library. If not, see <http://www.gnu.org/licenses/>.

package rpc

import "fmt"

// request is for an unknown service
type methodNotFoundError struct {
	service string
	method  string
}

func (e *methodNotFoundError) ErrorCode() int { return -32601 }

func (e *methodNotFoundError) Error() string {
	return fmt.Sprintf("The method %s%s%s does not exist/is not available", e.service, serviceMethodSeparator, e.method)
}

// received message isn't a valid request
type invalidRequestError struct{ message string }

func (e *invalidRequestError) ErrorCode() int { return -32600 }

func (e *invalidRequestError) Error() string { return e.message }

// received message is invalid
type invalidMessageError struct{ message string }

func (e *invalidMessageError) ErrorCode() int { return -32700 }

func (e *invalidMessageError) Error() string { return e.message }

// unable to decode supplied params, or an invalid number of parameters
type invalidParamsError struct{ message string }

func (e *invalidParamsError) ErrorCode() int { return -32602 }

func (e *invalidParamsError) Error() string { return e.message }

// logic error, callback returned an error
type callbackError struct{ message string }

func (e *callbackError) ErrorCode() int { return -32000 }

func (e *callbackError) Error() string { return e.message }

// issued when a request is received after the server is issued to stop.
type shutdownError struct{}

func (e *shutdownError) ErrorCode() int { return -32000 }

func (e *shutdownError) Error() string { return "server is shutting down" }

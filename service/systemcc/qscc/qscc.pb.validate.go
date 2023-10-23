// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: systemcc/qscc/qscc.proto

package qscc

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/anypb"
)

// ensure the imports are used
var (
	_ = bytes.MinRead
	_ = errors.New("")
	_ = fmt.Print
	_ = utf8.UTFMax
	_ = (*regexp.Regexp)(nil)
	_ = (*strings.Reader)(nil)
	_ = net.IPv4len
	_ = time.Duration(0)
	_ = (*url.URL)(nil)
	_ = (*mail.Address)(nil)
	_ = anypb.Any{}
	_ = sort.Sort
)

// Validate checks the field values on GetChainInfoRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *GetChainInfoRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on GetChainInfoRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// GetChainInfoRequestMultiError, or nil if none found.
func (m *GetChainInfoRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *GetChainInfoRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for ChannelName

	if len(errors) > 0 {
		return GetChainInfoRequestMultiError(errors)
	}

	return nil
}

// GetChainInfoRequestMultiError is an error wrapping multiple validation
// errors returned by GetChainInfoRequest.ValidateAll() if the designated
// constraints aren't met.
type GetChainInfoRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m GetChainInfoRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m GetChainInfoRequestMultiError) AllErrors() []error { return m }

// GetChainInfoRequestValidationError is the validation error returned by
// GetChainInfoRequest.Validate if the designated constraints aren't met.
type GetChainInfoRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GetChainInfoRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GetChainInfoRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GetChainInfoRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GetChainInfoRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GetChainInfoRequestValidationError) ErrorName() string {
	return "GetChainInfoRequestValidationError"
}

// Error satisfies the builtin error interface
func (e GetChainInfoRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGetChainInfoRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GetChainInfoRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GetChainInfoRequestValidationError{}

// Validate checks the field values on GetBlockByNumberRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *GetBlockByNumberRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on GetBlockByNumberRequest with the
// rules defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// GetBlockByNumberRequestMultiError, or nil if none found.
func (m *GetBlockByNumberRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *GetBlockByNumberRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for ChannelName

	// no validation rules for BlockNumber

	if len(errors) > 0 {
		return GetBlockByNumberRequestMultiError(errors)
	}

	return nil
}

// GetBlockByNumberRequestMultiError is an error wrapping multiple validation
// errors returned by GetBlockByNumberRequest.ValidateAll() if the designated
// constraints aren't met.
type GetBlockByNumberRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m GetBlockByNumberRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m GetBlockByNumberRequestMultiError) AllErrors() []error { return m }

// GetBlockByNumberRequestValidationError is the validation error returned by
// GetBlockByNumberRequest.Validate if the designated constraints aren't met.
type GetBlockByNumberRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GetBlockByNumberRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GetBlockByNumberRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GetBlockByNumberRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GetBlockByNumberRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GetBlockByNumberRequestValidationError) ErrorName() string {
	return "GetBlockByNumberRequestValidationError"
}

// Error satisfies the builtin error interface
func (e GetBlockByNumberRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGetBlockByNumberRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GetBlockByNumberRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GetBlockByNumberRequestValidationError{}

// Validate checks the field values on GetBlockByHashRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *GetBlockByHashRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on GetBlockByHashRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// GetBlockByHashRequestMultiError, or nil if none found.
func (m *GetBlockByHashRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *GetBlockByHashRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for ChannelName

	// no validation rules for BlockHash

	if len(errors) > 0 {
		return GetBlockByHashRequestMultiError(errors)
	}

	return nil
}

// GetBlockByHashRequestMultiError is an error wrapping multiple validation
// errors returned by GetBlockByHashRequest.ValidateAll() if the designated
// constraints aren't met.
type GetBlockByHashRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m GetBlockByHashRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m GetBlockByHashRequestMultiError) AllErrors() []error { return m }

// GetBlockByHashRequestValidationError is the validation error returned by
// GetBlockByHashRequest.Validate if the designated constraints aren't met.
type GetBlockByHashRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GetBlockByHashRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GetBlockByHashRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GetBlockByHashRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GetBlockByHashRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GetBlockByHashRequestValidationError) ErrorName() string {
	return "GetBlockByHashRequestValidationError"
}

// Error satisfies the builtin error interface
func (e GetBlockByHashRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGetBlockByHashRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GetBlockByHashRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GetBlockByHashRequestValidationError{}

// Validate checks the field values on GetTransactionByIDRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *GetTransactionByIDRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on GetTransactionByIDRequest with the
// rules defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// GetTransactionByIDRequestMultiError, or nil if none found.
func (m *GetTransactionByIDRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *GetTransactionByIDRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for ChannelName

	// no validation rules for TxId

	if len(errors) > 0 {
		return GetTransactionByIDRequestMultiError(errors)
	}

	return nil
}

// GetTransactionByIDRequestMultiError is an error wrapping multiple validation
// errors returned by GetTransactionByIDRequest.ValidateAll() if the
// designated constraints aren't met.
type GetTransactionByIDRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m GetTransactionByIDRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m GetTransactionByIDRequestMultiError) AllErrors() []error { return m }

// GetTransactionByIDRequestValidationError is the validation error returned by
// GetTransactionByIDRequest.Validate if the designated constraints aren't met.
type GetTransactionByIDRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GetTransactionByIDRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GetTransactionByIDRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GetTransactionByIDRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GetTransactionByIDRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GetTransactionByIDRequestValidationError) ErrorName() string {
	return "GetTransactionByIDRequestValidationError"
}

// Error satisfies the builtin error interface
func (e GetTransactionByIDRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGetTransactionByIDRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GetTransactionByIDRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GetTransactionByIDRequestValidationError{}

// Validate checks the field values on GetBlockByTxIDRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *GetBlockByTxIDRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on GetBlockByTxIDRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// GetBlockByTxIDRequestMultiError, or nil if none found.
func (m *GetBlockByTxIDRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *GetBlockByTxIDRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for ChannelName

	// no validation rules for TxId

	if len(errors) > 0 {
		return GetBlockByTxIDRequestMultiError(errors)
	}

	return nil
}

// GetBlockByTxIDRequestMultiError is an error wrapping multiple validation
// errors returned by GetBlockByTxIDRequest.ValidateAll() if the designated
// constraints aren't met.
type GetBlockByTxIDRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m GetBlockByTxIDRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m GetBlockByTxIDRequestMultiError) AllErrors() []error { return m }

// GetBlockByTxIDRequestValidationError is the validation error returned by
// GetBlockByTxIDRequest.Validate if the designated constraints aren't met.
type GetBlockByTxIDRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GetBlockByTxIDRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GetBlockByTxIDRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GetBlockByTxIDRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GetBlockByTxIDRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GetBlockByTxIDRequestValidationError) ErrorName() string {
	return "GetBlockByTxIDRequestValidationError"
}

// Error satisfies the builtin error interface
func (e GetBlockByTxIDRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGetBlockByTxIDRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GetBlockByTxIDRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GetBlockByTxIDRequestValidationError{}

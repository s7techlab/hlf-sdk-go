// Code generated by protoc-gen-validate. DO NOT EDIT.
// source: systemcc/cscc/cscc.proto

package cscc

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

// Validate checks the field values on JoinChainRequest with the rules defined
// in the proto definition for this message. If any rules are violated, the
// first error encountered is returned, or nil if there are no violations.
func (m *JoinChainRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on JoinChainRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// JoinChainRequestMultiError, or nil if none found.
func (m *JoinChainRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *JoinChainRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Channel

	if all {
		switch v := interface{}(m.GetGenesisBlock()).(type) {
		case interface{ ValidateAll() error }:
			if err := v.ValidateAll(); err != nil {
				errors = append(errors, JoinChainRequestValidationError{
					field:  "GenesisBlock",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		case interface{ Validate() error }:
			if err := v.Validate(); err != nil {
				errors = append(errors, JoinChainRequestValidationError{
					field:  "GenesisBlock",
					reason: "embedded message failed validation",
					cause:  err,
				})
			}
		}
	} else if v, ok := interface{}(m.GetGenesisBlock()).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return JoinChainRequestValidationError{
				field:  "GenesisBlock",
				reason: "embedded message failed validation",
				cause:  err,
			}
		}
	}

	if len(errors) > 0 {
		return JoinChainRequestMultiError(errors)
	}

	return nil
}

// JoinChainRequestMultiError is an error wrapping multiple validation errors
// returned by JoinChainRequest.ValidateAll() if the designated constraints
// aren't met.
type JoinChainRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m JoinChainRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m JoinChainRequestMultiError) AllErrors() []error { return m }

// JoinChainRequestValidationError is the validation error returned by
// JoinChainRequest.Validate if the designated constraints aren't met.
type JoinChainRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e JoinChainRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e JoinChainRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e JoinChainRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e JoinChainRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e JoinChainRequestValidationError) ErrorName() string { return "JoinChainRequestValidationError" }

// Error satisfies the builtin error interface
func (e JoinChainRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sJoinChainRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = JoinChainRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = JoinChainRequestValidationError{}

// Validate checks the field values on GetConfigBlockRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *GetConfigBlockRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on GetConfigBlockRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// GetConfigBlockRequestMultiError, or nil if none found.
func (m *GetConfigBlockRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *GetConfigBlockRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Channel

	if len(errors) > 0 {
		return GetConfigBlockRequestMultiError(errors)
	}

	return nil
}

// GetConfigBlockRequestMultiError is an error wrapping multiple validation
// errors returned by GetConfigBlockRequest.ValidateAll() if the designated
// constraints aren't met.
type GetConfigBlockRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m GetConfigBlockRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m GetConfigBlockRequestMultiError) AllErrors() []error { return m }

// GetConfigBlockRequestValidationError is the validation error returned by
// GetConfigBlockRequest.Validate if the designated constraints aren't met.
type GetConfigBlockRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GetConfigBlockRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GetConfigBlockRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GetConfigBlockRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GetConfigBlockRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GetConfigBlockRequestValidationError) ErrorName() string {
	return "GetConfigBlockRequestValidationError"
}

// Error satisfies the builtin error interface
func (e GetConfigBlockRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGetConfigBlockRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GetConfigBlockRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GetConfigBlockRequestValidationError{}

// Validate checks the field values on GetChannelConfigRequest with the rules
// defined in the proto definition for this message. If any rules are
// violated, the first error encountered is returned, or nil if there are no violations.
func (m *GetChannelConfigRequest) Validate() error {
	return m.validate(false)
}

// ValidateAll checks the field values on GetChannelConfigRequest with the
// rules defined in the proto definition for this message. If any rules are
// violated, the result is a list of violation errors wrapped in
// GetChannelConfigRequestMultiError, or nil if none found.
func (m *GetChannelConfigRequest) ValidateAll() error {
	return m.validate(true)
}

func (m *GetChannelConfigRequest) validate(all bool) error {
	if m == nil {
		return nil
	}

	var errors []error

	// no validation rules for Channel

	if len(errors) > 0 {
		return GetChannelConfigRequestMultiError(errors)
	}

	return nil
}

// GetChannelConfigRequestMultiError is an error wrapping multiple validation
// errors returned by GetChannelConfigRequest.ValidateAll() if the designated
// constraints aren't met.
type GetChannelConfigRequestMultiError []error

// Error returns a concatenation of all the error messages it wraps.
func (m GetChannelConfigRequestMultiError) Error() string {
	var msgs []string
	for _, err := range m {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// AllErrors returns a list of validation violation errors.
func (m GetChannelConfigRequestMultiError) AllErrors() []error { return m }

// GetChannelConfigRequestValidationError is the validation error returned by
// GetChannelConfigRequest.Validate if the designated constraints aren't met.
type GetChannelConfigRequestValidationError struct {
	field  string
	reason string
	cause  error
	key    bool
}

// Field function returns field value.
func (e GetChannelConfigRequestValidationError) Field() string { return e.field }

// Reason function returns reason value.
func (e GetChannelConfigRequestValidationError) Reason() string { return e.reason }

// Cause function returns cause value.
func (e GetChannelConfigRequestValidationError) Cause() error { return e.cause }

// Key function returns key value.
func (e GetChannelConfigRequestValidationError) Key() bool { return e.key }

// ErrorName returns error name.
func (e GetChannelConfigRequestValidationError) ErrorName() string {
	return "GetChannelConfigRequestValidationError"
}

// Error satisfies the builtin error interface
func (e GetChannelConfigRequestValidationError) Error() string {
	cause := ""
	if e.cause != nil {
		cause = fmt.Sprintf(" | caused by: %v", e.cause)
	}

	key := ""
	if e.key {
		key = "key for "
	}

	return fmt.Sprintf(
		"invalid %sGetChannelConfigRequest.%s: %s%s",
		key,
		e.field,
		e.reason,
		cause)
}

var _ error = GetChannelConfigRequestValidationError{}

var _ interface {
	Field() string
	Reason() string
	Key() bool
	Cause() error
	ErrorName() string
} = GetChannelConfigRequestValidationError{}

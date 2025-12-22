// Copyright (C) 2026 Storj Labs, Inc.
// See LICENSE for copying information.

// Package main implements a linter that validates zap logger field names.
//
// # Overview
//
// The check-zap-fields linter ensures that field names used in zap logger calls
// follow best practices. Valid field names must:
//   - Not be empty
//   - Only contain lowercase ASCII letters (a-z)
//   - Only contain numbers (0-9)
//   - Only contain underscores (_), but only in the middle (not at start or end)
//   - Not contain spaces
//   - Not contain uppercase letters
//   - Not contain special characters
//
// # Usage
//
// Run the linter on a package or directory:
//
//	check-zap-fields ./...
//	check-zap-fields ./path/to/package
//
// # Examples
//
// Valid field names:
//
//	logger.Info("message", zap.String("user_id", id))
//	logger.Error("error", zap.Int("status_code", 500))
//	logger.Debug("debug", zap.Duration("elapsed_time", dur))
//
// Invalid field names:
//
//	logger.Info("message", zap.String("User ID", id))        // spaces and uppercase
//	logger.Error("error", zap.Int("Status-Code", 500))      // uppercase and dash
//	logger.Debug("debug", zap.Duration("_elapsed", dur))    // leading underscore
//
// # Ignoring Violations
//
// There are three ways to ignore linter violations when necessary:
//
// ## Line-Level Ignore
//
// Add a comment on the line before or on the same line as the violation:
//
//	// On the line before
//	//zapfields:ignore
//	logger.Info("test", zap.String("Invalid Field", "value"))
//
//	// As an inline comment
//	logger.Info("test", zap.String("Bad Name", "value")) //zapfields:ignore
//
// NOTE that the directive must not have a space after //.
//
// ## File-Level Ignore
//
// Ignore all violations in a file by adding a comment in the first 50 lines
// after the package keyword:
//
//	package mypackage
//
//	//zapfields:ignore-file
//
//	// All zap field violations in this file will be ignored
//
// NOTE that the directive must not have a space after //.
// # Detected Zap Field Functions
//
// The linter checks field names in the following zap functions:
//   - Basic types: String, Strings, Bool, Bools, Int, Int64, Int32, Int16, Int8,
//     Uint, Uint64, Uint32, Uint16, Uint8, Uintptr, Float64, Float32,
//     Complex128, Complex64
//   - Time types: Duration, Durations, Time, Times
//   - Error types: Errors, NamedError
//   - Other types: Any, Reflect, Stringer, ByteString, ByteStrings, Binary,
//     Namespace, Inline, Object, Array, Stack, StackSkip
//
// Note: zap.Error is not checked as it only takes an error value, not a field name.
package main

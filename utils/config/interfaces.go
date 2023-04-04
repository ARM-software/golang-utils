/*
 * Copyright (C) 2020-2022 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

// Package config provides utilities to load configuration from an environment and perform validation at load time.
package config

//go:generate mockgen -destination=../mocks/mock_$GOPACKAGE.go -package=mocks github.com/ARM-software/golang-utils/utils/$GOPACKAGE IServiceConfiguration,Validator

// IServiceConfiguration defines a typical service configuration.
type IServiceConfiguration interface {
	Validator
}

// Validator defines an object which can perform some validation on itself.
type Validator interface {
	Validate() error
}

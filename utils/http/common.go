/*
 * Copyright (C) 2020-2021 Arm Limited or its affiliates and Contributors. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */
package http

import "net/http"

func setTransportConfiguration(cfg *HttpClientConfiguration, transport *http.Transport) {
	if cfg == nil || transport == nil {
		return
	}
	transport.IdleConnTimeout = cfg.IdleConnTimeout
	transport.ExpectContinueTimeout = cfg.ExpectContinueTimeout
	transport.TLSHandshakeTimeout = cfg.TLSHandshakeTimeout
	transport.MaxIdleConns = cfg.MaxIdleConns
	transport.MaxConnsPerHost = cfg.MaxConnsPerHost
	transport.MaxIdleConnsPerHost = cfg.MaxIdleConnsPerHost
}

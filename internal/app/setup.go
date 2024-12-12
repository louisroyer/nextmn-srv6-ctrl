// Copyright 2024 Louis Royer and the NextMN contributors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.
// SPDX-License-Identifier: MIT

package app

import (
	"context"
	"time"

	pfcp_networking "github.com/nextmn/go-pfcp-networking/pfcp"

	"github.com/nextmn/srv6-ctrl/internal/config"
)

type Setup struct {
	HTTPServer  *HttpServerEntity
	Upf         *Upf
	PFCPServer  *pfcp_networking.PFCPEntityUP
	RulesPusher *RulesPusher
}

func NewSetup(conf *config.CtrlConfig) Setup {
	upf := NewUpf(conf)
	return Setup{
		HTTPServer:  NewHttpServer(conf, upf.pfcpentity),
		Upf:         upf,
		RulesPusher: NewRulesPusher(conf),
	}
}

func (s Setup) Run(ctx context.Context) error {
	if err := PFCPServerAddHooks(s.PFCPServer, s.RulesPusher); err != nil {
		return err
	}
	if err := s.Upf.Start(ctx); err != nil {
		return err
	}
	if err := s.HTTPServer.Start(ctx); err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		// Try to end before 1s…
		s.HTTPServer.WaitShutdown(ctxShutdown)
		s.Upf.WaitShutdown(ctxShutdown)
	}
	return nil
}

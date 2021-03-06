// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//+build !wireinject

package http

import (
	"context"
	"crypto/tls"
	"github.com/elastic/harp-plugins/cmd/harp-server/internal/config"
	"github.com/elastic/harp-plugins/cmd/harp-server/internal/dispatchers/http/routes"
	"github.com/elastic/harp-plugins/cmd/harp-server/pkg/server/manager"
	"github.com/elastic/harp-plugins/cmd/harp-server/pkg/server/storage/backends/container"
	"github.com/elastic/harp/pkg/sdk/log"
	"github.com/elastic/harp/pkg/sdk/tlsconfig"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// Injectors from wire.go:

func setup(ctx context.Context, cfg *config.Configuration) (*http.Server, error) {
	backend, err := backendManager(ctx, cfg)
	if err != nil {
		return nil, err
	}
	server, err := httpServer(ctx, cfg, backend)
	if err != nil {
		return nil, err
	}
	return server, nil
}

// wire.go:

func backendManager(ctx context.Context, cfg *config.Configuration) (manager.Backend, error) {

	bm := manager.Default()

	for _, b := range cfg.Backends {

		if err := bm.Register(ctx, b.NS, b.URL); err != nil {
			return nil, err
		}
	}

	return bm, nil
}

func httpServer(ctx context.Context, cfg *config.Configuration, bm manager.Backend) (*http.Server, error) {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Use(middleware.Timeout(60 * time.Second))
	container.SetKeyring(cfg.Keyring)

	backendRouter, err := routes.Backends(ctx, cfg, bm)
	if err != nil {
		return nil, err
	}

	r.Route("/api/v1", func(r chi.Router) {
		r.Mount("/", http.StripPrefix("/api/v1", backendRouter))
	})

	server := &http.Server{
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
		Handler:           r,
	}

	if cfg.HTTP.UseTLS {

		clientAuth := tls.VerifyClientCertIfGiven
		if cfg.HTTP.TLS.ClientAuthenticationRequired {
			clientAuth = tls.RequireAndVerifyClientCert
		}

		tlsConfig, err := tlsconfig.Server(&tlsconfig.Options{
			KeyFile:    cfg.HTTP.TLS.PrivateKeyPath,
			CertFile:   cfg.HTTP.TLS.CertificatePath,
			CAFile:     cfg.HTTP.TLS.CACertificatePath,
			ClientAuth: clientAuth,
		})
		if err != nil {
			log.For(ctx).Error("Unable to build TLS configuration from settings", zap.Error(err))
			return nil, err
		}

		server.TLSConfig = tlsConfig
	} else {
		log.For(ctx).Info("No transport encryption enabled for HTTP server")
	}

	return server, nil
}

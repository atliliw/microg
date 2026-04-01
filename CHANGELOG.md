# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

### Added
- Initial release
- gRPC server with interceptors (crash recovery, tracing, metrics, timeout)
- HTTP server based on Gin
- Consul service registration and discovery
- OpenTelemetry tracing support
- Prometheus metrics support
- Structured logging with zap
- Error handling with error codes
- Graceful shutdown support

## [1.0.0] - 2024-01-01

### Added
- `app` package for application lifecycle management
- `server/rpcserver` gRPC server with built-in interceptors
- `server/restserver` HTTP server based on Gin
- `registry/consul` Consul service registry implementation
- `log` structured logging interface
- `errors` error handling with error codes
- `core/trace` OpenTelemetry tracing
- `core/metric` Prometheus metrics
- Examples for basic, HTTP, and full microservice
# stereoscope

[![Go Report Card](https://goreportcard.com/badge/github.com/anchore/stereoscope)](https://goreportcard.com/report/github.com/anchore/stereoscope)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/anchore/stereoscope.svg)](https://github.com/anchore/stereoscope)
[![License: Apache-2.0](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/anchore/stereoscope/blob/main/LICENSE)
[![Slack Invite](https://img.shields.io/badge/Slack-Join-blue?logo=slack)](https://anchore.com/slack)

A library for working with container image contents, layer file trees, and squashed file trees.

## Getting Started

See `examples/basic.go`

```bash
docker image save centos:8 -o centos.tar
go run examples/basic.go ./centos.tar
```

Note: To run tests you will need `skopeo` installed.

## Overview

This library provides the means to:
- parse and read images from multiple sources, supporting:
  - docker V2 schema images from the docker daemon, podman, or archive
  - OCI images from disk, directory, or registry
  - singularity formatted image files
- build a file tree representing each layer blob
- create a squashed file tree representation for each layer
- search one or more file trees for selected paths
- catalog file metadata in all layers
- query the underlying image tar for content (file content within a layer)

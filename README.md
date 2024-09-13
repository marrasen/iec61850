# iec61850

[![License](https://img.shields.io/badge/license-Apache--2.0-green.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)
[![PkgGoDev](https://pkg.go.dev/badge/mod/github.com/morris-kelly/iec61850)](https://pkg.go.dev/mod/github.com/morris-kelly/iec61850)
![Go Version](https://img.shields.io/badge/go%20version-%3E=1.0-61CFDD.svg?style=flat-square)
[![Go Report Card](https://goreportcard.com/badge/github.com/morris-kelly/iec61850?style=flat-square)](https://goreportcard.com/report/github.com/morris-kelly/iec61850)

English | [中文](README_zh_CN.md)

cgo version of IEC61850 library, reference [libiec61850](https://github.com/mz-automation/libiec61850)

## Overview

iec61850 is an open source (Apache-2.0 license) implementation of the IEC 61850 client and server library that implements the MMS, GOOSE and SV protocols.
It can be used to implement IEC 61850 compliant clients and PCs on embedded systems and PCs running Linux, Windows Server application.
This project relies on and refers to [libiec61850](https://github.com/mz-automation/libiec61850).

## Features

The library support the following IEC 61850 protocol features:

- MMS client/server, GOOSE (IEC 61850-8-1)
- Sampled Values (SV - IEC 61850-9-2)
- Support for buffered and unbuffered reports
- Online report control block configuration
- Data access service (get data, set data)
- online data model discovery and browsing
- all data set services (get values, set values, browse)
- dynamic data set services (create and delete)
- log service
- MMS file services (browse, get file, set file, delete/rename file)
- Setting group handling
- Support for service tracking
- GOOSE and SV control block handling
- TLS support

## How to use

```shell
go get -u github.com/morris-kelly/iec61850
```

- [Client reads and writes values](test/client_test.go)
- [Client control](test/client_control_test.go)
- [Client reads icd file](test/scl_test.go)
- [Create server](test/server_test.go)

## License

iec61850 is based on the [GPL-3.0 license](./LICENSE) agreement, and iec61850 relies on some third-party components whose open source agreement is GPL-3.0 and MIT.

## Contact

- Email：<wendy512@yeah.net>

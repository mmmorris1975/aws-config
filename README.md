# aws-config
[![Go Report Card](https://goreportcard.com/badge/github.com/mmmorris1975/aws-config)](https://goreportcard.com/report/github.com/mmmorris1975/aws-config)

Utility library to handle AWS cli/sdk config and credential data.

The library provides some AWS wrapping around the `go-ini` library in order handle some idiosyncrasies around profile
section naming.  It also provides a (hopefully) simple interface for managing the credentials file for multiple profiles.

# aws-config
Utility library to handle AWS cli/sdk config and credential files.

The library provides some AWS wrapping around the `go-ini` library in order handle some idiosyncrasies around profile
section naming.  It also provides a (hopefully) simple interface for managing the credentials file for multiple profiles.

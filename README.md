# GCS Consolidate mtimes

A little utility to restore file mtimes following a migration from bare-metal to gcsfuse-mounted Google Cloud Storage buckets.

## Problem

- Files copied by `gsutil rsync` record mtimes in the metadata keyed `goog-reserved-file-mtime`, formatted as a unix timestamp
- gcsfuse records/consumed mtimes from the metadata keyed `gcsfuse_mtime`, formatted as per RFC-3339
- The keys are different
- Files viewed via gcsfuse will report their mtimes as the date & time the rsync occured

## Solution

- Interpet the rsync's mtimes (`goog-reserved-file-mtime`)
- Translate them into an RFC-3339 compatible string
- Jam that bad boy into gcsfuse's metadata attribute (`gcsfuse_mtime`)

The utility will skip objects that:

- Do not have the rsync mtime (this object was not rsynced)
- Already has a gcsfuse-compatible mtime (this object has been created / updated via gcsfuse)

## Build

```
go build -o gcs_consolidate_mtimes cmd/gcs_consolidate_mtimes/main.go

# Or to deploy on a Linux server, if you're building from another platform
env GOOS=linux GOARCH=amd64 go build -o gcs_consolidate_mtimes cmd/gcs_consolidate_mtimes/main.go
```

## Usage

```
GOOGLE_APPLICATION_CREDENTIALS=/a/b/c.json gcs_consolidate_mtimes example-bucket
```
[![Build Status](https://travis-ci.org/pakesson/cfs-server-go.svg?branch=master)](https://travis-ci.org/pakesson/cfs-server-go)

# CFS Go Server

See the [CFS](https://github.com/pakesson/cfs) repository for more info.

## Dependencies

 * [The AWS Go SDK](https://github.com/aws/aws-sdk-go)
 * [gorilla/mux](http://www.gorillatoolkit.org/pkg/mux)

## Building
```bash
 $ go get -u github.com/pakesson/cfs-server-go
```

## Running

The environment variables `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` will
be automatically recognized and used. Alternatively, credentials can be stored
in `~/.aws/credentials`. See the AWS SDK docs for more info.

Additionally, `AWS_REGION` and `S3_BUCKET` must be set.

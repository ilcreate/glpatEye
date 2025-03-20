# glpatEye

## Description
`glpatEye` - is a implement idea as monitoring Gitlab project access tokens and group access tokens. 

## Usage

In both configurations, you must use main environment variable: `GITLAB_TOKEN`. 
Otherwise, application won't started.

You can use this app for 2 ways:

#### 1: Using config file (config.yaml) for app.

In this case, everything is simply. You can mount config file with `-v` with `docker run` or mount it in `docker-compose` with directive `volumes`. There is a example config in directory `configs`.

#### 2: Using environment variables. 

This method requires some environment variables: \
`GITLAB_URL` - url your Gitlab instance. \
`GITLAB_PATTERN` - regex pattern for searching tokens by name. \
`CRON` - period of checkout and updating metrics. \
`OBJECTS_PER_PAGE` - quatity of returned objects per 1 request from Gitlab API. (Maximum: 100) \
`POOL_SIZE` - size of goroutines pool for checking tokens. (maybe it's a useless variable, because Gitlab returns maximum 100 objects from API, and the pool isn't completely utilized). \
`SERVER_PORT` - listening port for exported metrics.

## Support
You can create an issue so that I can make some kind of revision, etc. No one is stopping you from making a fork of my project and fine-tuning it the way you need it.

## Authors and acknowledgment

- Owner, main contributor - [@ilcreate](https://github.com/ilcreate)

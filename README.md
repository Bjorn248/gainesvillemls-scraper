# gainesvillemls-scraper
A simple tool that parses, filters, and notifies users of listings on http://www.gainesvillemls.com/

Runs every hour, designed to run in heroku. Only emails new listings.

[![Build Status](https://travis-ci.org/BjornTwitchBot/gainesvillemls-scraper.svg?branch=master)](https://travis-ci.org/BjornTwitchBot/gainesvillemls-scraper)

## Dependencies
* [Redis](http://redis.io/)
* [Sendgrid](https://sendgrid.com/)

## Environment Variables
Variable Name | Description
------------ | -------------
REDIS_HOST_PORT | The host:port for redis (e.g. localhost:6379)
REDIS_PASSWORD | The password used to AUTH to redis
SENDGRID_API_TOKEN | Sendgrid API Token for Email
EMAIL_FROM_ADDRESS | When the application sends email, the From address
EMAIL_TO_ADDRESS | When the application sends email, the To address

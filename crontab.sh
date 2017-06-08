#!/usr/bin/env bash
crontab -l > mycron
echo "$MIN $HOUR $DAYOFMONTH $MONTH $DAYOFWEEK $COMMAND"
echo "$MIN $HOUR $DAYOFMONTH $MONTH $DAYOFWEEK $COMMAND" >> mycron
crontab mycron
rm mycron

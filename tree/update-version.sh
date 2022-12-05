#!/usr/bin/env bash

cat > version.go <<STR
package main

var appVersion = "$(date "+%Y-%m-%d-%H:%M:%S")"
STR

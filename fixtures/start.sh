#!/bin/sh
dnet_ioserv -c ./ioserv.json&
bg && disown

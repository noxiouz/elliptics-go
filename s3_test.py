#!/usr/bin/env python

import time

import boto
import boto.s3.connection
from boto.s3 import key

suffix = "%d" % int(time.time())
access_key = 'noxiouz'
secret_key = 'noxiouz'
host = "localhost"

test_bucket = "testbucket" + suffix
test_key = "testkey" + suffix

conn = boto.connect_s3(aws_access_key_id=access_key,
                       aws_secret_access_key=secret_key,
                       host=host,
                       port=9000,
                       debug=10,
                       is_secure=False,
                       calling_format=boto.s3.connection.OrdinaryCallingFormat(),
                       )

print("Connect to bucket %s" % test_bucket)
b = conn.get_bucket(test_bucket)
b = conn.create_bucket(test_bucket)
# k = key.Key(b)
# k.key = 'xxx.jpg'
# print k.exists()
# print "SetContent %s" % k.set_contents_from_string("TEST")
# print k.get_contents_as_string()
# print k.set_metadata('meta1', 'This is the first metadata value')
# print k.set_contents_from_string("push through proxy")


# possible_key = b.get_key('xxx.jpg')
# print possible_key
# possible_key = b.get_key('xxxdsdsdsd.jpg')
# print possible_key

# res = conn.create_bucket('mybucket')
# print res

# k = key.Key(b)
# k.key = 'testKey'

#!/usr/bin/env python

import boto
import boto.s3.connection
from boto.s3 import key
access_key = 'put your access key here!'
secret_key = 'put your secret key here!'

# host = 'cocaine-cloud02g.kit.yandex.net'
host = "localhost"

conn = boto.connect_s3(aws_access_key_id=access_key,
                       aws_secret_access_key=secret_key,
                       host=host,
                       port=9000,
                       is_secure=False,
                       calling_format=boto.s3.connection.OrdinaryCallingFormat(),
                       # calling_format=boto.s3.connection.VHostCallingFormat(),
                       )

# b = conn.get_bucket('testns')
# k = key.Key(b)
# k.key = 'xxx.jpg'
# # print k.get_contents_as_string()
# print k.set_metadata('meta1', 'This is the first metadata value')
# print k.set_contents_from_string("push through proxy")


# possible_key = b.get_key('xxx.jpg')
# print possible_key
# possible_key = b.get_key('xxxdsdsdsd.jpg')
# print possible_key

res = conn.create_bucket('mybucket')
print res

# k = key.Key(b)
# k.key = 'testKey'

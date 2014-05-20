#!/usr/bin/env python

import time

import boto
import boto.s3.connection
from boto.s3 import key
from boto.exception import S3ResponseError

import pytest

suffix = "%d" % int(time.time())
access_key = 'noxiouz'
secret_key = 'noxiouz'
host = "localhost"

test_bucket = "testbucket" + suffix
test_key = "testkey" + suffix


class TestS3:
    def setup_class(self):
        self.conn = boto.connect_s3(aws_access_key_id=access_key,
                                    aws_secret_access_key=secret_key,
                                    host=host,
                                    port=9000,
                                    debug=10,
                                    is_secure=False,
                                    calling_format=boto.s3.connection.OrdinaryCallingFormat(),
                                    )
        self.bucket = None

    def test_bucket(self):
        with pytest.raises(S3ResponseError):
            self.conn.get_bucket(test_bucket)
        self.bucket = self.conn.create_bucket(test_bucket)
        self.bucket = self.conn.get_bucket(test_bucket)

    def test_key(self):
        self.bucket = self.conn.get_bucket(test_bucket)
        k = key.Key(self.bucket)
        k.key = test_key
        k.set_contents_from_string("TEST")
        assert True, k.exists()
        possible_key = self.bucket.get_key('xxx.jpg')
        print possible_key

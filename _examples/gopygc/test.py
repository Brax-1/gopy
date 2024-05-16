# Copyright 2020 The go-python Authors.  All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# py2/py3 compat
from __future__ import print_function

import gopygc
import _gopygc

print(_gopygc.NumGopyvars())


# test literals
a = gopygc.StructA()
b = gopygc.SliceA()
c = gopygc.MapA()
print(_gopygc.NumGopyvars())
del a
del b
del c

print(_gopygc.NumGopyvars())
a = [gopygc.StructValue(), gopygc.StructValue(), gopygc.StructValue()]
print(_gopygc.NumGopyvars())  # 3
b = [gopygc.SliceScalarValue(), gopygc.SliceScalarValue()]
print(_gopygc.NumGopyvars())  # 5
c = gopygc.SliceStructValue()
print(_gopygc.NumGopyvars())  # 6
d = gopygc.MapValue()
print(_gopygc.NumGopyvars())  # 7
e = gopygc.MapValueStruct()
print(_gopygc.NumGopyvars())  # 8

del a
print(_gopygc.NumGopyvars())  # 5
del b
print(_gopygc.NumGopyvars())  # 3
del c
print(_gopygc.NumGopyvars())  # 2
del d
print(_gopygc.NumGopyvars())  # 1
del e
print(_gopygc.NumGopyvars())  # 0

e1 = gopygc.ExternalType()
print(_gopygc.NumGopyvars())  # 1
del e1
print(_gopygc.NumGopyvars())  # 0

# test reference counting
f = gopygc.SliceStructValue()
print(_gopygc.NumGopyvars())  # 1
g = gopygc.StructA(gopyvar=f.gopyvar)
print(_gopygc.NumGopyvars())  # 1
del g
print(_gopygc.NumGopyvars())  # 1
del f
print(_gopygc.NumGopyvars())  # 0


print("OK")

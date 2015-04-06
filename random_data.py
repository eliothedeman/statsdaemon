import statsd
import sys
import random
import time
import string

def genTime():
	return time.time() + randInt()

def randInt():
	return random.randint(-1000,1000)

def randFloat():
	return random.randrange(10000)

def randString():
	return ''.join(random.choice(string.ascii_uppercase) for i in range(12))

def sendTime():
	t = statsd.Timer(randString())
	t.send("",randFloat())

def sendCounter():
	c = statsd.Counter(randString())
	c + random.randint(0,100000)

def sendGauge():
	g = statsd.Gauge(randString())
	g.send("",randFloat())



funcs = [sendTime,sendCounter,sendGauge]


num = 1000

if len(sys.argv) > 1:
	num = int(sys.argv[1])

for i in range(0, num):
	funcs[i % len(funcs)]()



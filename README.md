statsdaemon
==========

Fork of Bitly's statsdaemon which allows for custom backends

Supports

* Timing (with optional percentiles)
* Counters (positive and negative with optional sampling)
* Gauges (including relative operations)
* Sets
* Custome backends

Installing
==========

```bash
go get github.com/eliothedeman/statsdaemon
```
Config File Format
====================

```javascript
{
	"flush_interval": "1s",
	"backends": {
		"console": {

		}
	}
}
```


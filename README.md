### hashR

Simple tool that walks a directory recursively,
and outputs a list of hashes suitable for Redis.

**Usage:**
```
# Write directly to Redis
./hashr /directory | redis-cli

# Write to file and import
./hashr /directory > hashes.redis
redis-cli --pipe < hashes.redis
```

**Info:**

* The output (on stdout) looks like this:
  ```
  SET "<filename>" "<file size>|<extension>|<md5>|<sha1>|<sha256>|<sha512>"
  ```
* All hashes are hex-encoded.
* `-threads=x` (default: number of cores)
* `-prefix=/path/to/dir` (pretend files are in this dir)
* `-key=HASHDB` (Redis key of the [HASHES](https://redis.io/commands#hash) object)

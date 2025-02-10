[![progress-banner](https://backend.codecrafters.io/progress/redis/e0c3f2e0-21f6-4f38-812f-a78489eb9959)](https://app.codecrafters.io/users/codecrafters-bot?r=2qF)

This is  **my solution** for Codecraftes challenge:

[&#34;Build Your Own Redis&#34; Challenge](https://codecrafters.io/challenges/redis).

### Dependencies

Install redis-cli: [instructions
](https://redis.io/docs/latest/operate/oss_and_stack/install/install-redis/)

Install go: [instructions](https://go.dev/doc/install)

# Running the project

```sh
./run.sh
```

## Use as the Client

In another terminal, simulate the client using redis-cli.
You can test the existing commands: PING, ECHO, SET, GET, LIST, DELETE.
Any command can be sent in **camelcase**, **uppercase** or **lowercase**.

Key not found or return **nil**: `$-1`

```sh
redis-cli ping
PONG
```

```sh
redis-cli Echo hey
hey
```

```sh
redis-cli set foo bar
OK
```

```sh
redis-cli get foo
bar
```

```sh
redis-cli list
1) "foo"
```

```sh
redis-cli DEL foo
1  # deleted key numbers
```

```sh
redis-cli set foo bar px 100 # Sets the key "foo" to "bar" with an expiry of 100 milliseconds
OK
```

```sh
redis-cli get foo
$-1
```

Key not found or return **nil**: `$-1`

In another scenario, Delete **(DEL)** can receive more than one key, for example in Store there is [foo, foo2, foo3]:

```sh
redis-cli del foo foo3
2
```

```sh
redis-cli list
1) "foo2"
```

# Tests

Runnig tests in local:

```
gotestsum --junitfile test-results.xml --format testname
```


__

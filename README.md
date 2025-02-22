Server runs on locahost:9000
The proxy runs on locahost:8000

Doing a curl `locahost:8000` should give us response from `locahost:9000`

```shell
➜  ~ curl localhost:9000
hello world!
```
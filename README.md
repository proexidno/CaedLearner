# CardLearner
## Command Build/Run
```console
go build -o main cmd/*.go
./main
```

## Docker
```console
$ docker build -t proexidno/cardlearner:1.0 .
$ docker run --detach -it \
      --volume=./data:/data \
      -e TELEGRAM_API_TOKEN='foo bar' \
      --name cardlearner proexidno/cardlearner:1.0
```

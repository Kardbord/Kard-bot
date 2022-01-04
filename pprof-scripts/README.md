# pprof Scripts

Per the [pprof repo](https://github.com/google/pprof) and [module docs](https://pkg.go.dev/net/http/pprof):

> pprof is a tool for visualization and analysis of profiling data.

Kard-bot will optionally start a pprof server if configured to do so in `config/setup.json`.
The intention behind the scripts in this directory (and this README) is to give me a nudge in
the right direction when I inevitably can't remember how to use pprof. :)

These scripts are intended to be run from the directory they live in. They depend on a lot of relative paths,
so they probably won't work if you run them from elsewhere. ¯\\_(ツ)_/¯

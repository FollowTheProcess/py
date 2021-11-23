# py

[![License](https://img.shields.io/github/license/FollowTheProcess/py)](https://github.com/FollowTheProcess/py)
[![Go Report Card](https://goreportcard.com/badge/github.com/FollowTheProcess/py)](https://goreportcard.com/report/github.com/FollowTheProcess/py)
[![GitHub](https://img.shields.io/github/v/release/FollowTheProcess/py?logo=github&sort=semver)](https://github.com/FollowTheProcess/py)
[![CI](https://github.com/FollowTheProcess/py/workflows/CI/badge.svg)](https://github.com/FollowTheProcess/py/actions?query=workflow%3ACI)

 (Approximate) port of Brett Cannon's [python-launcher] for Unix to Go, with a few tweaks 😉

* Free software: MIT License

## Project Description

**Description of the original [python-launcher]:**

> *Taken directly from the official [README]*

Launch your Python interpreter the lazy/smart way!

This project is an implementation of the `py` command for Unix-based platforms
(with some potential experimentation for good measure 😉).

The goal is to have `py` become the cross-platform command that Python users
typically use to launch an interpreter while doing development. By having a
command that is version-agnostic when it comes to Python, it side-steps
the "what should the `python` command point to?" debate by clearly specifying
that upfront (i.e. the newest version of Python that can be found). This also
unifies the suggested command to document for launching Python on both Windows
as Unix as `py` has existed as the preferred
[command on Windows](https://docs.python.org/3/using/windows.html#launcher)
since 2012 with the release of Python 3.3.

Typical usage would be:

``` shell
py -m venv .venv
py ...  # Normal `python` usage.
```

This creates a virtual environment in a `.venv` directory using the latest
version of Python installed. Subsequent uses of `py` will then use that virtual
environment as long as it is in the current (or higher) directory; no
environment activation required (although the Python Launcher supports activated
environments as well)!

A non-goal of this project is to become the way to launch the Python
interpreter _all the time_. If you know the exact interpreter you want to launch
then you should launch it directly; same goes for when you have
requirements on the type of interpreter you want (e.g. 32-bit, framework build
on macOS, etc.). The Python Launcher should be viewed as a tool of convenience,
not necessity.

### Why are you reimplementing it in Go?

* I don't know, fun I guess?
* I love the original [python-launcher] and I love Go, so why not combine them!
* Learning and stuff

### What's different about this one vs the original?

Initially, I wanted to do a 100% pure port copying the functionality of the original [python-launcher] exactly.

Then I realised that's boring and pointless because why wouldn't I just use the original? It's written really well, easy to understand,
is fast to launch, 100% test coverage etc etc.

So I thought why not tweak it a little bit?

As a result, this version behaves slightly differently in a few ways:

1. It won't let you do anything with `python2`, because it's deprecated and using it is naughty! In fact, it completely ignores any python2 interpreters it finds, so if you use this `py` there is 0 chance of accidentally launching `python2`. You're welcome macOS users!
2. It won't climb the file tree looking for a `.venv` in any parent directory, it only looks in `cwd` (personally I only ever really use python in a virtual environment when I'm actively working on a python project, and 99% of the time for that I'm sitting in the project root where the `.venv` is anyway)
3. The change above allows this one to easily support both virtual environments named `.venv` **and** `venv` (although `.venv` will be preferred)

## Installation

There are binaries published in the [GitHub releases] section, and a homebrew formula:

```shell
brew tap FollowTheProcess/homebrew-tap

brew install FollowTheProcess/homebrew-tap/py
```

## Quickstart

### Launch the most appropriate python

```shell
py ...
```

### Launch the latest python3 on `$PATH`

```shell
py -3 ...
```

### Launch an exact version

```shell
py -3.10 ...
```

### Debugging

If you want to see what `py` is doing to find your python, set the `PYLAUNCH_DEBUG` environment variable to 1 (or anything really, the value doesn't matter) before running `py`.

You will see something like this:

```shell
PYLAUNCH_DEBUG=1 py -c "import sys; print(sys.executable)"
DEBUG py called with multiple arguments             arguments="[-c import sys; print(sys.executable)]"
DEBUG Unrecognised arguments                        arguments="[-c import sys; print(sys.executable)]"
DEBUG Looking for $VIRTUAL_ENV environment variable
DEBUG Looking for virtual environment in cwd        cwd=/Users/tomfleet/Development/py
DEBUG Looking for $PY_PYTHON environment variable
DEBUG Found environment variable                    $PY_PYTHON=3.10
DEBUG Searching for exact python version            version=3.10
DEBUG Checking $PATH environment variable
DEBUG $PATH: [/Users/tomfleet/.pyenv/shims /Users/tomfleet/go/bin /Users/tomfleet/.local/bin /Users/tomfleet/.cargo/bin /usr/local/bin /usr/bin /bin /usr/sbin /sbin]
DEBUG Looking through $PATH for python3 interpreters
DEBUG Found matching interpreters                   matching interpreters="[/Users/tomfleet/.pyenv/shims/python3.10]"
DEBUG Launching exact python                        python=/Users/tomfleet/.pyenv/shims/python3.10
/Users/tomfleet/.pyenv/versions/3.10.0/bin/python3.10
```

## Control Flow

As previously mentioned, this experimental port behaves slightly differently than the original [python-launcher]. The adjusted control flow diagram is shown below:

![control_flow](https://raw.githubusercontent.com/FollowTheProcess/py/main/docs/control_flow/control_flow.svg)

## Benchmarks

Although I've not made any special efforts to optimise `py`, it is very close to the original [python-launcher] in terms of performance:

* **Left:** This version of `py`, written in Go
* **Right:** The original [python-launcher] written in Rust

![comparison](https://raw.githubusercontent.com/FollowTheProcess/py/main/docs/img/comp.png)

## FAQ

### Does this version still work with [Starship] and display the Python version?

**Short answer:** Yes! :tada: If you already have [Starship] set up for the original [python-launcher], this one should work just fine!

To set it up:

Add the following to your [Starship configuration file]:

```TOML
[python]
python_binary = ["py"]
# The following isn't necessary, but convenient.
detect_folders = [".venv"]
```

![Starship virtual env demo](https://github.com/FollowTheProcess/py/raw/main/docs/img/starship_demo.png)

By using the Launcher with Starship, your prompt will tell you which Python version will be used if you run `py`. Since the Launcher supports virtual
environments, the prompt will properly reflect both what global install of Python will be used, but also the local virtual environment.

[python-launcher]: https://github.com/brettcannon/python-launcher
[README]: https://github.com/brettcannon/python-launcher/blob/main/README.md
[Github releases]: https://github.com/FollowTheProcess/py/releases
[Starship]: https://starship.rs/
[Starship configuration file]: https://starship.rs/config/
